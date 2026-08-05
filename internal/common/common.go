// Package common provides trivial utilitiy functions for s3tdl.
package common

import (
	"path/filepath"
	"strings"

	"github.com/apache/iceberg-go/table"
)

func ID2Str(id table.Identifier) string {
	return strings.Join(id, ".")
}

func ID2Path(id table.Identifier) string {
	return filepath.Join(id...)
}
