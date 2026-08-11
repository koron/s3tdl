// Package inspect provides S3Tables (Iceberg catalog) inspection.
package inspect

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/koron-go/subcmd"
	"github.com/koron/s3tdl/internal/common"
)

var Inspect = subcmd.DefineCommand("inspect", "inspect S3 Tables (Iceberg catalog)", inspectCommand)

func inspectCommand(ctx context.Context, args []string) error {
	fs := subcmd.FlagSet(ctx)
	fs.Parse(args)
	arns := fs.Args()
	if len(arns) == 0 {
		return errors.New("please specify one or more ARNs")
	}
	for _, arn := range arns {
		err := inspectCatalog(ctx, arn)
		if err != nil {
			return err
		}
	}
	return nil
}

func inspectCatalog(ctx context.Context, arn string) error {
	cat, err := common.NewCatalog(ctx, arn)
	if err != nil {
		return err
	}

	namespaces, err := cat.ListNamespaces(ctx, nil)
	if err != nil {
		return err
	}

	for _, ns := range namespaces {
		fmt.Printf("Namespace: %s\n", common.ID2Str(ns))
		props, err := cat.LoadNamespaceProperties(ctx, ns)
		if err != nil {
			log.Printf("failed to LoadNamespaceProperties: %s", err)
		}
		fmt.Printf("  Properties (%d)\n", len(props))
		for k, v := range props {
			fmt.Printf("    %s: s\n", k, v)
		}

		for tableID, err := range cat.ListTables(ctx, ns) {
			if err != nil {
				log.Printf("failed to ListTables: %s", err)
				break
			}
			fmt.Printf("Table: %s\n", common.ID2Str(tableID))
			table, err := cat.LoadTable(ctx, tableID)
			if err != nil {
				log.Printf("failed to LoadTable: %s", err)
				continue
			}

			snapshot := table.CurrentSnapshot()
			fmt.Printf("  Current Snapshot: %d\n", snapshot.SnapshotID)
			fmt.Printf("    %+v\n", *snapshot)

			schema := table.Schema()
			fmt.Println("  Schema:")
			for _, f := range schema.Fields() {
				fmt.Printf("    %s\n", f.String())
			}

			partition := table.Spec()
			fmt.Printf("  Partition Spec: %+v\n", partition)
		}
	}

	return nil
}
