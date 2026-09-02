# koron/s3tdl - S3 Tables Data File Downloader

[![PkgGoDev](https://pkg.go.dev/badge/github.com/koron/s3tdl)](https://pkg.go.dev/github.com/koron/s3tdl)
[![Actions/Go](https://github.com/koron/s3tdl/actions/workflows/go.yml/badge.svg)](https://github.com/koron/s3tdl/actions/workflows/go.yml)
[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/koron/s3tdl)

[Engilish](README.md) | [日本語](README-ja.md)

Iceberg REST カタログ内のテーブルを構成するParquetファイルをダウンロードし、
カタログを inspect できるツールです。
YAML 設定ファイルを通じて、AWS S3 Tables 以外の Iceberg REST カタログ (例: Cloudflare R2) でも動作します。

## インストール

Go がインストールされている環境では、以下のコマンドでインストール・更新できます。

```console
$ go install github.com/koron/s3tdl@latest
```

または [最新のリリース](https://github.com/koron/s3tdl/releases/latest) からコンパイル済みバイナリを直接ダウンロードすることもできます。

## 設定

カタログへのアクセスは、YAML 設定ファイル (デフォルト: `tmp/config.yaml`) によって駆動されます。
トップレベルのキーは `catalog` で、名前をキーとするカタログエントリのマップです:

```yaml
catalog:
  default:
    uri: https://<account-id>.s3tables.<region>.amazonaws.com/iceberg
    warehouse: arn:aws:s3tables:us-east-1:<account-id>:bucket/<bucket-name>
    aws:
      profile: myprofile
      region: us-east-1
      s3-endpoint: https://<account>.r2.cloudflarestorage.com # 省略可
    rest:
      sigv4-enabled: true
      signing-name: s3tables
      signing-region: us-east-1
```

| キー | 説明 |
| --- | --- |
| `uri` | Iceberg REST カタログのエンドポイント |
| `warehouse` | ウェアハウス位置 (例: S3 Tables バケットのARN) |
| `token` | REST カタログ用の静的 OAuth トークン (省略可) |
| `credential` | REST カタログ用の OAuth 認証情報 (省略可、`token` と相互排他) |
| `scope` | REST カタログ用の OAuth scope (省略可) |
| `aws.profile` | 使用する AWS 共有設定プロファイル (省略可) |
| `aws.region` | 使用する AWS リージョン (省略可) |
| `aws.s3-endpoint` | データファイルのダウンロードに使用する代替 S3 エンドポイント。例: Cloudflare R2 (省略可) |
| `rest.sigv4-enabled` | REST リクエストの SigV4 署名を有効にする (AWS S3 Tables) |
| `rest.signing-name` | SigV4 署名用のサービス名 (例: `s3tables`) |
| `rest.signing-region` | SigV4 署名用のリージョン |

## 使い方

このツールには、以下のサブコマンドがあります。両方とも `-config` と `-catalog` オプションで上記の設定を読み込みます:

```console
$ s3tdl download [-config PATH] [-catalog NAME] [-namespace RE] [-table RE] [-outdir DIR] [-dryrun] [-verbose]
$ s3tdl inspect  [-config PATH] [-catalog NAME] [-verbose] [-deleted]
```

| オプション   | デフォルト       | 説明                               |
|--------------|------------------|------------------------------------|
| `-config`    | `tmp/config.yaml`| 設定 YAML の場所                   |
| `-catalog`   | `default`        | 設定から使用するカタログ名         |

### `download` サブコマンド

`download` サブコマンドを実行すると、カタログ内のテーブルを構成する Parquet ファイルをダウンロードします。

```console
$ s3tdl download -namespace rx -table rx -outdir ./out
```

デフォルトでは、カタログ内のすべての名前空間 (Namespace) の全テーブルがダウンロード対象となります。
対象を一部に制限したい場合は、-namespace や -table オプションで正規表現フィルターを指定してください。

ファイルは `-outdir/<namespace>/<table>/<ファイルベース名>` に書き出されます。
各データファイルのベース名のみが使用されるため、異なる S3 プレフィックス配下の同名ファイルは上書きし合う点にご注意ください。

#### オプション

| オプション   | 説明                                                           |
|--------------|----------------------------------------------------------------|
| `-config`    | 設定 YAML の場所 (デフォルト: `tmp/config.yaml`)               |
| `-catalog`   | 設定から使用するカタログ名 (デフォルト: `default`)             |
| `-dryrun`    | 実際のダウンロードを行わずに動作を確認                         |
| `-namespace` | ダウンロード対象の Namespace を正規表現でフィルター            |
| `-table`     | ダウンロード対象のテーブル名を正規表現でフィルター             |
| `-outdir`    | ファイルの出力先ディレクトリ (デフォルト: `.`)                 |
| `-verbose`   | 詳細ログを表示                                                 |

`-dryrun` を使うと、実際にダウンロードせずにダウンロード対象のファイルを列挙できます。
実際のカタログに対して設定を検証する最安の方法です。

### `inspect` サブコマンド

`inspect` サブコマンドを実行すると、Iceberg カタログを inspect します。

```console
$ s3tdl inspect
```

すべての名前空間を列挙し、各テーブルについて名前空間のプロパティ、テーブルメタデータ、
現在のスナップショット、スキーマ、ソート順、パーティション仕様、およびマニフェストリストを表示します。
読み取り専用で、データファイルはダウンロードしません。

#### オプション

| オプション   | 説明                                                           |
|--------------|----------------------------------------------------------------|
| `-config`    | 設定 YAML の場所 (デフォルト: `tmp/config.yaml`)               |
| `-catalog`   | 設定から使用するカタログ名 (デフォルト: `default`)             |
| `-verbose`   | データファイルの詳細統計を表示                                 |
| `-deleted`   | マニフェストエントリに削除済みデータファイルを含める           |

### 認証情報

AWS へのアクセスには、AWS SDK for Go v2 標準の資格情報プロバイダチェーンを使用します。
以下の順序等で自動取得されます。

-   環境変数の静的資格情報: `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `AWS_SESSION_TOKEN`
-   Web Identity Token: `AWS_WEB_IDENTITY_TOKEN_FILE`
-   共有設定ファイル: `~/.aws/credentials`, `~/.aws/config`, and `AWS_PROFILE`
-   ECSタスク: タスクに割り当てられた IAM ロール
-   EC2: EC2インスタンスに割り当てられた IAM ロール

設定の `aws.profile` と `aws.region` キーにより、特定のカタログについてプロファイルとリージョンを上書きできます。
詳細については [AWS SDK for Go v2 Developer Guide](https://docs.aws.amazon.com/sdk-for-go/v2/developer-guide/configure-gosdk.html) を参照してください。
