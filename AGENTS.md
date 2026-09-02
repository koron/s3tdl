# AGENTS.md

s3tdl is a single-module Go CLI (not a library) that downloads Iceberg table data files (Parquet) via an Iceberg REST catalog. It works with AWS S3 Tables and other REST catalogs (e.g. Cloudflare R2) through a YAML config file.

## Commands

- Build: `go build .` (the root main package). Avoid `go build ./...` while scratch `.go` files exist in `tmp/` (see Gotchas).
- Check (CI-equivalent): `make checkall` = `go vet` + `staticcheck` over `./...`. The `staticcheck` binary must be on PATH (`go install honnef.co/go/tools/cmd/staticcheck@latest`); `staticcheck.conf` enables `checks = ["all"]` (strict).
- Test: `make test` — no test files exist yet; CI only runs `go test ./...` plus a cross-compile build matrix (`.github/workflows/go.yml`).
- CGO is not required: `CGO_ENABLED=0 go build . ./internal/...` succeeds.

Go version: `go 1.25.9` (go.mod); CI uses `oldstable`.

## Usage (code is the source of truth; both READMEs are stale)

There are **no positional arguments**. All catalog access is driven by a YAML config:

```console
$ s3tdl download [-config tmp/config.yaml] [-catalog default] [-namespace rx] [-table rx] [-outdir DIR] [-dryrun] [-verbose]
$ s3tdl inspect  [-config tmp/config.yaml] [-catalog default]
```

- `-config` (default `tmp/config.yaml`) and `-catalog` (default `default`) are defined in `internal/common/config.go:130` and shared by both subcommands.
- Config format: top-level `catalog:` map; each entry has `uri` (REST endpoint), `warehouse`, optional `token`/`credential`/`scope`, `aws: {profile, region, s3-endpoint}`, `rest: {sigv4-enabled, signing-name, signing-region}`.
- `aws.s3-endpoint` redirects the S3 GetObject client (e.g. Cloudflare R2); SigV4 options target AWS S3 Tables.
- `-dryrun` lists/plans files without downloading — the cheapest way to verify a config against a real catalog.
- Files are written to `-outdir/<ns>/<table>/<file basename>` (default `.`). Only the basename is used, so same-named files under different S3 prefixes overwrite each other.

## Structure

- `main.go` — wires subcommands via `github.com/koron-go/subcmd`
- `internal/common` — catalog config loading (`config.go`); `common.go` holds `ParseARN`/`NewCatalog`, which are now **unused leftovers** from the pre-config ARN-positional design — do not treat ARN-positional usage as valid
- `internal/download` — `download` subcommand: lists namespaces/tables, `table.Scan().PlanFiles`, downloads via the S3 client
- `internal/inspect` — `inspect` subcommand: read-only catalog dump (schema, snapshot, sort order, partition spec)

AWS credentials use the standard AWS SDK for Go v2 default chain; profile/region can be overridden per catalog in the config's `aws:` section.

## Gotchas

- `tmp/` is inside the Go module and is used as a scratch area (git-ignored). It currently holds a real `config.yaml` with live tokens — never commit or echo it. Any `.go` file left in `tmp/` breaks `go build ./...`, `make test`, and `make checkall`; scope builds to `go build . ./internal/...` or clean `tmp/` first.
- `README.md` and `README-ja.md` still document the old positional-ARN usage; they are stale.
- Releases are draft GitHub releases auto-created by the workflow from `vX.Y` tags; a `.norelease` file in a main package's directory excludes it from release.
