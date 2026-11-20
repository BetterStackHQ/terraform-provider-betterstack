#!/usr/bin/env python3
"""Offline, narrowly reviewed Terraform raw-state field handoff (Python 3.9+).

Read an ORIGINAL `terraform state pull` backup and the CURRENT post-move/import pull.
Prepare current state using reviewed refresh-only moves, read-only CLI imports and
state-only forget operations, never an ordinary apply that would reassert credentials.
This program never invokes Terraform, contacts an API, modifies either input, removes
resources or writes a backend. It creates one exclusive mode-0600 candidate file.

Keep concurrent Terraform runs disabled. Review the candidate diff privately, re-pull
and verify the backend still has the expected lineage/serial, then install the candidate
with Terraform's supported state-push/locking procedure (never force). Require a normal
no-op plan before any ordinary apply. A saved current file cannot prove backend freshness.

Only supported transfers: legacy AWS/GCP source facets, one inline collector database,
or one outgoing webhook's existing basic-auth password, all into already-moved/imported
targets.
The selected obsolete legacy address must already be absent from current state.
No general mapping language, arbitrary field selection, remote reassertion or raw secrets
in command arguments is supported. Repeat against the preceding candidate for each child.
"""
import argparse
import copy
from decimal import Decimal
import json
import os
from pathlib import Path
import re
import sys
import tempfile


class TransferError(Exception):
    pass


IDENTIFIER = r"[A-Za-z_][A-Za-z0-9_-]*"
INDEX = r'\[(?:0|[1-9][0-9]*|"(?:[^"\\\x00-\x1f]|\\["\\/bfnrt]|\\u[0-9a-fA-F]{4})*")\]'
MODULE = rf"(?:module\.{IDENTIFIER}(?:{INDEX})?\.)*"
PROVIDER = re.compile(rf'{MODULE}provider\["([^"\\]+)"\](?:\.{IDENTIFIER})?\Z')
MODULE_ADDRESS = re.compile(rf'(?:module\.{IDENTIFIER}(?:{INDEX})?)(?:\.module\.{IDENTIFIER}(?:{INDEX})?)*\Z')
AWS = {"aws_account_id": "aws_account_id", "aws_role_arn": "aws_role_arn", "aws_external_id": "aws_external_id", "auto_subscribe_log_groups": "aws_auto_subscribe_log_groups"}
GCP = {"gcp_account_id": "gcp_account_id", "gcp_project_id": "gcp_project_id", "gcp_project_number": "gcp_project_number"}
DATABASE = {"host": "host", "port": "port", "username": "username", "password": "password", "ssl_mode": "ssl_mode", "tls": "tls"}
# source type -> reviewed provider, target type, equivalent old managed owner type.
HANDOFFS = {
    "logtail_source_aws_account": ("betterstackhq/logtail", "betterstack_source", "logtail_source"),
    "logtail_source_gcp_project": ("betterstackhq/logtail", "betterstack_source", "logtail_source"),
    "logtail_collector": ("betterstackhq/logtail", "betterstack_collector_target", "logtail_collector_target"),
    "betteruptime_outgoing_webhook": ("betterstackhq/better-uptime", "betterstack_outgoing_webhook_integration", "betteruptime_outgoing_webhook"),
}


def decode(text):
    def object_pairs(pairs):
        result = {}
        for key, value in pairs:
            if key in result:
                raise TransferError("Duplicate JSON object keys are not supported.")
            result[key] = value
        return result

    def invalid_number(_):
        raise TransferError("State contains an invalid JSON number.")

    try:
        return json.loads(text, object_pairs_hook=object_pairs, parse_float=Decimal, parse_constant=invalid_number)
    except (ValueError, RecursionError) as error:
        raise TransferError("Input is not a single valid raw-state JSON document.") from error


def encode(value):
    # Preserve arbitrary integer IDs and unrelated JSON decimals without float64 rounding.
    if isinstance(value, Decimal):
        if not value.is_finite():
            raise TransferError("State contains an invalid JSON number.")
        return str(value)
    if isinstance(value, dict):
        return "{" + ",".join(json.dumps(key, ensure_ascii=False) + ":" + encode(child) for key, child in value.items()) + "}"
    if isinstance(value, list):
        return "[" + ",".join(encode(child) for child in value) + "]"
    return json.dumps(value, ensure_ascii=False, allow_nan=False)


def validate_state(document):
    if not isinstance(document, dict) or type(document.get("version")) is not int or document["version"] != 4:
        raise TransferError("Only Terraform raw-state version 4 is supported; do not use terraform show -json.")
    if type(document.get("serial")) is not int or document["serial"] < 0:
        raise TransferError("State serial is missing or invalid.")
    if not isinstance(document.get("lineage"), str) or not document["lineage"]:
        raise TransferError("State lineage is missing.")
    if not isinstance(document.get("resources"), list):
        raise TransferError("State resources are missing or invalid.")


