package archiver

import (
	"strings"
	"time"
)

func getShardId(instance string) string {
	return strings.SplitN(instance, ".", 2)[0]
}

func unixMicro(tm time.Time) uint64 {
	return uint64(tm.UnixNano()) / 1000
}
