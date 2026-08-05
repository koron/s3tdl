// Package download provides table data file download sub command.
package download

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/apache/iceberg-go/catalog/rest"
	"github.com/apache/iceberg-go/table"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/koron-go/subcmd"
	"github.com/koron/s3tdl/internal/common"

	// Necessary for S3(s3://) scheme
	_ "github.com/apache/iceberg-go/io/gocloud"
)

var (
	optARN     string
	optOutdir  string
	optDryrun  bool
	optVerbose bool

	optNS    string
	optTable string
)

func downloadDataFiles(ctx context.Context, args []string) error {
	fs := subcmd.FlagSet(ctx)
	fs.StringVar(&optARN, "arn", "", `ARN for S3 Tables bucket`)
	fs.StringVar(&optOutdir, "outdir", ".", `Output dir for downloaded data files`)
	fs.BoolVar(&optDryrun, "dryrun", false, `Dryrun, not actually download`)
	fs.BoolVar(&optVerbose, "verbose", false, `Show verbose message`)
	fs.StringVar(&optNS, "namespace", "", `Namespace filter regexp`)
	fs.StringVar(&optTable, "table", "", `Table filter regexp`)
	fs.Parse(args)
	return run(ctx)
}

var DownloadCommand = subcmd.DefineCommand("download", "download data files", downloadDataFiles)

var rxNS = sync.OnceValue(func() *regexp.Regexp {
	return regexp.MustCompile(optNS)
})

func matchNamespaceFilter(id table.Identifier) bool {
	if optNS == "" {
		return true
	}
	return rxNS().MatchString(common.ID2Str(id))
}

var rxTable = sync.OnceValue(func() *regexp.Regexp {
	return regexp.MustCompile(optTable)
})

func matchTableFilter(id table.Identifier) bool {
	if optTable == "" {
		return true
	}
	return rxTable().MatchString(common.ID2Str(id))
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

	// Prepare S3 client to access the head of data file.
	client := s3.NewFromConfig(cfg)

	// Retrieve data files for all tables from each namespace.
	for _, ns := range namespaces {
		if !matchNamespaceFilter(ns) {
			log.Printf("Skip namespace: %s", common.ID2Str(ns))
			continue
		}
		if optVerbose {
			log.Printf("Namespace: %s", common.ID2Str(ns))
		}
		for tbl, err := range cat.ListTables(ctx, ns) {
			if err != nil {
				return err
			}
			if !matchTableFilter(tbl) {
				log.Printf("Skip table: %s", common.ID2Str(tbl))
				continue
			}
			if optVerbose {
				log.Printf("Table: %s", common.ID2Str(tbl))
			}
			table, err := cat.LoadTable(ctx, tbl)
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
				outdir := filepath.Join(optOutdir, common.ID2Path(tbl))
				_, err := getBody(ctx, client, i, path, outdir)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func getBody(ctx context.Context, client *s3.Client, n int, s3url, outdir string) (string, error) {
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

	if optVerbose {
		log.Printf("getBody#%d:\n    bucket=%s\n    key=%s\n    name=%s\n    localName=%s", n, bucket, key, name, localName)
	}
	if optDryrun {
		log.Printf("Dryrun: skip download: %s", localName)
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
