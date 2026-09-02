package common

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/apache/iceberg-go/catalog/rest"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Catalogs map[string]CatalogConfig `yaml:"catalog"`
}

type AWSOptions struct {
	Profile    string `yaml:"profile"`
	Region     string `yaml:"region"`
	S3Endpoint string `yaml:"s3-endpoint"`
}

type RestOptions struct {
	SigV4Enabled  bool   `yaml:"sigv4-enabled"`
	SigningName   string `yaml:"signing-name"`
	SigningRegion string `yaml:"signing-region"`
}

type CatalogConfig struct {
	URI         string       `yaml:"uri"`
	Warehouse   string       `yaml:"warehouse"`
	Token       string       `yaml:"token"`
	Credential  string       `yaml:"credential"`
	Scope       string       `yaml:"scope"`
	AWS         *AWSOptions  `yaml:"aws,omitempty"`
	RestOptions *RestOptions `yaml:"rest,omitempty"`
}

func (cc *CatalogConfig) S3Endpoint() string {
	if cc.AWS == nil {
		return ""
	}
	return cc.AWS.S3Endpoint
}

func LoadCatalogConfig(configPath string, catalogName string) (*CatalogConfig, error) {
	b, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	var config Config
	if err := yaml.Unmarshal(b, &config); err != nil {
		return nil, err
	}
	if catalogName == "" {
		catalogName = "default"
	}
	cc, ok := config.Catalogs[catalogName]
	if !ok {
		return nil, fmt.Errorf("no catalogs with name: %s", catalogName)
	}
	return &cc, nil
}

func NewCatalogFromConfig(ctx context.Context, configPath string, catalogName string) (*Catalog, error) {
	cfg, err := LoadCatalogConfig(configPath, catalogName)
	if err != nil {
		return nil, err
	}

	var opts []rest.Option

	if cfg.Warehouse != "" {
		opts = append(opts, rest.WithWarehouseLocation(cfg.Warehouse))
	}
	if cfg.Token != "" {
		opts = append(opts, rest.WithOAuthToken(cfg.Token))
	} else if cfg.Credential != "" {
		opts = append(opts, rest.WithCredential(cfg.Credential))
	}
	if cfg.Scope != "" {
		opts = append(opts, rest.WithScope(cfg.Scope))
	}

	if ro := cfg.RestOptions; ro != nil {
		var ro = cfg.RestOptions
		if ro.SigV4Enabled {
			opts = append(opts, rest.WithSigV4())
		}
		if ro.SigningName != "" || ro.SigningRegion != "" {
			opts = append(opts, rest.WithSigV4RegionSvc(ro.SigningRegion, ro.SigningName))
		}
	}

	var awscfg aws.Config
	if ao := cfg.AWS; ao != nil {
		var awsopts []func(*awsconfig.LoadOptions) error
		if ao.Profile != "" {
			awsopts = append(awsopts, awsconfig.WithSharedConfigProfile(ao.Profile))
		}
		if ao.Region != "" {
			awsopts = append(awsopts, awsconfig.WithRegion(ao.Region))
		}
		awscfg, err = awsconfig.LoadDefaultConfig(ctx, awsopts...)
		if err != nil {
			return nil, err
		}
		opts = append(opts, rest.WithAwsConfig(awscfg))
	}

	cat, err := rest.NewCatalog(ctx, catalogName, cfg.URI, opts...)
	if err != nil {
		return nil, err
	}

	return &Catalog{
		Catalog:   cat,
		Config:    cfg,
		AWSConfig: awscfg,
	}, nil
}

var (
	ConfigPath  = "tmp/config.yaml"
	CatalogName = "default"
)

func InitFlagSet(fs *flag.FlagSet) {
	fs.StringVar(&ConfigPath, "config", ConfigPath, "location of config YAML")
	fs.StringVar(&CatalogName, "catalog", CatalogName, "catalog name to use")
}

func DefaultCatalog(ctx context.Context) (*Catalog, error) {
	return NewCatalogFromConfig(ctx, ConfigPath, CatalogName)
}
