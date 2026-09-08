# `-verbose` で Data File に追加表示される内容の意味

参考: <https://iceberg.apache.org/spec/#data-file-fields>

-   `Partition` - パーティションのデータのタプル。
    `partition specのfield id` + `なぞの数字`
-   `Column Sizes` - カラム毎の、総データサイズ。
    `column id` + `バイト数` の形式。
    行指向フォーマットでは `null`
-   `Value Counts` - カラム毎の行数。 `null` や `NaN` の行を含む
-   `Null Value Counts` - カラム毎の `null` の数
-   `NaN Value Counts` - カラム毎の `NaN` の数
-   `Distinct Value Counts` - Deprecated
-   `Lower Bound Values` - カラム毎のシリアライズされた値の最小値
-   `Upper Bound Values` - カラム毎のシリアライズされた値の最大値

これらの値はoptionalなので、空の時は表示しないようにする
