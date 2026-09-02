# koron/s3tdl - S3 Tables Data File Downloader

[![PkgGoDev](https://pkg.go.dev/badge/github.com/koron/s3tdl)](https://pkg.go.dev/github.com/koron/s3tdl)
[![Actions/Go](https://github.com/koron/s3tdl/actions/workflows/go.yml/badge.svg)](https://github.com/koron/s3tdl/actions/workflows/go.yml)
[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/koron/s3tdl)

[Engilish](README.md) | [日本語](README-ja.md)

A tool to download Parquet files that compose tables in an Iceberg REST catalog, and to inspect the catalog.
It works with AWS S3 Tables and other Iceberg REST catalogs (e.g. Cloudflare R2) through a YAML config file.

## Installation

If you have Go installed, you can install or update the tool with the following command:

```console
$ go install github.com/koron/s3tdl@latest
```

Alternatively, you can download compiled binaries directly from the [latest releases](https://github.com/koron/s3tdl/releases/latest).

## Configuration

Catalog access is driven by a YAML config file (default: `tmp/config.yaml`).
The top-level key is `catalog`, which is a map of catalog entries keyed by name:

```yaml
catalog:
  default:
    uri: https://<account-id>.s3tables.<region>.amazonaws.com/iceberg
    warehouse: arn:aws:s3tables:us-east-1:<account-id>:bucket/<bucket-name>
    aws:
      profile: myprofile
      region: us-east-1
      s3-endpoint: https://<account>.r2.cloudflarestorage.com # optional
    rest:
      sigv4-enabled: true
      signing-name: s3tables
      signing-region: us-east-1
```

| Key | Description |
| --- | --- |
| `uri` | Iceberg REST catalog endpoint |
| `warehouse` | Warehouse location (e.g. an S3 Tables bucket ARN) |
| `token` | Static OAuth token for the REST catalog (optional) |
| `credential` | OAuth credential for the REST catalog (optional, alternative to `token`) |
| `scope` | OAuth scope for the REST catalog (optional) |
| `aws.profile` | AWS shared config profile to use (optional) |
| `aws.region` | AWS region to use (optional) |
| `aws.s3-endpoint` | Alternate S3 endpoint for downloading data files, e.g. Cloudflare R2 (optional) |
| `rest.sigv4-enabled` | Enable SigV4 signing for REST requests (AWS S3 Tables) |
| `rest.signing-name` | Service name for SigV4 signing (e.g. `s3tables`) |
| `rest.signing-region` | Region for SigV4 signing |

## Usage

This tool provides the following sub-commands. Both read the config described above via the `-config` and `-catalog` options:

```console
$ s3tdl download [-config PATH] [-catalog NAME] [-namespace RE] [-table RE] [-outdir DIR] [-dryrun] [-verbose]
$ s3tdl inspect  [-config PATH] [-catalog NAME] [-verbose] [-deleted]
```

| Name | Default | Description |
| --- | --- | --- |
| `-config` | `tmp/config.yaml` | Location of the config YAML |
| `-catalog` | `default` | Catalog name to use from the config |

### `download` sub-command

The `download` sub-command downloads the Parquet data files of the tables in the catalog:

```console
$ s3tdl download -namespace rx -table rx -outdir ./out
```

By default, all tables across all namespaces in the catalog are included for download.
If you wish to limit the scope, please specify a regular expression filter using the `-namespace` or `-table` options.

Files are written to `-outdir/<namespace>/<table>/<file basename>`.
Note that only the basename of each data file is used, so same-named files under different S3 prefixes overwrite each other.

#### Options

| Name | Description |
| --- | --- |
| `-config` | Location of the config YAML (default: `tmp/config.yaml`) |
| `-catalog` | Catalog name to use from the config (default: `default`) |
| `-dryrun` | Check execution behavior without downloading files |
| `-namespace` | Filter target namespaces using regular expressions |
| `-table` | Filter target table names using regular expressions |
| `-outdir` | Output directory for downloaded files (default: `.`) |
| `-verbose` | Display detailed logs |

Use `-dryrun` to list the files that would be downloaded without actually downloading them; this is the cheapest way to verify a config against a real catalog.

### `inspect` sub-command

The `inspect` sub-command inspects the Iceberg catalog:

```console
$ s3tdl inspect
```

It lists all namespaces, and for each table it shows the namespace properties, the table metadata, the current snapshot, the schema, the sort order, the partition spec, and the manifest list. It is read-only and does not download any data files.

#### Options

| Name | Description |
| --- | --- |
| `-config` | Location of the config YAML (default: `tmp/config.yaml`) |
| `-catalog` | Catalog name to use from the config (default: `default`) |
| `-verbose` | Display detailed stats of data files |
| `-deleted` | Include deleted data files in the manifest entries |

### Authentication Credentials

For accessing AWS, this tool uses the standard credential provider chain from AWS SDK for Go v2.
Credentials are automatically retrieved in the following order:

* Environment variables for static credentials: `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `AWS_SESSION_TOKEN`
* Web Identity Token: `AWS_WEB_IDENTITY_TOKEN_FILE`
* Shared configuration files: `~/.aws/credentials`, `~/.aws/config`, and `AWS_PROFILE`
* ECS Tasks: IAM role assigned to the task
* EC2: IAM role assigned to the EC2 instance

The `aws.profile` and `aws.region` keys in the config can override the profile and region for a specific catalog.
For details, refer to the [AWS SDK for Go v2 Developer Guide](https://docs.aws.amazon.com/sdk-for-go/v2/developer-guide/configure-gosdk.html).
