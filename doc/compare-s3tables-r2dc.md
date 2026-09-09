# S3 Tables と R2 Data Catalog の比較

S3 Tables と R2 Data Catalog の違いを見るための実験記録。

## 基本的なテーブル

### 準備

S3とR2双方で、DuckDBを用いて、テーブルを作成しデータを投入。

実行コマンドの例

```console
$ ./s3tdl inspect -config ./tmp/config.yaml -catalog default -verbose
```

テーブル作成

```sql
CREATE TABLE daily_sales (sale_date date,
        product_category string,
        sales_amount double)
PARTITIONED BY (month(sale_date))
    TBLPROPERTIES ('table_type' = 'iceberg');
```

-   `TBLPROPERTIES ...` は S3 のみで、R2 には指定していない(できない)。
-   `sale_date` を月毎でパーティションしている

データ投入

```sql
INSERT INTO daily_sales VALUES
    (DATE '2024-01-15', 'Laptop', 900.00),
    (DATE '2024-01-15', 'Monitor', 250.00),
    (DATE '2024-01-16', 'Laptop', 1350.00),
    (DATE '2024-02-01', 'Monitor', 300.00),
    (DATE '2024-02-01', 'Keyboard', 60.00),
    (DATE '2024-02-02', 'Mouse', 25.00),
    (DATE '2024-02-02', 'Laptop', 1050.00),
    (DATE '2024-02-03', 'Laptop', 1200.00),
    (DATE '2024-02-03', 'Monitor', 375.00);
```

クエリ

```sql
SELECT product_category, COUNT(*) as units_sold,
       SUM(sales_amount) as total_revenue,
       AVG(sales_amount) as average_price
   FROM daily_sales
       WHERE sale_date BETWEEN DATE '2024-02-01' and DATE '2024-02-29'
       GROUP BY product_category
       ORDER BY total_revenue DESC;
```

### inspect 結果(抜粋)

<details>
<summary>S3 Tables</summary>

