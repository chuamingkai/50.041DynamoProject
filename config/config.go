package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

var REPLICATION_FACTOR int             // Replication factor
var MIN_WRITES int                     // Minimum no. of nodes that participate in a successful write operation
var MIN_READS int                      // Minimum no. of nodes that participate in a successful read operation
var NUM_VIRTUAL_NODES int              // No. of virtual nodes per physical node
var RING_BACKUP_FILEFORMAT string      // File format of the ring backup file. %d format is compulsory
var RING_BACKUP_INTERVAL time.Duration // Time interval in seconds that ring backup is performed periodically
var HINT_CHECK_INTERVAL time.Duration  // Time interval in seconds that hint bucket is checked periodically
var HINT_BUCKETNAME string             // Reserved name for hint bucket
var MAIN_BUCKETNAME string

func LoadEnvFile() error {
	var err error
	err = godotenv.Load()
	if err != nil {
		return err
	}

	if REPLICATION_FACTOR, err = strconv.Atoi(os.Getenv("REPLICATION_FACTOR")); err != nil {
		return err
	}

	if MIN_WRITES, err = strconv.Atoi(os.Getenv("MIN_WRITES")); err != nil {
		return err
	}

	if MIN_READS, err = strconv.Atoi(os.Getenv("MIN_READS")); err != nil {
		return err
	}

	if NUM_VIRTUAL_NODES, err = strconv.Atoi(os.Getenv("NUM_VIRTUAL_NODES")); err != nil {
		return err
	}

	RING_BACKUP_FILEFORMAT = os.Getenv("RING_BACKUP_FILEFORMAT")

	intermediate_time := 1

	if intermediate_time, err = strconv.Atoi(os.Getenv("RING_BACKUP_INTERVAL")); err != nil {
		return err
	}

	RING_BACKUP_INTERVAL = time.Duration(intermediate_time) * time.Second

	if intermediate_time, err = strconv.Atoi(os.Getenv("HINT_CHECK_INTERVAL")); err != nil {
		return err
	}

	HINT_CHECK_INTERVAL = time.Duration(intermediate_time) * time.Second

	HINT_BUCKETNAME = os.Getenv("HINT_BUCKETNAME")

	MAIN_BUCKETNAME = os.Getenv("MAIN_BUCKETNAME")

	return nil
}
