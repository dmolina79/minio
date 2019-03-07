package cmd

import (
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"time"

	"github.com/minio/minio/cmd/logger"
)

type KVInterface interface {
	Put(keyStr string, value []byte) error
	Get(keyStr string, value []byte) ([]byte, error)
	Delete(keyStr string) error
}

const kvNSEntryPaddingMultiple = 4 * 1024

type KVNSEntry struct {
	Key     string
	Size    int64
	ModTime time.Time
	IDs     []string
}

func KVNSEntryMarshal(entry KVNSEntry) ([]byte, error) {
	b, err := json.Marshal(entry)
	if err != nil {
		return nil, err
	}
	padded := make([]byte, ceilFrac(int64(len(b)), kvNSEntryPaddingMultiple)*kvNSEntryPaddingMultiple)
	copy(padded, b)
	return padded, nil
}

func KVNSEntryUnmarshal(b []byte, entry *KVNSEntry) error {
	for i := range b {
		if b[i] == '\x00' {
			b = b[:i]
			break
		}
	}
	return json.Unmarshal(b, entry)
}

var errValueTooLong = errors.New("value too long")

const kvDataDir = ".minio.sys/.data"

var kvMaxValueSize = getKVMaxValueSize()

func getKVMaxValueSize() int {
	str := os.Getenv("MINIO_NKV_MAX_VALUE_SIZE")
	if str == "" {
		return 2 * 1024 * 1024
	}
	valSize, err := strconv.Atoi(str)
	logger.FatalIf(err, "parsing MINIO_NKV_MAX_VALUE_SIZE")
	return valSize
}