```
└─ daily_sales
   ├─ Identifier: t1.daily_sales
   ├─ Metadata: (location: s3://.../metadata/00001-87166fdc-cc87-41ed-96de-ab18bfea0b1d.metadata.json)
   │  ├─ Version: V2
   │  ├─ Table UUID: 67017a69-165f-4139-8609-ceab131065fa
   │  ├─ Location: s3://...
   │  ├─ Last Updated Millis: 1782271193198 (2026-06-24T12:19:53+09:00)
   │  ├─ Last Column ID: 3
   │  ├─ Current Schema ID: 0
   │  ├─ Default Partition Spec: 0
   │  ├─ Current Snapshot ID: 4489985133624857397
   │  ├─ Properties (2):
   │  │  ├─ write.object-storage.enabled: true
   │  │  └─ write.parquet.compression-codec: zstd
   │  └─ Last Sequence Number: 1
   ├─ Current Schema: (ID: 0)
   │  ├─ [1] sale_date: date (optional)
   │  ├─ [2] product_category: string (optional)
   │  └─ [3] sales_amount: double (optional)
   ├─ Partition Spec: (ID: 0)
   │  └─ sale_date_month (Field ID:1000) : month(sale_date)
   └─ Current Snapshot:
      ├─ Snapshot ID: 4489985133624857397
      ├─ Sequence Number: 1
      ├─ Timestamp MS: 1782271193198 (2026-06-24T12:19:53+09:00)
      ├─ Manifest List: s3://.../metadata/snap-4489985133624857397-1-d9897fb6-ed6c-466e-9d2b-608c39dfa292.avro
      │  └─ [0] Manifest
      │     ├─ Version: 2
      │     ├─ File Path: s3://.../metadata/d9897fb6-ed6c-466e-9d2b-608c39dfa292-m0.avro
      │     ├─ Length: 7188
      │     ├─ Partition Spec ID: 0
      │     ├─ Snapshot ID: 4489985133624857397
      │     ├─ Added Data Files: 2
      │     ├─ Existing Data Files: 0
      │     ├─ Added Rows: 9
      │     ├─ Existing Rows: 0
      │     ├─ Sequence Number: 1
      │     ├─ Min Sequence Num: 1
      │     └─ Manifest Entries:
      │        ├─ [0] Manifest Entry
      │        │  ├─ Status: 1:ADDED
      │        │  ├─ Snapshot ID: 4489985133624857397
      │        │  ├─ Sequence Num: 1
      │        │  ├─ File SequenceNum: 1
      │        │  └─ Data File
      │        │     ├─ Content Type: Data
      │        │     ├─ File Path: s3://.../data/lI6hLQ/sale_date_month=2024-01/20260624_031951_00160_s5zfz-c6bafe08-bfed-45fc-8b3e-ce116f871d54.parquet
      │        │     ├─ File Format: PARQUET
      │        │     ├─ Partition: map[1000:648]
      │        │     ├─ Count: 3
      │        │     ├─ File Size Bytes: 647
      │        │     ├─ Column Sizes: map[1:73 2:86 3:56]
      │        │     ├─ Value Counts: map[1:3 2:3 3:3]
      │        │     ├─ Null Value Counts: map[1:0 2:0 3:0]
      │        │     ├─ Lower Bound Values: map[1:[25 77 0 0] 2:[76 97 112 116 111 112] 3:[0 0 0 0 0 64 111 64]]
      │        │     └─ Upper Bound Values: map[1:[26 77 0 0] 2:[77 111 110 105 116 111 114] 3:[0 0 0 0 0 24 149 64]]
      │        └─ [1] Manifest Entry
      │           ├─ Status: 1:ADDED
      │           ├─ Snapshot ID: 4489985133624857397
      │           ├─ Sequence Num: 1
      │           ├─ File SequenceNum: 1
      │           └─ Data File
      │              ├─ Content Type: Data
      │              ├─ File Path: s3://.../data/QaNS4Q/sale_date_month=2024-02/20260624_031951_00160_s5zfz-2b3fc489-64bf-4527-9914-855ee19c3b2d.parquet
      │              ├─ File Format: PARQUET
      │              ├─ Partition: map[1000:649]
      │              ├─ Count: 6
      │              ├─ File Size Bytes: 686
      │              ├─ Column Sizes: map[1:78 2:108 3:65]
      │              ├─ Value Counts: map[1:6 2:6 3:6]
      │              ├─ Null Value Counts: map[1:0 2:0 3:0]
      │              ├─ Lower Bound Values: map[1:[42 77 0 0] 2:[75 101 121 98 111 97 114 100] 3:[0 0 0 0 0 0 57 64]]
      │              └─ Upper Bound Values: map[1:[44 77 0 0] 2:[77 111 117 115 101] 3:[0 0 0 0 0 192 146 64]]
      ├─ Summary:
      │  ├─ Operation: append
      │  └─ Properties (11):
      │     ├─ added-data-files: 2
      │     ├─ added-files-size: 1333
      │     ├─ added-records: 9
      │     ├─ changed-partition-count: 2
      │     ├─ total-data-files: 2
      │     ├─ total-delete-files: 0
      │     ├─ total-equality-deletes: 0
      │     ├─ total-files-size: 1333
      │     ├─ total-position-deletes: 0
      │     ├─ total-records: 9
      │     └─ trino_query_id: 20260624_031951_00160_s5zfz
      └─ Schema ID: 0
```

</details>

<details>
<summary>R2 Data Catalog</summary>

