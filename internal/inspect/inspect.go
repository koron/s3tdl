// Package inspect provides S3Tables (Iceberg catalog) inspection.
package inspect

import (
	"context"
	"fmt"
	"iter"
	"log"
	"slices"
	"strings"
	"time"

	"github.com/apache/iceberg-go"
	icebergio "github.com/apache/iceberg-go/io"
	"github.com/apache/iceberg-go/table"
	"github.com/jedib0t/go-pretty/v6/list"
	"github.com/koron-go/subcmd"
	"github.com/koron/s3tdl/internal/common"
)

var (
	verbose bool
	deleted bool
)

var Inspect = subcmd.DefineCommand("inspect", "inspect S3 Tables (Iceberg catalog)", inspectCommand)

func inspectCommand(ctx context.Context, args []string) error {
	fs := subcmd.FlagSet(ctx)
	common.InitFlagSet(fs)
	fs.BoolVar(&verbose, "verbose", false, "show detailed stats")
	fs.BoolVar(&deleted, "deleted", false, "show deleted data files")
	fs.Parse(args)
	return inspectCatalog(ctx)
}

func collectSeq2[V any](it iter.Seq2[V, error]) ([]V, error) {
	var arr []V
	for v, err := range it {
		if err != nil {
			return nil, err
		}
		arr = append(arr, v)
	}
	return arr, nil
}

type tsmillis int64

func (ms tsmillis) String() string {
	n := int64(ms)
	return fmt.Sprintf("%d (%s)", n, time.UnixMilli(n).Format(time.RFC3339))
}

type listWriter struct {
	list.Writer
}

func (lw listWriter) Append(v any) {
	lw.Writer.AppendItem(v)
}

func (lw listWriter) Appendf(format string, a ...any) {
	lw.Writer.AppendItem(fmt.Sprintf(format, a...))
}

func (lw listWriter) IndentFunc(fn func(listWriter)) {
	lw.Writer.Indent()
	fn(lw)
	lw.Writer.UnIndent()
}

func inspectCatalog(ctx context.Context) error {
	cat, err := common.DefaultCatalog(ctx)
	if err != nil {
		return err
	}

	namespaces, err := cat.ListNamespaces(ctx, nil)
	if err != nil {
		return err
	}

	lw := listWriter{Writer: list.NewWriter()}

	lw.Appendf("Namespaces (%d):", len(namespaces))
	lw.Indent()

	for _, ns := range namespaces {
		tableIDs, err := collectSeq2(cat.ListTables(ctx, ns))
		if err != nil {
			return err
		}
		lw.Appendf("%s (tables: %d)", common.ID2Str(ns), len(tableIDs))
		lw.Indent()

		// Namespace properties:
		props, err := cat.LoadNamespaceProperties(ctx, ns)
		if err != nil {
			log.Printf("failed to LoadNamespaceProperties: %s", err)
		}
		appendProperties(lw, props)

		for _, tableID := range tableIDs {
			table, err := cat.LoadTable(ctx, tableID)
			if err != nil {
				log.Printf("failed to LoadTable: %s", err)
				continue
			}

			lw.AppendItem(common.ID2Str(tableID[len(ns):]))
			lw.Indent()

			lw.Appendf("Identifier: %s", common.ID2Str(table.Identifier()))
			appendMetadata(lw, table.Metadata(), table.MetadataLocation())

			schema := table.Schema()
			lw.Appendf("Current Schema: (ID: %d)", schema.ID)
			lw.Indent()
			for _, f := range schema.Fields() {
				lw.Appendf("[%d] %s: %s (%s)", f.ID, f.Name, f.Type, optOrReq(f.Required))
			}
			lw.UnIndent()

			// Partition Spec:
			partition := table.Spec()
			if partition.IsUnpartitioned() {
				lw.AppendItem("Current Partition Spec: unpartitioned")
			} else {
				lw.Appendf("Partition Spec: (ID: %d)", partition.ID())
				lw.Indent()
				for _, pf := range partition.Fields() {
					src := sourceFields(schema, pf.SourceIDs)
					lw.Appendf("%s: %s(%s)", pf.Name, pf.Transform, src)
				}
				lw.UnIndent()
			}

			// Sort order
			if sortOrder := table.SortOrder(); !sortOrder.IsUnsorted() {
				lw.Appendf("Current Sort Order: (ID: %d)", sortOrder.OrderID())
				lw.Indent()
				for _, field := range sortOrder.Fields() {
					lw.AppendItem(field)
				}
				lw.UnIndent()
			}

			// Current Snapshot:
			snapshot := table.CurrentSnapshot()
			if snapshot != nil {
				// Manifest List
				tableIO, err := table.FS(ctx)
				if err != nil {
					return err
				}
				lw.Append("Current Snapshot:")
				lw.IndentFunc(func(lw listWriter) {
					appendSnapshot(lw, snapshot, tableIO)
				})
			}

			lw.UnIndent()
		}

		lw.UnIndent()
	}

	style := list.StyleConnectedLight
	style.CharItemSingle = style.CharItemBottom
	style.CharItemTop = style.CharItemFirst
	lw.SetStyle(style)

	fmt.Printf("Warehouse: %s\n", cat.Config.Warehouse)
	fmt.Println(lw.Render())

	return nil
}