def instances(document):
    for resource in document["resources"]:
        if not isinstance(resource, dict) or resource.get("mode") not in ("managed", "data"):
            raise TransferError("State contains an invalid resource record.")
        kind, name, module = resource.get("type"), resource.get("name"), resource.get("module", "")
        if not isinstance(kind, str) or not re.fullmatch(IDENTIFIER, kind) or not isinstance(name, str) or not re.fullmatch(IDENTIFIER, name):
            raise TransferError("State contains an invalid resource address.")
        if not isinstance(module, str) or (module and not MODULE_ADDRESS.fullmatch(module)):
            raise TransferError("State contains an unsupported module address.")
        prefix = (module + "." if module else "") + ("data." if resource["mode"] == "data" else "") + kind + "." + name
        rows = resource.get("instances")
        if not isinstance(rows, list):
            raise TransferError("Resource instances are missing or invalid.")
        for instance in rows:
            if not isinstance(instance, dict):
                raise TransferError("State contains an invalid instance.")
            address = prefix
            key = instance.get("index_key")
            if key is not None:
                if not (type(key) is str or (type(key) is int and key >= 0)):
                    raise TransferError("State contains an unsupported instance key.")
                address += "[" + json.dumps(key, ensure_ascii=False) + "]"
            yield address, resource, instance


def select(document, address):
    matches = [(resource, instance) for candidate, resource, instance in instances(document) if candidate == address]
    if len(matches) != 1:
        raise TransferError("An exact selected address is absent or ambiguous; include module and instance keys.")
    return matches[0]


def provider_identity(resource):
    config = resource.get("provider")
    match = PROVIDER.fullmatch(config) if isinstance(config, str) else None
    address = match.group(1).split("/") if match else []
    if len(address) != 3 or not all(part and part.isascii() for part in address):
        return None
    if any(any(char.isspace() or char in "?#@\\" for char in part) for part in address):
        return None
    return "/".join(address[1:]).lower()


def validate_instance(resource, instance, provider, kind):
    if resource.get("mode") != "managed" or resource.get("type") != kind:
        raise TransferError("The selected resource type does not support this handoff.")
    if provider_identity(resource) != provider:
        raise TransferError("The selected instance has the wrong provider identity.")
    if type(instance.get("schema_version")) is not int or instance["schema_version"] != 0:
        raise TransferError("Only reviewed instance schema version 0 is supported.")
    if instance.get("deposed") or instance.get("status") not in (None, ""):
        raise TransferError("Tainted, deposed or non-ready instances cannot be transferred.")
    attributes = instance.get("attributes")
    if not isinstance(attributes, dict) or "attributes_flat" in instance:
        raise TransferError("The selected instance does not contain supported structured attributes.")
    return attributes


def identity(value):
    if type(value) is int and value > 0:
        return str(value)
    if type(value) is str and re.fullmatch(r"[1-9][0-9]*", value):
        return value
    raise TransferError("A source, target or parent identity is missing or invalid.")


def ownership_identity(kind, attributes):
    remote_id = attributes.get("id")
    if kind not in ("betterstack_collector_target", "logtail_collector_target"):
        return identity(remote_id)
    parent_id = attributes.get("collector_id")
    if kind == "logtail_collector_target" and isinstance(remote_id, str) and ":" in remote_id:
        composite_parent, remote_id = remote_id.split(":", 1)
        composite_parent = identity(composite_parent)
        if parent_id is not None and identity(parent_id) != composite_parent:
            raise TransferError("A legacy target has conflicting composite and attribute parent identities.")
        parent_id = composite_parent
    return identity(parent_id), identity(remote_id)


def text(attributes, key, required=False):
    value = attributes.get(key)
    if value is None and not required:
        return None
    if type(value) is not str or (required and not value):
        raise TransferError("A required credential or configuration field is missing or has the wrong type.")
    return value