```
└─ daily_sales
   ├─ Identifier: default.daily_sales
   ├─ Metadata: (location: s3://.../__r2_data_catalog/01a01806-2161-7883-bc04-a443ba0ce53e/01a084a0-54b7-7be0-9d37-e790fb0b1012/metadata/00001-01a084a0-97bd-7692-8303-fcd3188e55f2.gz.metadata.json)
   │  ├─ Version: V2
   │  ├─ Table UUID: 01a084a0-54b7-7be0-9d37-e790fb0b1012
   │  ├─ Location: s3://.../__r2_data_catalog/01a01806-2161-7883-bc04-a443ba0ce53e/01a084a0-54b7-7be0-9d37-e790fb0b1012
   │  ├─ Last Updated Millis: 1788931511998 (2026-09-09T14:25:11+09:00)
   │  ├─ Last Column ID: 3
   │  ├─ Current Schema ID: 0
   │  ├─ Default Partition Spec: 0
   │  ├─ Current Snapshot ID: 3382049451819024693
   │  └─ Last Sequence Number: 1
   ├─ Current Schema: (ID: 0)
   │  ├─ [1] sale_date: date (optional)
   │  ├─ [2] product_category: string (optional)
   │  └─ [3] sales_amount: double (optional)
   ├─ Partition Spec: (ID: 0)
   │  └─ month_sale_date_1 (Field ID:1000) : month(sale_date)
   └─ Current Snapshot:
      ├─ Snapshot ID: 3382049451819024693
      ├─ Sequence Number: 1
      ├─ Timestamp MS: 1788931511998 (2026-09-09T14:25:11+09:00)
      ├─ Manifest List: s3://.../__r2_data_catalog/01a01806-2161-7883-bc04-a443ba0ce53e/01a084a0-54b7-7be0-9d37-e790fb0b1012/metadata/snap-3382049451819024693-a0d8c37b-a52c-4487-8500-7e1153f610a7.avro
      │  └─ [0] Manifest
      │     ├─ Version: 2
      │     ├─ File Path: s3://.../__r2_data_catalog/01a01806-2161-7883-bc04-a443ba0ce53e/01a084a0-54b7-7be0-9d37-e790fb0b1012/metadata/48151aa4-aa78-4814-8dd2-81ffe32b6124-m0.avro
      │     ├─ Length: 2765
      │     ├─ Partition Spec ID: 0
      │     ├─ Snapshot ID: 3382049451819024693
      │     ├─ Added Data Files: 2
      │     ├─ Existing Data Files: 0
      │     ├─ Added Rows: 9
      │     ├─ Existing Rows: 0
      │     ├─ Sequence Number: 1
      │     ├─ Min Sequence Num: 1
      │     └─ Manifest Entries:
      │        ├─ [0] Manifest Entry
      │        │  ├─ Status: 1:ADDED
      │        │  ├─ Snapshot ID: 3382049451819024693
      │        │  ├─ Sequence Num: 1
      │        │  ├─ File SequenceNum: 1
      │        │  └─ Data File
      │        │     ├─ Content Type: Data
      │        │     ├─ File Path: s3://.../__r2_data_catalog/01a01806-2161-7883-bc04-a443ba0ce53e/01a084a0-54b7-7be0-9d37-e790fb0b1012/data/month_sale_date_1=648/01a084a0-9551-75c0-b9f4-02c7d10fdf4e.parquet
      │        │     ├─ File Format: parquet
      │        │     ├─ Partition: map[1000:648]
      │        │     ├─ Count: 3
      │        │     ├─ File Size Bytes: 538
      │        │     ├─ Null Value Counts: map[1:0 2:0 3:0]
      │        │     ├─ Lower Bound Values: map[1:[25 77 0 0] 2:[76 97 112 116 111 112] 3:[0 0 0 0 0 64 111 64]]
      │        │     └─ Upper Bound Values: map[1:[26 77 0 0] 2:[77 111 110 105 116 111 114] 3:[0 0 0 0 0 24 149 64]]
      │        └─ [1] Manifest Entry
      │           ├─ Status: 1:ADDED
      │           ├─ Snapshot ID: 3382049451819024693
      │           ├─ Sequence Num: 1
      │           ├─ File SequenceNum: 1
      │           └─ Data File
      │              ├─ Content Type: Data
      │              ├─ File Path: s3://.../__r2_data_catalog/01a01806-2161-7883-bc04-a443ba0ce53e/01a084a0-54b7-7be0-9d37-e790fb0b1012/data/month_sale_date_1=649/01a084a0-9552-7780-ade6-603d6e3238ae.parquet
      │              ├─ File Format: parquet
      │              ├─ Partition: map[1000:649]
      │              ├─ Count: 6
      │              ├─ File Size Bytes: 592
      │              ├─ Null Value Counts: map[1:0 2:0 3:0]
      │              ├─ Lower Bound Values: map[1:[42 77 0 0] 2:[75 101 121 98 111 97 114 100] 3:[0 0 0 0 0 0 57 64]]
      │              └─ Upper Bound Values: map[1:[44 77 0 0] 2:[77 111 117 115 101] 3:[0 0 0 0 0 192 146 64]]
      ├─ Summary:
      │  ├─ Operation: append
      │  └─ Properties (8):
      │     ├─ added-data-files: 2
      │     ├─ added-records: 9
      │     ├─ deleted-data-files: 0
      │     ├─ deleted-records: 0
      │     ├─ total-data-files: 2
      │     ├─ total-delete-files: 0
      │     ├─ total-position-deletes: 0
      │     └─ total-records: 9
      └─ Schema ID: 0
```