func sourceFields(schema *iceberg.Schema, ids []int) string {
	if len(ids) == 0 {
		return "N/A"
	}
	fields := make([]string, 0, len(ids))
	for _, id := range ids {
		if f, ok := schema.FindFieldByID(id); ok {
			fields = append(fields, f.Name)
		}
	}
	return strings.Join(fields, ", ")
}

func optOrReq(required bool) string {
	if required {
		return "required"
	}
	return "optional"
}

func appendMetadata(lw listWriter, metadata table.Metadata, loc string) {
	if metadata == nil {
		return
	}
	if loc == "" {
		lw.Append("Metadata:")
	} else {
		lw.Appendf("Metadata: (location: %s)", loc)
	}
	lw.IndentFunc(func(lw listWriter) {
		lw.Appendf("Version: V%d", metadata.Version())
		lw.Appendf("Table UUID: %v", metadata.TableUUID())
		lw.Appendf("Location: %s", metadata.Location())
		lw.Appendf("Last Updated Millis: %s", tsmillis(metadata.LastUpdatedMillis()))
		lw.Appendf("Last Column ID: %d", metadata.LastColumnID())
		if schema := metadata.CurrentSchema(); schema != nil {
			lw.Appendf("Current Schema ID: %d", schema.ID)
		}
		lw.Appendf("Default Partition Spec: %d", metadata.DefaultPartitionSpec())
		if snapshot := metadata.CurrentSnapshot(); snapshot != nil {
			lw.Appendf("Current Snapshot ID: %d", snapshot.SnapshotID)
		}
		appendProperties(lw, metadata.Properties())
		lw.Appendf("Last Sequence Number: %d", metadata.LastSequenceNumber())
		if metadata.Version() >= 3 {
			lw.Appendf("Next Row ID: %d", metadata.NextRowID())
		}
	})
}

func appendSnapshot(lw listWriter, snapshot *table.Snapshot, tableIO icebergio.IO) {
	lw.Appendf("Snapshot ID: %d", snapshot.SnapshotID)
	if snapshot.ParentSnapshotID != nil {
		lw.Appendf("Parent Snapshot ID: %d", snapshot.ParentSnapshotID)
	}
	lw.Appendf("Sequence Number: %d", snapshot.SequenceNumber)
	lw.Appendf("Timestamp MS: %s", tsmillis(snapshot.TimestampMs))
	if snapshot.ManifestList != "" {
		lw.Appendf("Manifest List: %s", snapshot.ManifestList)
		lw.Indent()
		if err := appendManifestList(lw, snapshot, tableIO); err != nil {
			log.Printf("failed to append ManifestList: %s", err)
		}
		lw.UnIndent()
	}
	// Summary
	if snapshot.Summary != nil {
		lw.AppendItem("Summary:")
		lw.IndentFunc(func(lw listWriter) {
			lw.Appendf("Operation: %s", snapshot.Summary.Operation)
			appendProperties(lw, snapshot.Summary.Properties)
		})
	}
	if snapshot.SchemaID != nil {
		lw.Appendf("Schema ID: %d", *snapshot.SchemaID)
	}
	if snapshot.FirstRowID != nil {
		lw.Appendf("First Row ID: %d", snapshot.FirstRowID)
	}
	if snapshot.AddedRows != nil {
		lw.Appendf("Added Rows: %d", snapshot.AddedRows)
	}
}

