package provider

// Changes the API makes through an action instead of the update: a volume's resize, a database's
// rescale, a network attached to a server. A create-only field would replace the object, and lose what
// it holds, for a change the API makes in place. The binding names the actions; nothing here knows a
// product.

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type inPlaceChange struct {
	Field string
	// Set sets the field, or adds one item of a list; Unset, where the API has it, clears or removes.
	Set   *endpoint
	Unset *endpoint
	Items bool
	// Param is the body key that carries the value (as a list of one, with ItemList). ItemVar is the
	// path variable an item fills.
	Param    string
	ItemList bool
	ItemVar  string
	// SelfParam and SelfVar name this object where the actions belong to the item's own resource.
	SelfParam string
	SelfList  bool
	SelfVar   string
	// Wait and UnsetWait are what the object settles on after each, as JSON (wait.go).
	Wait      string
	UnsetWait string
}

// itemsLimit is how long a list may take to read back as it was set.
var itemsLimit = 5 * time.Minute

// UpdateInPlace makes the update and every in-place change the configuration asks for, and returns
// the object as it is afterwards. done holds, on failure too, the items of each list the API never
// returns, as far as they were changed.
func (c *Client) UpdateInPlace(ctx context.Context, b *apiBinding, vars map[string]string, self string, update, updated, wanted, prior map[string]interface{}, imported bool, diags *diag.Diagnostics) (map[string]interface{}, map[string][]string, error) {
	done := map[string][]string{}
	var refused error
	if bodyChanges(update, updated) {
		if _, err := c.Update(ctx, b, vars, update); err != nil {
			// The new value may only fit after an in-place change (a cache size in a larger type), so
			// the update is sent again after it.
			var apiErr *APIError
			if !errors.As(err, &apiErr) || (apiErr.Status != 409 && apiErr.Status != 422) || !c.changesInPlace(b, wanted, prior) {
				return nil, done, err
			}
			refused = err
		}
	}
	before, found, err := c.Get(ctx, b, vars)
	if err != nil {
		return nil, done, err
	}
	if !found || b.gone(before) {
		return nil, done, fmt.Errorf("%s no longer exists. Run plan again to create it.", b.Resource)
	}
	for _, change := range b.InPlace {
		want, ok := wanted[change.Field]
		// Only what the plan changes: a value changed outside Terraform since the last refresh isn't in it.
		if !ok || !change.changes(want, prior[change.Field]) {
			continue
		}
		// An import leaves an input the API never returns empty in state: the object has it already.
		if _, returned := before[change.Field]; imported && !returned && prior[change.Field] == nil {
			continue
		}
		if change.Items {
			err = c.changeItems(ctx, b, change, vars, self, before, want, prior[change.Field], done, diags)
		} else {
			err = c.changeValue(ctx, b, change, vars, before, want, diags)
		}
		if err != nil {
			return nil, done, err
		}
	}
	if refused != nil {
		if _, err := c.Update(ctx, b, vars, update); err != nil {
			return nil, done, err
		}
	}
	out, found, err := c.Get(ctx, b, vars)
	if err == nil && !found {
		err = fmt.Errorf("%s was not found after the change.", b.Resource)
	}
	return out, done, err
}

// bodyChanges is true when the request sets anything the object doesn't hold already.
func bodyChanges(request, current map[string]interface{}) bool {
	for key, value := range request {
		held, ok := current[key]
		if !ok || !sameRequestValue(value, held) {
			return true
		}
	}
	return false
}

func sameRequestValue(a, b interface{}) bool {
	left, errA := json.Marshal(a)
	right, errB := json.Marshal(b)
	return errA == nil && errB == nil && string(left) == string(right)
}

// changes says whether the plan changes the field. A list's order is never a change.
func (change inPlaceChange) changes(want, prior interface{}) bool {
	if change.Items {
		return !sameItems(itemStrings(want), itemStrings(prior))
	}
	return !sameRequestValue(want, prior)
}

