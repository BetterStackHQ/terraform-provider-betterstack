#!/usr/bin/env ruby
# The export receipt is created by bin/export_terraform.rb, never by API generation.
require "digest"
require "fileutils"
require "json"
require "open3"
require "tmpdir"
require "rubygems/version"

module ProviderPackage
  GO_VERSION = "1.25.9"
  GORELEASER_VERSION = "2.8.2"
  NAME = "terraform-provider-betterstack"

  def self.capture(*command, env: {})
    out, _err, status = Open3.capture3(env, *command)
    raise "#{command.first} failed (exit #{status.exitstatus})" unless status.success?
    out
  end

  def self.run(*command, env: {})
    raise "#{command.first} failed" unless system(env, *command)
  end

  def self.verify
    revision = File.read("SOURCE_REVISION").strip
    raise "invalid source revision" unless revision.match?(/\A(?:[0-9a-f]{40}|[0-9a-f]{64})\z/)
    expected = checksum_entries("SOURCE_SHA256SUMS")
    files = Dir.glob("**/*", File::FNM_DOTMATCH).reject do |path|
      path == "." || path == ".git" || path.start_with?(".git/") || path == "dist" || path.start_with?("dist/") ||
        path == "SOURCE_SHA256SUMS" || (File.directory?(path) && !File.symlink?(path))
    end.sort
    raise "export file inventory differs from verified revision" unless expected.keys.sort == files
    expected.each do |path, checksum|
      raise "non-regular export file: #{path}" unless File.file?(path) && !File.symlink?(path)
      raise "source checksum mismatch: #{path}" unless Digest::SHA256.file(path).hexdigest == checksum
    end
    manifest = JSON.parse(File.read("terraform-registry-manifest.json"))
    raise "protocol-6 registry manifest required" unless manifest == { "version" => 1, "metadata" => { "protocol_versions" => ["6.0"] } }
    raise "Apache-2.0 license required" unless File.read("LICENSE").include?("Apache License\n                           Version 2.0, January 2004")
    raise "pending license cannot be exported" if File.exist?("LICENSE-PENDING.md")
    migration = JSON.parse(File.read("migration-manifest.json"))
    minimum = migration.fetch("minimum_provider_version")
    raise "invalid minimum provider version" unless minimum.match?(/\A(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\z/)
    identity = ->(entry) { [entry.fetch("source_provider"), entry.fetch("source_type"), entry.fetch("schema_versions").sort, entry.fetch("target_type")] }
    promised = migration.fetch("mappings").select { |entry| entry["kind"] == "resource" && entry["handling"] == "move" }.map(&identity).sort
    emitted = migration.fetch("emitted_movers").map(&identity).sort
    raise "migration metadata promises movers absent from provider" unless promised == emitted && emitted.uniq == emitted
    if (tag = ENV["GORELEASER_CURRENT_TAG"])
      raise "explicit vX.Y.Z release tag required" unless tag.match?(/\Av(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\z/)
      raise "release predates migration contract" if Gem::Version.new(tag.delete_prefix("v")) < Gem::Version.new(minimum)
    end
    puts "Verified provider export from #{revision}"
  end

  # Go stamps a binary built from a tree with untracked or changed files as modified (vX.Y.Z+dirty),
  # and such a release can't be rebuilt from its tag.
  def self.check_build(binary)
    info = capture("go", "version", "-m", binary).lines.map { |line| line.strip.split("\t") }
    modified = info.find { |kind, value| kind == "build" && value.to_s.start_with?("vcs.modified=") }&.last
    raise "#{binary} was built from a tree with uncommitted files" if modified == "vcs.modified=true"
    if (tag = ENV["GORELEASER_CURRENT_TAG"])
      raise "#{binary} has no clean git stamp" unless modified == "vcs.modified=false"
      version = info.find { |kind, path| kind == "mod" && path == "github.com/betterstackhq/terraform-provider-betterstack" }&.fetch(2, nil)
      raise "#{binary} is stamped #{version.inspect}, not #{tag}" unless version == tag
    end
    puts "Build stamp of #{File.basename(binary)} is clean"
  end

  def self.checksum_entries(path)
    entries = {}
    File.foreach(path) do |line|
      match = line.match(/\A([0-9a-f]{64})  ([a-zA-Z0-9_.\/-]+)\n?\z/)
      raise "invalid checksum entry" unless match
      name = match[2]
      raise "unsafe checksum path" if name.start_with?("/") || name.split("/").any? { |part| part == "." || part == ".." || part.empty? }
      raise "duplicate checksum path: #{name}" if entries.key?(name)
      entries[name] = match[1]
    end
    raise "empty checksum file" if entries.empty?
    entries
  end

  # release_key: sign with the key already imported for GPG_FINGERPRINT (the release workflow's
  # dry run) instead of a disposable one. Either way nothing is published.
  def self.snapshot(release_key: false)
    verify
    raise "dist already exists; use a fresh export for packaging" if File.exist?("dist") || File.symlink?("dist")
    env = { "GOTOOLCHAIN" => "go#{GO_VERSION}", "GITHUB_TOKEN" => nil, "GH_TOKEN" => nil,
            "GITLAB_TOKEN" => nil, "GITEA_TOKEN" => nil, "GPG_PRIVATE_KEY" => nil, "PASSPHRASE" => nil }
    raise "Go #{GO_VERSION} required" unless capture("go", "env", "GOVERSION", env: env).strip == "go#{GO_VERSION}"
    version = capture("goreleaser", "--version", env: env)
    raise "GoReleaser #{GORELEASER_VERSION} required" unless version.match?(/(?:GitVersion:|goreleaser version)\s+#{Regexp.escape(GORELEASER_VERSION)}(?:\s|\z)/)
    run("go", "test", "-mod=readonly", "./...", env: env)
    run("python3", "-B", "scripts/test_transfer_state.py", env: env)
    if release_key
      env["GPG_FINGERPRINT"] = ENV.fetch("GPG_FINGERPRINT", "")
      raise "GPG_FINGERPRINT of the imported release key required" unless env["GPG_FINGERPRINT"].match?(/\A[0-9A-F]{40}\z/)
      sign_snapshot(env)
      puts "Snapshot archives, protocol manifest, checksums and release-key signature (#{env["GPG_FINGERPRINT"]}) verified in dist/; nothing published."
      return
    end
    # GPG socket paths must fit macOS's short Unix-socket limit.
    Dir.mktmpdir("gpg-") do |home|
      File.chmod(0700, home)
      env["GNUPGHOME"] = home
      begin
        run("gpg", "--batch", "--pinentry-mode", "loopback", "--passphrase", "", "--quick-generate-key",
            "Better Stack local verification <package@example.invalid>", "rsa2048", "sign", "1d", env: env)
        env["GPG_FINGERPRINT"] = capture("gpg", "--batch", "--with-colons", "--list-secret-keys", env: env)
          .lines.find { |line| line.start_with?("fpr:") }&.split(":")&.fetch(9)
        raise "disposable signing key was not created" unless env["GPG_FINGERPRINT"]
        sign_snapshot(env)
      ensure
        system(env, "gpgconf", "--kill", "all", out: File::NULL, err: File::NULL)
      end
    end
    puts "Snapshot archives, protocol manifest, checksums and disposable-key signature verified in dist/; nothing published."
  end

  def self.sign_snapshot(env)
    run("goreleaser", "release", "--snapshot", "--skip=publish", env: env)
    checksums = Dir.glob("dist/#{NAME}_*_SHA256SUMS")
    raise "expected one provider checksum file" unless checksums.one?
    entries = checksum_entries(checksums.first)
    manifests = entries.keys.grep(/\A#{NAME}_.+_manifest\.json\z/)
    raise "manifest missing from checksums" unless manifests.one?
    # GoReleaser's release pipe uploads this extra file; snapshots skip that pipe.
    FileUtils.copy_file("terraform-registry-manifest.json", File.join("dist", manifests.first))
    raise "no platform archives" unless entries.keys.any? { |name| name.end_with?(".zip") }
    entries.each do |name, checksum|
      raise "unexpected provider artifact name" unless name.match?(/\A#{NAME}_[a-zA-Z0-9_.-]+\.(?:zip|json)\z/)
      artifact = File.join("dist", name)
      raise "artifact checksum mismatch: #{name}" unless File.file?(artifact) && Digest::SHA256.file(artifact).hexdigest == checksum
      next unless name.end_with?(".zip")
      binary = "#{NAME}_v0.0.0-snapshot#{name.include?('_windows_') ? '.exe' : ''}"
      members = capture("unzip", "-Z1", artifact, env: env).lines.map(&:strip).sort
      raise "unexpected archive contents: #{name}" unless members == ["LICENSE", binary].sort
      raise "archive license differs: #{name}" unless capture("unzip", "-p", artifact, "LICENSE", env: env) == File.read("LICENSE")
    end
    status = capture("gpg", "--batch", "--status-fd", "1", "--verify", "#{checksums.first}.sig", checksums.first, env: env)
    # VALIDSIG names the signing (sub)key first and the primary key last.
    signers = status.lines.find { |line| line.start_with?("[GNUPG:] VALIDSIG ") }&.split&.values_at(2, -1)
    raise "checksum signature is not by #{env.fetch("GPG_FINGERPRINT")}" unless signers&.include?(env.fetch("GPG_FINGERPRINT"))
    File.binwrite("dist/#{NAME}_SNAPSHOT_SIGNING_KEY.asc", capture("gpg", "--batch", "--armor", "--export", env.fetch("GPG_FINGERPRINT"), env: env))
    verify
  end
end

if $PROGRAM_NAME == __FILE__
  begin
    Dir.chdir(File.expand_path("..", __dir__)) do
      case ARGV
      when ["--verify"] then ProviderPackage.verify
      when ["--snapshot"] then ProviderPackage.snapshot
      when ["--signed-snapshot"] then ProviderPackage.snapshot(release_key: true)
      when ->(args) { args.length == 2 && args.first == "--check-build" } then ProviderPackage.check_build(ARGV[1])
      else raise "usage: ruby scripts/package.rb --verify|--snapshot|--signed-snapshot|--check-build BINARY (never publishes)"
      end
    end
  rescue StandardError => error
    warn "Provider packaging failed: #{error.message}"
    exit 1
  end
end
