# koron/s3tdl - S3 Tables Data File Downloader

[![PkgGoDev](https://pkg.go.dev/badge/github.com/koron/s3tdl)](https://pkg.go.dev/github.com/koron/s3tdl)
[![Actions/Go](https://github.com/koron/s3tdl/actions/workflows/go.yml/badge.svg)](https://github.com/koron/s3tdl/actions/workflows/go.yml)
[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/koron/s3tdl)

[Engilish](README.md) | [日本語](README-ja.md)

AWS S3 Tables バケットを指定して、そのバケット内のテーブルを構成するParquetファイルをダウンロードし、
バケットの Iceberg カタログを inspect できるツールです。

## インストール

Go がインストールされている環境では、以下のコマンドでインストール・更新できます。

```console
$ go install github.com/koron/s3tdl@latest
```

または [最新のリリース](https://github.com/koron/s3tdl/releases/latest) からコンパイル済みバイナリを直接ダウンロードすることもできます。

## 使い方

このツールには、以下のサブコマンドがあります。

```console
$ s3tdl download {S3 Tables bucket ARN} [...]
$ s3tdl inspect {S3 Tables bucket ARN}
```

### `download` サブコマンド

`download` サブコマンドに S3 Tables バケットのARNを指定すると、
テーブルを構成する Parquet ファイルをダウンロードします。

```console
$ s3tdl download {S3 Tables bucket ARN}
```

デフォルトでは、S3 Tables バケット内のすべての名前空間 (Namespace) の全テーブルがダウンロード対象となります。
対象を一部に制限したい場合は、-namespace や -table オプションで正規表現フィルターを指定してください。

#### オプション

| オプション   | 説明                                                           |
|--------------|----------------------------------------------------------------|
| `-dryrun`    | 実際のダウンロードを行わずに動作を確認                         |
| `-namespace` | ダウンロード対象の Namespace を正規表現でフィルター            |
| `-table`     | ダウンロード対象のテーブル名を正規表現でフィルター             |
| `-outdir`    | ファイルの出力先ディレクトリ (デフォルト: `.`)                 |
| `-verbose`   | 詳細ログを表示                                                 |

### `inspect` サブコマンド

`inspect` サブコマンドに S3 Tables バケットのARNを指定すると、
バケットの Iceberg カタログを inspect します。

```console
$ s3tdl inspect {S3 Tables bucket ARN}
```

すべての名前空間を列挙し、各テーブルについて名前空間のプロパティ、テーブルメタデータ、
現在のスナップショット、スキーマ、ソート順、パーティション仕様を表示します。
読み取り専用で、データファイルはダウンロードしません。

このサブコマンドは、バケットARN以外のオプションは受け付けません。

### 認証情報

AWS へのアクセスには、AWS SDK for Go v2 標準の資格情報プロバイダチェーンを使用します。
以下の順序等で自動取得されます。

-   環境変数の静的資格情報: `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `AWS_SESSION_TOKEN`
-   Web Identity Token: `AWS_WEB_IDENTITY_TOKEN_FILE`
-   共有設定ファイル: `~/.aws/credentials`, `~/.aws/config`, and `AWS_PROFILE`
-   ECSタスク: タスクに割り当てられた IAM ロール
-   EC2: EC2インスタンスに割り当てられた IAM ロール

詳細については [AWS SDK for Go v2 Developer Guide](https://docs.aws.amazon.com/sdk-for-go/v2/developer-guide/configure-gosdk.html) を参照してください。