func (c *Client) changesInPlace(b *apiBinding, wanted, prior map[string]interface{}) bool {
	for _, change := range b.InPlace {
		if want, ok := wanted[change.Field]; ok && change.changes(want, prior[change.Field]) {
			return true
		}
	}
	return false
}

func (c *Client) changeValue(ctx context.Context, b *apiBinding, change inPlaceChange, vars map[string]string, before map[string]interface{}, want interface{}, diags *diag.Diagnostics) error {
	held := before[change.Field]
	if waitEqual(held, want) {
		return nil
	}
	if change.Unset != nil && held != nil {
		if err := c.unsetValue(ctx, b, change, vars, before); err != nil {
			return err
		}
	}
	row, err := c.setValue(ctx, b, change, vars, before, want)
	if err != nil {
		// Set back only what ended without the new value: a change still being made isn't undone.
		var unsettled *waitError
		ended := row == nil || (errors.As(err, &unsettled) && unsettled.Object[change.Field] == nil)
		if change.Unset != nil && held != nil && ended {
			if _, undo := c.setValue(ctx, b, change, vars, before, held); undo != nil {
				diags.AddError(fmt.Sprintf("%s: %s is no longer %v", b.Resource, change.Field, held), fmt.Sprintf("It was cleared to make the change, and couldn't be set back: %s", undo))
			}
		}
		return err
	}
	// The API can end a change as done with a remark (a replica still catching up).
	if messages := waitMessages(row, c.waitMessageKeys(change.Wait)); messages != "" && messages != waitMessages(before, c.waitMessageKeys(change.Wait)) {
		diags.AddWarning(fmt.Sprintf("%s: changed %s", b.Resource, change.Field), "Better Stack reports:"+messages)
	}
	return nil
}

func (c *Client) waitMessageKeys(raw string) []string {
	var policy waitPolicy
	if raw == "" || json.Unmarshal([]byte(raw), &policy) != nil {
		return nil
	}
	return policy.Messages
}

func (c *Client) setValue(ctx context.Context, b *apiBinding, change inPlaceChange, vars map[string]string, before map[string]interface{}, want interface{}) (map[string]interface{}, error) {
	body := map[string]interface{}{change.Param: want}
	out, _, err := c.call(ctx, b, change.Set, vars, body)
	if err != nil {
		return nil, err
	}
	row, err := c.settleOn(ctx, b, change.Wait, vars, body, out, before)
	var unsettled *waitError
	if errors.As(err, &unsettled) && unsettled.Object != nil && !waitEqual(unsettled.Object[change.Field], want) {
		return row, &waitError{Object: unsettled.Object, Message: fmt.Sprintf("the API accepted the change of %s to %v, but it is %v now. %s", change.Field, want, unsettled.Object[change.Field], unsettled.Message)}
	}
	return row, err
}

func (c *Client) unsetValue(ctx context.Context, b *apiBinding, change inPlaceChange, vars map[string]string, before map[string]interface{}) error {
	out, _, err := c.call(ctx, b, change.Unset, vars, nil)
	if err != nil {
		return err
	}
	_, err = c.settleOn(ctx, b, change.UnsetWait, vars, nil, out, before)
	return err
}

// Release clears what the API refuses to delete the object with: a volume is detached first.
func (c *Client) Release(ctx context.Context, b *apiBinding, vars map[string]string) error {
	row, found, err := c.Get(ctx, b, vars)
	if err != nil || !found || b.gone(row) {
		return err
	}
	for _, change := range b.InPlace {
		if change.Items || change.Unset == nil || row[change.Field] == nil {
			continue
		}
		if err := c.unsetValue(ctx, b, change, vars, row); err != nil {
			// What it was attached to can go at the same time, which detaches it.
			if now, found, readErr := c.Get(ctx, b, vars); conflicts(err) && readErr == nil && (!found || now[change.Field] == nil) {
				continue
			}
			return err
		}
	}
	return nil
}

