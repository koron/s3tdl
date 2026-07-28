# koron/s3tdl - S3 Tables Data File Downloader

[![PkgGoDev](https://pkg.go.dev/badge/github.com/koron/s3tdl)](https://pkg.go.dev/github.com/koron/s3tdl)
[![Actions/Go](https://github.com/koron/s3tdl/actions/workflows/go.yml/badge.svg)](https://github.com/koron/s3tdl/actions/workflows/go.yml)
[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/koron/s3tdl)

S3 Tables バケットを指定して、そのバケット内のテーブルを構成するParquetファイルをダウンロードするツールです。

## Getting Started

インストール・更新するには以下を実行してください。

```console
$ go intall github.com/koron/s3tdl@latest
```

または [最新のリリース](releases/latest) からコンパイル済みバイナリをダウンロードすることもできます。

## 使い方

基本的な使い方は以下の通りです。
オプション `-arn` で S3 Tables バケットのARNを指定すると、
Parquet ファイルをダウンロードします。

```console
$ s3tdl -arn {S3 Tables bucket ARN}
```

### オプション

| Name         | Description                                                    |
|--------------|----------------------------------------------------------------|
| `-arn`       | ARN for S3 Tables bucket                                       |
| `-dryrun`    | Dryrun, not actually download                                  |
| `-namespace` | Namespace filter regexp. Download only the matching namespaces |
| `-outdir`    | Output dir for downloaded data fiels (default ".")             |
| `-table`     | Table filter regexp. Download only the matching tables         |
| `-verbose`   | Show verbose messages                                          |

### 認証情報

AWSにアクセスする際の認証情報には以下のソースを使います。

-   環境変数:
    -   静的資格: `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `AWS_SESSION_TOKEN`
    -   Web Identity Token: `AWS_WEB_IDENTITY_TOKEN_FILE`
-   共有設定ファイル: `~/.aws/credentials`, `~/.aws/config`, and `AWS_PROFILE`
-   ECSタスク: タスクに割り当てられた IAM role
-   EC2: EC2に割り当てられたIAM role

参考: <https://docs.aws.amazon.com/sdk-for-go/v2/developer-guide/configure-gosdk.html>