func appendProperties(lw listWriter, props iceberg.Properties) {
	if len(props) == 0 {
		return
	}
	lw.Appendf("Properties (%d):", len(props))
	lw.Indent()
	defer lw.UnIndent()
	keys := make([]string, 0, len(props))
	for k := range props {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	for _, k := range keys {
		lw.Appendf("%s: %s", k, props[k])
	}
}

func appendManifestList(lw listWriter, snapshot *table.Snapshot, tableIO icebergio.IO) error {
	mfList, err := snapshot.Manifests(tableIO)
	if err != nil {
		return err
	}

	for mfIdx, mf := range mfList {
		lw.Appendf("[%d] Manifest", mfIdx)
		lw.Indent()
		lw.Appendf("Version: %d", mf.Version())
		lw.Appendf("File Path: %s", mf.FilePath())
		lw.Appendf("Length: %d", mf.Length())
		lw.Appendf("Partition Spec ID: %d", mf.PartitionSpecID())
		lw.Appendf("Snapshot ID: %d", mf.SnapshotID())
		lw.Appendf("Added Data Files: %d", mf.AddedDataFiles())
		lw.Appendf("Existing Data Files: %d", mf.ExistingDataFiles())
		if deleted {
			lw.Appendf("Deleted Data Files: %d", mf.DeletedDataFiles())
		}
		lw.Appendf("Added Rows: %d", mf.AddedRows())
		lw.Appendf("Existing Rows: %d", mf.ExistingRows())
		if deleted {
			lw.Appendf("Deleted Rows: %d", mf.DeletedRows())
		}
		lw.Appendf("Sequence Number: %d", mf.SequenceNum())
		lw.Appendf("Min Sequence Num: %d", mf.MinSequenceNum())

		lw.Append("Manifest Entries:")
		lw.Indent()
		var meIdx = 0
		for me, err := range mf.Entries(tableIO, !deleted) {
			if err != nil {
				return err
			}
			lw.Appendf("[%d] Manifest Entry", meIdx)
			lw.Indent()
			lw.Appendf("Status: %v", me.Status())
			lw.Appendf("Snapshot ID: %d", me.SnapshotID())
			lw.Appendf("Sequence Num: %d", me.SequenceNum())
			if p := me.FileSequenceNum(); p != nil {
				lw.Appendf("File SequenceNum: %d", *p)
			}
			appendDataFile(lw, me.DataFile())
			lw.UnIndent()
			meIdx++
		}
		lw.UnIndent()

		lw.UnIndent()
	}
	return nil
}

func appendDataFile(lw listWriter, df iceberg.DataFile) {
	if df == nil {
		return
	}
	lw.AppendItem("Data File")
	lw.Indent()
	lw.Appendf("Content Type: %v", df.ContentType())
	lw.Appendf("File Path: %s", df.FilePath())
	lw.Appendf("File Format: %v", df.FileFormat())
	if verbose {
		lw.Appendf("Partition: %v", df.Partition())
	}
	lw.Appendf("Count: %d", df.Count())
	lw.Appendf("File Size Bytes: %d", df.FileSizeBytes())

	if verbose {
		lw.Appendf("Column Sizes: %v", df.ColumnSizes())
		lw.Appendf("Value Counts: %v", df.ValueCounts())
		lw.Appendf("Null Value Counts: %v", df.NullValueCounts())
		lw.Appendf("NaN Value Counts: %v", df.NaNValueCounts())
		lw.Appendf("Distinct Value Counts: %v", df.DistinctValueCounts())
		lw.Appendf("Lower Bound Values: %v", df.LowerBoundValues())
		lw.Appendf("Upper Bound Values: %v", df.UpperBoundValues())
	}

	lw.UnIndent()
}