// callDelete sends the delete, again while it is refused with a code the description says ends by itself.
func (c *Client) callDelete(ctx context.Context, b *apiBinding, vars map[string]string) (map[string]interface{}, int, error) {
	started := time.Now()
	for attempt := 0; ; attempt++ {
		out, status, err := c.call(ctx, b, b.Delete, vars, nil)
		var apiErr *APIError
		if !errors.As(err, &apiErr) || !waitPending(apiErr, b.DeleteRetryCodes) || time.Since(started) > itemsLimit {
			return out, status, err
		}
		delay := waitDelays[len(waitDelays)-1]
		if attempt < len(waitDelays) {
			delay = waitDelays[attempt]
		}
		select {
		case <-ctx.Done():
			return out, status, err
		case <-time.After(delay):
		}
	}
}

// changeItems makes the list's planned difference. It adds before it removes, so a refused item leaves
// the old ones as they are. Where the API says there can be no more of them (it can allow one
// firewall), it removes first, and adds the old ones back if the new one is still refused.
func (c *Client) changeItems(ctx context.Context, b *apiBinding, change inPlaceChange, vars map[string]string, self string, before map[string]interface{}, want, prior interface{}, done map[string][]string, diags *diag.Diagnostics) error {
	held, returned := before[change.Field]
	if !returned {
		held = prior
	}
	current := itemStrings(held)
	record := func() {
		if !returned {
			done[change.Field] = append([]string{}, current...)
		}
	}
	record()
	var adding, removing []string
	for _, item := range itemStrings(want) {
		if !containsString(itemStrings(prior), item) && !containsString(current, item) {
			adding = append(adding, item)
		}
	}
	for _, item := range itemStrings(prior) {
		if !containsString(itemStrings(want), item) && containsString(current, item) {
			removing = append(removing, item)
		}
	}
	add := func(item string) error {
		err := c.callItem(ctx, b, change, change.Set, vars, self, item)
		if conflicts(err) && returned {
			// A repeated request can find its own work done.
			if row, found, readErr := c.Get(ctx, b, vars); readErr == nil && found && containsString(itemStrings(row[change.Field]), item) {
				err = nil
			}
		}
		if err == nil {
			current = append(current, item)
			record()
		}
		return err
	}
	var removed []string
	remove := func() error {
		for _, item := range removing {
			err := c.callItem(ctx, b, change, change.Unset, vars, self, item)
			var apiErr *APIError
			if errors.As(err, &apiErr) && apiErr.Status == 404 {
				err = nil // Not there any more, which is what was asked for.
			}
			if err != nil {
				return err
			}
			removed = append(removed, item)
			current = withoutString(current, item)
			record()
		}
		removing = nil
		return nil
	}
	restore := func() {
		for _, item := range removed {
			if err := add(item); err != nil {
				diags.AddError(fmt.Sprintf("%s: %s lost an item", b.Resource, change.Field), fmt.Sprintf("%s was removed to make the change, and couldn't be added back: %s", item, err))
			}
		}
	}
	for _, item := range adding {
		err := add(item)
		if limitReached(err) && len(removing) > 0 {
			if err = remove(); err == nil {
				err = add(item)
			}
			if err != nil {
				restore()
			}
		}
		if err != nil {
			return err
		}
	}
	if err := remove(); err != nil {
		return err
	}
	if !returned {
		return nil
	}
	return c.itemsReadBack(ctx, b, change.Field, vars, itemStrings(want), itemStrings(prior))
}

// limitReached is the one refusal that removing an item ends. Any other leaves the old items in place.
func limitReached(err error) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr) && apiErr.Status == 409 && strings.Contains(apiErr.Body, `_limit_reached"`)
}

func conflicts(err error) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr) && apiErr.Status == 409
}

