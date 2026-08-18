# AGENTS.md

s3tdl is a single-module Go CLI (not a library) that downloads Parquet data files from an AWS S3 Tables bucket via the Iceberg REST catalog.

## Commands

- Build: `go build ./...` or `make build`
- Check (CI-equivalent local): `make checkall` (runs `go vet` + `staticcheck`)
- Test: `make test` (currently no test files exist; this is a no-op)
- Race: `make race`

Go version: `go 1.25.5` in go.mod; CI uses `oldstable` (`.github/workflows/go.yml`).

## Structure

- `main.go` — wires subcommands via `github.com/koron-go/subcmd`
- `internal/download` — `download` subcommand: lists namespaces/tables, plans files, downloads Parquet
- `internal/inspect` — `inspect` subcommand: read-only catalog inspection
- `internal/common` — ARN parsing and Iceberg REST catalog construction (endpoint: `https://s3tables.<region>.amazonaws.com/iceberg`, SigV4)

ARNs are **positional arguments**, not flags: `s3tdl download <arn> [arn...]`. (The README's `-arn` flag usage is stale; the code is the source of truth.)

## Gotchas

- AWS access uses the standard AWS SDK for Go v2 credential chain; region is derived from the ARN. There is no `-arn` flag and no env-based override.
- `-dryrun` flag on `download` verifies listing without downloading — use it to verify behavior against a real bucket.
- Downloads land in `./<namespace>/<table>/` under `-outdir` (default `.`); `tmp/` is the scratch/output area (git-ignored).
- No CI lint step beyond what's in the workflow (go test + build on a matrix); releases are draft GitHub releases triggered by `vX.Y` tags.
