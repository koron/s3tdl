package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/apache/iceberg-go/catalog/rest"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	// Necessary for S3(s3://) scheme
	_ "github.com/apache/iceberg-go/io/gocloud"
)

var (
	optARN    string
	optOutdir string
	optDryrun bool
)

func main() {
	flag.StringVar(&optARN, "arn", "", `ARN for S3 Tables bucket`)
	flag.StringVar(&optOutdir, "outdir", ".", `Output dir for downloaded data files`)
	flag.BoolVar(&optDryrun, "dryrun", false, `Dryrun, not actually download`)
	flag.Parse()

	if err := run(context.Background()); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context) error {
	// Parse the ARN to extract the service name and region.
	parts := strings.SplitN(optARN, ":", 5)
	if len(parts) < 4 {
		return fmt.Errorf("too short ARN: want=4 got=%d", len(parts))
	}
	service, region := parts[2], parts[3]

	// Load AWS default configuration.
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return err
	}

	// Create a catalog to access S3 Tables.
	cat, err := rest.NewCatalog(
		ctx,
		"Arbitrary catalog name",
		fmt.Sprintf("https://s3tables.%s.amazonaws.com/iceberg", region),
		rest.WithWarehouseLocation(optARN),
		rest.WithSigV4RegionSvc(region, service),
		rest.WithAwsConfig(cfg),
	)
	if err != nil {
		return err
	}

	// List namespaces
	namespaces, err := cat.ListNamespaces(ctx, nil)
	if err != nil {
		return err
	}
	fmt.Printf("Namespaces: %+v\n", namespaces)

	// Prepare S3 client to access the head of data file.
	client := s3.NewFromConfig(cfg)

	// Retrieve data files for all tables from each namespace.
	for _, ns := range namespaces {
		for id, err := range cat.ListTables(ctx, ns) {
			if err != nil {
				return err
			}
			fmt.Println()
			fmt.Printf("Table: %s\n", id)
			table, err := cat.LoadTable(ctx, id)
			if err != nil {
				return err
			}
			scan := table.Scan()
			tasks, err := scan.PlanFiles(ctx)
			if err != nil {
				return err
			}
			for i, task := range tasks {
				// Retrieve the header of a data file from a path using the AWS S3 SDK.
				path := task.File.FilePath()
				fmt.Printf("#%d FilePath=%s\n", i, path)
				_, err := getBody(ctx, client, path, optOutdir)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func getBody(ctx context.Context, client *s3.Client, s3url, outdir string) (string, error) {
	// Parse S3 URL
	u, err := url.Parse(s3url)
	if err != nil {
		return "", err
	}
	bucket, key := u.Host, strings.TrimLeft(u.Path, "/")
	name := path.Base(key)
	if name == "." || name == "/" {
		return "", fmt.Errorf("invalid name: %q", name)
	}
	localName := filepath.Join(outdir, name)

	fmt.Printf("getBody:\n    bucket=%s\n    key=%s\n    name=%s\n    localName=%s\n", bucket, key, name, localName)
	if optDryrun {
		return localName, nil
	}

	out, err := client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return "", err
	}
	defer out.Body.Close()

	if err := os.MkdirAll(outdir, 0755); err != nil {
		return "", err
	}
	localFile, err := os.Create(localName)
	if err != nil {
		return "", err
	}
	defer localFile.Close()

	if _, err := io.Copy(localFile, out.Body); err != nil {
		return "", err
	}

	return localName, nil
}