func (c *Client) callItem(ctx context.Context, b *apiBinding, change inPlaceChange, ep *endpoint, vars map[string]string, self, item string) error {
	call := map[string]string{}
	var body map[string]interface{}
	if change.SelfParam == "" {
		for k, v := range vars {
			call[k] = v
		}
		if ep.Method != "DELETE" {
			body = map[string]interface{}{change.Param: item}
			if change.ItemList {
				body[change.Param] = []string{item}
			}
		}
	} else if ep.Method != "DELETE" {
		body = map[string]interface{}{change.SelfParam: self}
		if change.SelfList {
			body[change.SelfParam] = []string{self}
		}
	}
	if change.SelfVar != "" {
		call[change.SelfVar] = self
	}
	if change.ItemVar != "" {
		call[change.ItemVar] = item
	}
	_, _, err := c.call(ctx, b, ep, call, body)
	return err
}

func (c *Client) itemsReadBack(ctx context.Context, b *apiBinding, field string, vars map[string]string, wanted, prior []string) error {
	started := time.Now()
	for attempt := 0; ; attempt++ {
		row, found, err := c.Get(ctx, b, vars)
		if err != nil {
			return err
		}
		if found && holdsChange(itemStrings(row[field]), wanted, prior) {
			return nil
		}
		if !found || time.Since(started) > itemsLimit {
			return &waitError{Object: row, Message: fmt.Sprintf("the API accepted the change of %s, but it doesn't show it yet. Run plan again to see where it is.", field)}
		}
		delay := waitDelays[len(waitDelays)-1]
		if attempt < len(waitDelays) {
			delay = waitDelays[attempt]
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
	}
}

// holdsChange is true when the list holds what was added and not what was removed.
func holdsChange(held, wanted, prior []string) bool {
	for _, item := range wanted {
		if !containsString(held, item) && !containsString(prior, item) {
			return false
		}
	}
	for _, item := range prior {
		if containsString(held, item) && !containsString(wanted, item) {
			return false
		}
	}
	return true
}

func itemStrings(v interface{}) []string {
	out := []string{}
	switch items := v.(type) {
	case []interface{}:
		for _, item := range items {
			out = append(out, stringifyScalar(item))
		}
	case []string:
		out = append(out, items...)
	}
	return out
}

func sameItems(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for _, item := range a {
		if !containsString(b, item) {
			return false
		}
	}
	return true
}

func withoutString(list []string, s string) []string {
	out := []string{}
	for _, v := range list {
		if v != s {
			out = append(out, v)
		}
	}
	return out
}

// refuseSeedChange refuses a change of the objects a create starts with: each is an object of its own
// afterwards, which its own resource manages. A configuration of an imported object may name them.
func refuseSeedChange(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse, name, manager string) {
	if req.Plan.Raw.IsNull() || req.State.Raw.IsNull() || keepImported(ctx, req.Private, true) {
		return
	}
	var config, state map[string]tftypes.Value
	if req.Config.Raw.As(&config) != nil || req.State.Raw.As(&state) != nil || holdsConfigured(config[name], state[name]) {
		return
	}
	resp.Diagnostics.AddError(fmt.Sprintf("%s can't be changed here", name), fmt.Sprintf("%s is what the object was created with. Add, change or remove one with %s, and set lifecycle { ignore_changes = [%s] } here if a value in it changed.", name, manager, name))
}

// holdsConfigured is true when every value the configuration sets is the one held.
func holdsConfigured(config, held tftypes.Value) bool {
	if !config.IsKnown() || config.IsNull() {
		return true
	}
	if !held.IsKnown() || held.IsNull() {
		return false
	}
	var fields, heldFields map[string]tftypes.Value
	if config.As(&fields) == nil && held.As(&heldFields) == nil {
		for name, value := range fields {
			if !holdsConfigured(value, heldFields[name]) {
				return false
			}
		}
		return true
	}
	var items, heldItems []tftypes.Value
	if config.As(&items) == nil && held.As(&heldItems) == nil {
		if len(items) != len(heldItems) {
			return false
		}
		for i := range items {
			if !holdsConfigured(items[i], heldItems[i]) {
				return false
			}
		}
		return true
	}
	return config.Equal(held)
}