</details>

注目ポイント

#### 各データのパス構成の違い

S3 ではバケツ直下に `metadata/` と `data/` のサブキーが置かれる。
R2 では `__r2_data_catalog/{バケツID?}/` という中間キーの後に、
`metadata/` と `data/` のサブキーが置かれる。

```
# S3
   ├─ Metadata: (location: s3://.../metadata/00001-87166fdc-cc87-41ed-96de-ab18bfea0b1d.metadata.json)
   │  ├─ Location: s3://...
      ├─ Manifest List: s3://.../metadata/snap-4489985133624857397-1-d9897fb6-ed6c-466e-9d2b-608c39dfa292.avro
      │     ├─ File Path: s3://.../metadata/d9897fb6-ed6c-466e-9d2b-608c39dfa292-m0.avro
      │        │     ├─ File Path: s3://.../data/lI6hLQ/sale_date_month=2024-01/20260624_031951_00160_s5zfz-c6bafe08-bfed-45fc-8b3e-ce116f871d54.parquet
      │              ├─ File Path: s3://.../data/QaNS4Q/sale_date_month=2024-02/20260624_031951_00160_s5zfz-2b3fc489-64bf-4527-9914-855ee19c3b2d.parquet

# R2
   ├─ Metadata: (location: s3://.../__r2_data_catalog/01a01806-2161-7883-bc04-a443ba0ce53e/01a084a0-54b7-7be0-9d37-e790fb0b1012/metadata/00001-01a084a0-97bd-7692-8303-fcd3188e55f2.gz.metadata.json)
   │  ├─ Location: s3://.../__r2_data_catalog/01a01806-2161-7883-bc04-a443ba0ce53e/01a084a0-54b7-7be0-9d37-e790fb0b1012
      ├─ Manifest List: s3://.../__r2_data_catalog/01a01806-2161-7883-bc04-a443ba0ce53e/01a084a0-54b7-7be0-9d37-e790fb0b1012/metadata/snap-3382049451819024693-a0d8c37b-a52c-4487-8500-7e1153f610a7.avro
      │     ├─ File Path: s3://.../__r2_data_catalog/01a01806-2161-7883-bc04-a443ba0ce53e/01a084a0-54b7-7be0-9d37-e790fb0b1012/metadata/48151aa4-aa78-4814-8dd2-81ffe32b6124-m0.avro
      │        │     ├─ File Path: s3://.../__r2_data_catalog/01a01806-2161-7883-bc04-a443ba0ce53e/01a084a0-54b7-7be0-9d37-e790fb0b1012/data/month_sale_date_1=648/01a084a0-9551-75c0-b9f4-02c7d10fdf4e.parquet
      │              ├─ File Path: s3://.../__r2_data_catalog/01a01806-2161-7883-bc04-a443ba0ce53e/01a084a0-54b7-7be0-9d37-e790fb0b1012/data/month_sale_date_1=649/01a084a0-9552-7780-ade6-603d6e3238ae.parquet
```

#### テーブルのメタデータのプロパティ

S3にはあるが、R2にはない。