def facet_values(attributes, kind):
    mapping = AWS if kind == "logtail_source_aws_account" else GCP
    if set(attributes) - set(mapping) - {"id", "source_id"}:
        raise TransferError("The legacy facet contains an unreviewed attribute.")
    if identity(attributes.get("id")) != identity(attributes.get("source_id")):
        raise TransferError("Legacy facet identity does not match its source parent.")
    values = {}
    for old, new in mapping.items():
        value = attributes.get(old)
        if old == "auto_subscribe_log_groups":
            if value is not None and type(value) is not bool:
                raise TransferError("The legacy subscription default must be a boolean.")
        else:
            value = text(attributes, old)
            if value == "":
                value = None  # Published SDK optional credential fields use empty strings for absence.
        if value is not None:
            values[new] = value
    if kind == "logtail_source_aws_account":
        pair = (values.get("aws_role_arn"), values.get("aws_external_id"))
        if bool(pair[0]) != bool(pair[1]) or not (values.get("aws_account_id") or all(pair)):
            raise TransferError("AWS account identity or the complete role/external-ID credentials are missing.")
    else:
        pair = (values.get("gcp_project_id"), values.get("gcp_project_number"))
        if bool(pair[0]) != bool(pair[1]) or not (values.get("gcp_account_id") or all(pair)):
            raise TransferError("GCP account identity or the complete project credentials are missing.")
        if pair[1]:
            identity(pair[1])
    return values


def database_values(backup, current, target, child_id):
    parent_id = identity(backup.get("id"))
    if identity(target.get("collector_id")) != parent_id:
        raise TransferError("Imported target belongs to a different collector.")
    child_id = identity(child_id)
    if identity(target.get("id")) != child_id:
        raise TransferError("Imported target identity does not match the selected child.")
    rows = backup.get("databases")
    if not isinstance(rows, list):
        raise TransferError("The original collector has no supported inline database state.")
    seen, selected = set(), None
    for row in rows:
        if not isinstance(row, dict):
            raise TransferError("The original collector has an invalid child record.")
        row_id = identity(row.get("id"))
        if row_id in seen:
            raise TransferError("The original collector contains ambiguous child identities.")
        seen.add(row_id)
        if row_id == child_id:
            selected = row
    if selected is None or set(selected) - set(DATABASE) - {"id", "service_type"}:
        raise TransferError("The selected child is absent or contains unreviewed fields.")
    kind = selected.get("service_type")
    if kind not in ("postgres", "mysql", "redis", "mongodb", "memcached") or target.get("kind") != kind:
        raise TransferError("Imported target kind does not match the legacy database.")
    parent_types = {"betterstack_collector": "betterstackhq/betterstack", "logtail_collector": "betterstackhq/logtail"}
    parents = [(resource, instance) for _, resource, instance in instances(current)
               if resource.get("mode") == "managed" and resource.get("type") in parent_types
               and provider_identity(resource) == parent_types[resource["type"]]
               and isinstance(instance.get("attributes"), dict) and identity(instance["attributes"].get("id")) == parent_id]
    if len(parents) != 1:
        raise TransferError("The moved collector parent is missing or ambiguous in current state.")
    parent = validate_instance(*parents[0], "betterstackhq/betterstack", "betterstack_collector")
    if text(parent, "secret", required=True) != text(backup, "secret", required=True):
        raise TransferError("The moved collector did not preserve its original enrollment secret.")
    values = {}
    for old, new in DATABASE.items():
        value = selected.get(old)
        if old == "port":
            if type(value) is not int or not 1 <= value <= 65535:
                raise TransferError("The legacy database port is missing or invalid.")
        else:
            value = text(selected, old, required=(old == "host"))
            if old == "password" and value is None:
                raise TransferError("The legacy database password is missing; recover original state/configuration before transfer.")
            if old not in ("host", "password") and value == "":
                value = None
        if value is not None:
            values[new] = value
    return values


