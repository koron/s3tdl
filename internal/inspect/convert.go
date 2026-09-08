package inspect

import (
	"fmt"

	"github.com/apache/iceberg-go"
)

func ManifestEntryStatus2String(status iceberg.ManifestEntryStatus) string {
	switch status {
	case iceberg.EntryStatusEXISTING:
		return "0:EXISTING"
	case iceberg.EntryStatusADDED:
		return "1:ADDED"
	case iceberg.EntryStatusDELETED:
		return "2:DELETED"
	default:
		return fmt.Sprintf("%d:UNKNOWN", status)
	}
}
