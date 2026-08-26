// Package common provides trivial utilitiy functions for s3tdl.
package common

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/apache/iceberg-go/catalog/rest"
	"github.com/apache/iceberg-go/table"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
)

func ID2Str(id table.Identifier) string {
	return strings.Join(id, ".")
}

func ID2Path(id table.Identifier) string {
	return filepath.Join(id...)
}

type ARN struct {
	Whole   string
	Service string
	Region  string
}

// ParseARN parses a string as the ARN, and extract the service name and region.
func ParseARN(s string) (ARN, error) {
	parts := strings.SplitN(s, ":", 5)
	if len(parts) < 4 {
		return ARN{}, fmt.Errorf("lack elements for ARN: want=4 got=%d", len(parts))
	}
	return ARN{
		Whole:   s,
		Service: parts[2],
		Region:  parts[3],
	}, nil
}

type Catalog struct {
	*rest.Catalog

	Config    *CatalogConfig
	ARN       ARN
	AwsConfig aws.Config
}

func NewCatalog(ctx context.Context, arnStr string) (Catalog, error) {
	arn, err := ParseARN(arnStr)
	if err != nil {
		return Catalog{}, err
	}

	// Load AWS default configuration.
	cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(arn.Region))
	if err != nil {
		return Catalog{}, fmt.Errorf("failed to load AWS default config: %w", err)
	}

	// Create a catalog to access S3 Tables.
	cat, err := rest.NewCatalog(
		ctx,
		"Arbitrary catalog name",
		fmt.Sprintf("https://s3tables.%s.amazonaws.com/iceberg", arn.Region),
		rest.WithWarehouseLocation(arn.Whole),
		rest.WithSigV4RegionSvc(arn.Region, arn.Service),
		rest.WithAwsConfig(cfg),
	)
	if err != nil {
		return Catalog{}, fmt.Errorf("failed to create REST catalog: %w", err)
	}

	return Catalog{
		Catalog:   cat,
		ARN:       arn,
		AwsConfig: cfg,
	}, nil
}