def prepare(backup, current, source_address, target_address, expected_serial, child_id):
    validate_state(backup)
    validate_state(current)
    if backup["lineage"] != current["lineage"]:
        raise TransferError("Original and current state lineages differ.")
    if type(expected_serial) is not int or current["serial"] != expected_serial or backup["serial"] > current["serial"]:
        raise TransferError("The expected current serial is stale or the backup is newer than current state.")
    source_resource, source_instance = select(backup, source_address)
    kind = source_resource["type"]
    if kind not in HANDOFFS:
        raise TransferError("The selected source type has no reviewed offline handoff.")
    source_provider, target_kind, legacy_owner_kind = HANDOFFS[kind]
    source = validate_instance(source_resource, source_instance, source_provider, kind)
    if any(address == source_address for address, _, _ in instances(current)):
        raise TransferError("The legacy address remains in current state; complete the reviewed state-only move/forget first.")
    candidate = copy.deepcopy(current)
    target_resource, target_instance = select(candidate, target_address)
    target = validate_instance(target_resource, target_instance, "betterstackhq/betterstack", target_kind)
    equivalent_types = {target_kind: "betterstackhq/betterstack", legacy_owner_kind: source_provider}
    target_identity = ownership_identity(target_kind, target)
    owners = [instance for _, resource, instance in instances(candidate)
              if resource.get("mode") == "managed" and resource.get("type") in equivalent_types
              and provider_identity(resource) == equivalent_types[resource["type"]]
              and isinstance(instance.get("attributes"), dict)
              and ownership_identity(resource["type"], instance["attributes"]) == target_identity]
    if len(owners) != 1:
        raise TransferError("Current state contains ambiguous owners of the target remote identity.")
    destination, sensitive_prefix = target, []
    if kind == "logtail_collector":
        if child_id is None:
            raise TransferError("Inline database handoff requires an exact --child-id.")
        values = database_values(source, candidate, target, child_id)
        sensitive = {"password"}
    elif kind == "betteruptime_outgoing_webhook":
        if child_id is not None:
            raise TransferError("This handoff does not accept --child-id.")
        if identity(target.get("id")) != identity(source.get("id")):
            raise TransferError("Imported target identity differs from the original resource.")
        templates = source.get("custom_webhook_template_attributes")
        if not isinstance(templates, list) or len(templates) != 1 or not isinstance(templates[0], dict):
            raise TransferError("The original webhook must contain exactly one custom template.")
        password = templates[0].get("auth_password")
        if type(password) is not str:
            raise TransferError("The original webhook password is missing or invalid; recover the original state before transfer.")
        destination = target.get("custom_webhook_template")
        if not isinstance(destination, dict):
            raise TransferError("The imported webhook has no existing custom template.")
        values, sensitive = {"auth_password": password}, {"auth_password"}
        sensitive_prefix = [{"type": "get_attr", "value": "custom_webhook_template"}]
    else:
        if child_id is not None:
            raise TransferError("Source-facet handoff does not accept --child-id.")
        if identity(target.get("id")) != identity(source.get("source_id")):
            raise TransferError("Target source identity does not match the original facet parent.")
        if target.get("platform") != ("aws" if kind == "logtail_source_aws_account" else "gcp"):
            raise TransferError("Target source platform does not match the legacy facet.")
        values = facet_values(source, kind)
        sensitive = {"aws_role_arn", "aws_external_id"}
    for field, value in values.items():
        if field not in destination:
            raise TransferError("The target state lacks a required handoff field; verify the target provider version.")
        prior = destination[field]
        if prior is not None and (prior != "" or source_provider == "betterstackhq/better-uptime") and not (type(prior) is type(value) and prior == value):
            raise TransferError("A target field conflicts with the original state; no values were replaced.")
        destination[field] = value
    paths = target_instance.setdefault("sensitive_attributes", [])
    if not isinstance(paths, list):
        raise TransferError("Target sensitivity metadata is invalid.")
    for field in sorted(sensitive.intersection(values)):
        path = sensitive_prefix + [{"type": "get_attr", "value": field}]
        if path not in paths:
            paths.append(path)
    candidate["serial"] += 1
    return candidate


def publish(candidate, output):
    destination = Path(output)
    fd, temporary = tempfile.mkstemp(prefix=".transfer-state-", suffix=".tmp", dir=destination.parent)
    try:
        with os.fdopen(fd, "w", encoding="utf-8") as stream:
            os.fchmod(stream.fileno(), 0o600)
            payload = encode(candidate) + "\n"
            if stream.write(payload) != len(payload):
                raise OSError("Short write")
            stream.flush()
            os.fsync(stream.fileno())
        os.link(temporary, destination)  # Atomic exclusive publication: never replace an existing file/symlink.
    finally:
        os.unlink(temporary)


def main(argv=None):
    parser = argparse.ArgumentParser(description="Prepare an exclusive private state-only credential or reviewed policy handoff candidate; never invokes Terraform or any API.")
    for flag in ("backup", "current", "source", "target", "output"):
        parser.add_argument("--" + flag, required=True)
    parser.add_argument("--expected-serial", required=True, type=int)
    parser.add_argument("--child-id")
    args = parser.parse_args(argv)
    try:
        backup = decode(Path(args.backup).read_text(encoding="utf-8"))
        current = decode(Path(args.current).read_text(encoding="utf-8"))
        candidate = prepare(backup, current, args.source, args.target, args.expected_serial, args.child_id)
        publish(candidate, args.output)
    except TransferError as error:
        print("Handoff refused: " + str(error), file=sys.stderr)
        return 1
    except (OSError, UnicodeError, ValueError, RecursionError):
        print("Handoff refused: input could not be read or an exclusive candidate could not be published. No existing file was replaced.", file=sys.stderr)
        return 1
    print("Private candidate created. Inputs and backend are unchanged. Review privately, verify current backend lineage/serial under the migration lock, then use supported state push and require a no-op plan.")
    return 0


if __name__ == "__main__":
    sys.exit(main())
