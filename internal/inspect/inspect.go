// Package inspect provides S3Tables (Iceberg catalog) inspection.
package inspect

import (
	"context"
	"errors"

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
	// TODO:
	println("Hello inspect")
	_ = cat
	return nil
}
