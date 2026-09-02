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

func inspectCatalog(ctx context.Context) error {
	cat, err := common.DefaultCatalog(ctx)
	if err != nil {
		return err
	}

	namespaces, err := cat.ListNamespaces(ctx, nil)
	if err != nil {
		return err
	}

	lw := list.NewWriter()
	lw.AppendItem(fmt.Sprintf("Namespaces (%d):", len(namespaces)))
	lw.Indent()

	for _, ns := range namespaces {
		tableIDs, err := collectSeq2(cat.ListTables(ctx, ns))
		if err != nil {
			return err
		}
		lw.AppendItem(fmt.Sprintf("%s (tables: %d)", common.ID2Str(ns), len(tableIDs)))
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

			metadata := table.Metadata()
			if metadata != nil {
				lw.AppendItem("Metadata:")
				lw.Indent()
				lw.AppendItem(fmt.Sprintf("Version: V%d", metadata.Version()))
				lw.AppendItem(fmt.Sprintf("Table UUID: %v", metadata.TableUUID()))
				lw.AppendItem(fmt.Sprintf("Location: %s", metadata.Location()))
				lw.UnIndent()
			}

			// Current Snapshot:
			snapshot := table.CurrentSnapshot()
			if snapshot != nil {
				lw.AppendItem("Current Snapshot:")
				appendSnapshot(lw, snapshot)
			}

			schema := table.Schema()
			lw.AppendItem(fmt.Sprintf("Schema: (ID: %d)", schema.ID))
			lw.Indent()
			for _, f := range schema.Fields() {
				lw.AppendItem(fmt.Sprintf("[%d] %s: %s (%s)", f.ID, f.Name, f.Type, optOrReq(f.Required)))
			}
			lw.UnIndent()

			sortOrder := table.SortOrder()
			if !sortOrder.IsUnsorted() {
				lw.AppendItem(fmt.Sprintf("Sort Order: (ID: %d)", sortOrder.OrderID()))
				lw.Indent()
				for _, field := range sortOrder.Fields() {
					lw.AppendItem(field)
				}
				lw.UnIndent()
			}

			// Partition Spec:
			partition := table.Spec()
			if partition.IsUnpartitioned() {
				lw.AppendItem("Partition Spec: unpartitioned")
			} else {
				lw.AppendItem(fmt.Sprintf("Partition Spec: (ID: %d)", partition.ID()))
				lw.Indent()
				for _, pf := range partition.Fields() {
					src := sourceFields(schema, pf.SourceIDs)
					lw.AppendItem(fmt.Sprintf("%s: %s(%s)", pf.Name, pf.Transform, src))
				}
				lw.UnIndent()
			}

			// Manifest List
			if snapshot != nil {
				tableIO, err := table.FS(ctx)
				if err != nil {
					return err
				}
				if err := appendManifestList(lw, tableIO, snapshot); err != nil {
					return err
				}
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

func appendSnapshot(lw list.Writer, snapshot *table.Snapshot) {
	lw.Indent()
	lw.AppendItem(fmt.Sprintf("Snapshot ID: %d", snapshot.SnapshotID))
	if snapshot.ParentSnapshotID != nil {
		lw.AppendItem(fmt.Sprintf("Parent Snapshot ID: %d", snapshot.ParentSnapshotID))
	}
	lw.AppendItem(fmt.Sprintf("Sequence Number: %d", snapshot.SequenceNumber))
	lw.AppendItem(fmt.Sprintf("Timestamp MS: %d (%s)",
		snapshot.TimestampMs,
		time.UnixMilli(snapshot.TimestampMs).Format(time.RFC3339)))
	if snapshot.ManifestList != "" {
		lw.AppendItem(fmt.Sprintf("Manifest List: %s", snapshot.ManifestList))
	}
	// Summary
	if snapshot.Summary != nil {
		lw.AppendItem("Summary:")
		lw.Indent()
		lw.AppendItem(fmt.Sprintf("Operation: %s", snapshot.Summary.Operation))
		appendProperties(lw, snapshot.Summary.Properties)
		lw.UnIndent()
	}
	if snapshot.SchemaID != nil {
		lw.AppendItem(fmt.Sprintf("Schema ID: %d", *snapshot.SchemaID))
	}
	if snapshot.FirstRowID != nil {
		lw.AppendItem(fmt.Sprintf("First Row ID: %d", snapshot.FirstRowID))
	}
	if snapshot.AddedRows != nil {
		lw.AppendItem(fmt.Sprintf("Added Rows: %d", snapshot.AddedRows))
	}
	lw.UnIndent()
}

func appendProperties(lw list.Writer, props iceberg.Properties) {
	if len(props) == 0 {
		return
	}
	lw.AppendItem(fmt.Sprintf("Properties (%d):", len(props)))
	lw.Indent()
	keys := make([]string, 0, len(props))
	for k := range props {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	for _, k := range keys {
		lw.AppendItem(fmt.Sprintf("%s: %s", k, props[k]))
	}
	lw.UnIndent()
}

func appendManifestList(lw list.Writer, tableIO icebergio.IO, snapshot *table.Snapshot) error {
	mfList, err := snapshot.Manifests(tableIO)
	if err != nil {
		return err
	}

	lw.AppendItem(fmt.Sprintf("Manifest List (%d)", len(mfList)))
	for mfIdx, mf := range mfList {
		lw.Indent()
		lw.AppendItem(fmt.Sprintf("[%d] Manifest", mfIdx))
		lw.Indent()
		lw.AppendItem(fmt.Sprintf("Version: %d", mf.Version()))
		lw.AppendItem(fmt.Sprintf("FilePath: %s", mf.FilePath()))
		lw.AppendItem(fmt.Sprintf("Length: %d", mf.Length()))
		lw.AppendItem(fmt.Sprintf("PartitionSpecID: %d", mf.PartitionSpecID()))
		lw.AppendItem(fmt.Sprintf("SnapshotID: %d", mf.SnapshotID()))
		lw.AppendItem(fmt.Sprintf("AddedDataFiles: %d", mf.AddedDataFiles()))
		lw.AppendItem(fmt.Sprintf("ExistingDataFiles: %d", mf.ExistingDataFiles()))
		if deleted {
			lw.AppendItem(fmt.Sprintf("DeletedDataFiles: %d", mf.DeletedDataFiles()))
		}
		lw.AppendItem(fmt.Sprintf("AddedRows: %d", mf.AddedRows()))
		lw.AppendItem(fmt.Sprintf("ExistingRows: %d", mf.ExistingRows()))
		if deleted {
			lw.AppendItem(fmt.Sprintf("DeletedRows: %d", mf.DeletedRows()))
		}
		lw.AppendItem(fmt.Sprintf("SequenceNumber: %d", mf.SequenceNum()))
		lw.AppendItem(fmt.Sprintf("MinSequenceNum: %d", mf.MinSequenceNum()))

		meList, err := collectSeq2(mf.Entries(tableIO, !deleted))
		if err != nil {
			return err
		}

		lw.AppendItem(fmt.Sprintf("ManifestEntries (%d)", len(meList)))
		lw.Indent()
		for meIdx, me := range meList {
			lw.AppendItem(fmt.Sprintf("[%d] ManifestEntry", meIdx))
			lw.Indent()
			lw.AppendItem(fmt.Sprintf("Status: %v", me.Status()))
			lw.AppendItem(fmt.Sprintf("SnapshotID: %d", me.SnapshotID()))
			lw.AppendItem(fmt.Sprintf("SequenceNum: %d", me.SequenceNum()))
			if p := me.FileSequenceNum(); p != nil {
				lw.AppendItem(fmt.Sprintf("FileSequenceNum: %d", *p))
			}
			appendDataFile(lw, me.DataFile())
			lw.UnIndent()
		}
		lw.UnIndent()

		lw.UnIndent()
		lw.UnIndent()
	}
	return nil
}

func appendDataFile(lw list.Writer, df iceberg.DataFile) {
	if df == nil {
		return
	}
	lw.AppendItem("DataFile")
	lw.Indent()
	lw.AppendItem(fmt.Sprintf("ContentType: %v", df.ContentType()))
	lw.AppendItem(fmt.Sprintf("FilePath: %s", df.FilePath()))
	lw.AppendItem(fmt.Sprintf("FileFormat: %v", df.FileFormat()))
	if verbose {
		lw.AppendItem(fmt.Sprintf("Partition: %v", df.Partition()))
	}
	lw.AppendItem(fmt.Sprintf("Count: %d", df.Count()))
	lw.AppendItem(fmt.Sprintf("FileSizeBytes: %d", df.FileSizeBytes()))

	if verbose {
		lw.AppendItem(fmt.Sprintf("ColumnSizes: %v", df.ColumnSizes()))
		lw.AppendItem(fmt.Sprintf("ValueCounts: %v", df.ValueCounts()))
		lw.AppendItem(fmt.Sprintf("NullValueCounts: %v", df.NullValueCounts()))
		lw.AppendItem(fmt.Sprintf("NaNValueCounts: %v", df.NaNValueCounts()))
		lw.AppendItem(fmt.Sprintf("DistinctValueCounts: %v", df.DistinctValueCounts()))
		lw.AppendItem(fmt.Sprintf("LowerBoundValues: %v", df.LowerBoundValues()))
		lw.AppendItem(fmt.Sprintf("UpperBoundValues: %v", df.UpperBoundValues()))
	}

	lw.UnIndent()
}