```
# S3
└─ daily_sales
   ├─ Metadata: (location: s3://.../metadata/00001-87166fdc-cc87-41ed-96de-ab18bfea0b1d.metadata.json)
(...中略...)
   │  ├─ Current Snapshot ID: 4489985133624857397
   │  ├─ Properties (2):
   │  │  ├─ write.object-storage.enabled: true
   │  │  └─ write.parquet.compression-codec: zstd
   │  └─ Last Sequence Number: 1

# R2
└─ daily_sales
   ├─ Identifier: default.daily_sales
   ├─ Metadata: (location: s3://.../__r2_data_catalog/01a01806-2161-7883-bc04-a443ba0ce53e/01a084a0-54b7-7be0-9d37-e790fb0b1012/metadata/00001-01a084a0-97bd-7692-8303-fcd3188e55f2.gz.metadata.json)
(...中略...)
   │  ├─ Current Snapshot ID: 3382049451819024693
   │  └─ Last Sequence Number: 1
```

名前空間に対するプロパティは、R2にはあるがS3にはない。

#### ファイルフォーマットの表記

ファイルフォーマット(マニフェストのデータファイル)の表記が、
S3は大文字だがR2は小文字になっている。

```
# S3
      │        │     ├─ File Format: PARQUET

# R2
      │        │     ├─ File Format: parquet
```

#### データファイルのオプショナル項目

データファイルのオプショナル項目の値が与えられた項目が異なる。
S3のほうが多く、R2のほうが少ない。

```
# S3
      │        │     ├─ Column Sizes: map[1:73 2:86 3:56]
      │        │     ├─ Value Counts: map[1:3 2:3 3:3]
      │        │     ├─ Null Value Counts: map[1:0 2:0 3:0]
      │        │     ├─ Lower Bound Values: map[1:[25 77 0 0] 2:[76 97 112 116 111 112] 3:[0 0 0 0 0 64 111 64]]
      │        │     └─ Upper Bound Values: map[1:[26 77 0 0] 2:[77 111 110 105 116 111 114] 3:[0 0 0 0 0 24 149 64]]

# R2
      │        │     ├─ Null Value Counts: map[1:0 2:0 3:0]
      │        │     ├─ Lower Bound Values: map[1:[25 77 0 0] 2:[76 97 112 116 111 112] 3:[0 0 0 0 0 64 111 64]]
      │        │     └─ Upper Bound Values: map[1:[26 77 0 0] 2:[77 111 110 105 116 111 114] 3:[0 0 0 0 0 24 149 64]]
```

具体的には S3 のほうだけ、 `Column Sizes` と `Value Counts` が与えられている。
これらのオプショナルに依存したロジックは相互運用ができないから、
見る価値が低いということ。

#### スナップショットサマリーのプロパティ

S3のほうが項目数は多いが、R2にしかない項目もある。

```
# S3
      ├─ Summary:
      │  ├─ Operation: append
      │  └─ Properties (11):
      │     ├─ added-data-files: 2
      │     ├─ added-files-size: 1333
      │     ├─ added-records: 9
      │     ├─ changed-partition-count: 2
      │     ├─ total-data-files: 2
      │     ├─ total-delete-files: 0
      │     ├─ total-equality-deletes: 0
      │     ├─ total-files-size: 1333
      │     ├─ total-position-deletes: 0
      │     ├─ total-records: 9
      │     └─ trino_query_id: 20260624_031951_00160_s5zfz

# R2
      ├─ Summary:
      │  ├─ Operation: append
      │  └─ Properties (8):
      │     ├─ added-data-files: 2
      │     ├─ added-records: 9
      │     ├─ deleted-data-files: 0
      │     ├─ deleted-records: 0
      │     ├─ total-data-files: 2
      │     ├─ total-delete-files: 0
      │     ├─ total-position-deletes: 0
      │     └─ total-records: 9
```

本サンプルで共通している項目は以下の6個。

- `added-data-files`
- `added-records`
- `total-data-files`
- `total-delete-files`
- `total-position-deletes`
- `total-records`
