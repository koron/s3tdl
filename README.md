# koron/s3tdl - S3 Tables Data File Downloader

[![PkgGoDev](https://pkg.go.dev/badge/github.com/koron/s3tdl)](https://pkg.go.dev/github.com/koron/s3tdl)
[![Actions/Go](https://github.com/koron/s3tdl/actions/workflows/go.yml/badge.svg)](https://github.com/koron/s3tdl/actions/workflows/go.yml)
[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/koron/s3tdl)

[Engilish](README.md) | [日本語](README-ja.md)

A tool to download Parquet files that compose tables within a specified AWS S3 Tables bucket.

## Installation

If you have Go installed, you can install or update the tool with the following command:

```console
$ go install github.com/koron/s3tdl@latest
```

Alternatively, you can download compiled binaries directly from the [latest releases](https://github.com/koron/s3tdl/releases/latest).

## Usage

Run the tool specifying the S3 Tables bucket ARN with the `-arn` option to download the Parquet files:

```console
$ s3tdl -arn {S3 Tables bucket ARN}
```

By default, all tables across all namespaces in the S3 Tables bucket are included for download.
If you wish to limit the scope, please specify a regular expression filter using the `-namespace` or `-table` options.

### Options

| Name | Description |
| --- | --- |
| `-arn` | Target S3 Tables bucket ARN (required) |
| `-dryrun` | Check execution behavior without downloading files |
| `-namespace` | Filter target Namespaces using regular expressions |
| `-table` | Filter target table names using regular expressions |
| `-outdir` | Output directory for downloaded files (default: `.`) |
| `-verbose` | Display detailed logs |

### Authentication Credentials

For accessing AWS, this tool uses the standard credential provider chain from AWS SDK for Go v2.
Credentials are automatically retrieved in the following order:

* Environment variables for static credentials: `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `AWS_SESSION_TOKEN`
* Web Identity Token: `AWS_WEB_IDENTITY_TOKEN_FILE`
* Shared configuration files: `~/.aws/credentials`, `~/.aws/config`, and `AWS_PROFILE`
* ECS Tasks: IAM role assigned to the task
* EC2: IAM role assigned to the EC2 instance

For details, refer to the [AWS SDK for Go v2 Developer Guide](https://docs.aws.amazon.com/sdk-for-go/v2/developer-guide/configure-gosdk.html).
