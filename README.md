# koron/s3tdl - S3 Tables Data File Downloader

[![PkgGoDev](https://pkg.go.dev/badge/github.com/koron/s3tdl)](https://pkg.go.dev/github.com/koron/s3tdl)
[![Actions/Go](https://github.com/koron/s3tdl/actions/workflows/go.yml/badge.svg)](https://github.com/koron/s3tdl/actions/workflows/go.yml)
[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/koron/s3tdl)

[Engilish](README.md) | [日本語](README-ja.md)

A tool to download Parquet files that compose tables within a specified AWS S3 Tables bucket, and to inspect the Iceberg catalog of that bucket.

## Installation

If you have Go installed, you can install or update the tool with the following command:

```console
$ go install github.com/koron/s3tdl@latest
```

Alternatively, you can download compiled binaries directly from the [latest releases](https://github.com/koron/s3tdl/releases/latest).

## Usage

This tool provides the following sub-commands:

```console
$ s3tdl download {S3 Tables bucket ARN} [...]
$ s3tdl inspect {S3 Tables bucket ARN}
```

### `download` sub-command

Run the `download` sub-command specifying one or more S3 Tables bucket ARNs to download the Parquet data files of the tables:

```console
$ s3tdl download {S3 Tables bucket ARN}
```

By default, all tables across all namespaces in the S3 Tables bucket are included for download.
If you wish to limit the scope, please specify a regular expression filter using the `-namespace` or `-table` options.

#### Options

| Name | Description |
| --- | --- |
| `-dryrun` | Check execution behavior without downloading files |
| `-namespace` | Filter target Namespaces using regular expressions |
| `-table` | Filter target table names using regular expressions |
| `-outdir` | Output directory for downloaded files (default: `.`) |
| `-verbose` | Display detailed logs |

### `inspect` sub-command

Run the `inspect` sub-command specifying one or more S3 Tables bucket ARNs to inspect the Iceberg catalog of the bucket:

```console
$ s3tdl inspect {S3 Tables bucket ARN}
```

It lists all namespaces, and for each table it shows the namespace properties, the table metadata, the current snapshot, the schema, the sort order, and the partition spec. It is read-only and does not download any data files.

This sub-command does not accept any options other than the bucket ARNs.

### Authentication Credentials

For accessing AWS, this tool uses the standard credential provider chain from AWS SDK for Go v2.
Credentials are automatically retrieved in the following order:

* Environment variables for static credentials: `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `AWS_SESSION_TOKEN`
* Web Identity Token: `AWS_WEB_IDENTITY_TOKEN_FILE`
* Shared configuration files: `~/.aws/credentials`, `~/.aws/config`, and `AWS_PROFILE`
* ECS Tasks: IAM role assigned to the task
* EC2: IAM role assigned to the EC2 instance

For details, refer to the [AWS SDK for Go v2 Developer Guide](https://docs.aws.amazon.com/sdk-for-go/v2/developer-guide/configure-gosdk.html).
