package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

var REPLICATION_FACTOR int // Replication factor
var	MIN_WRITES int // Minimum no. of nodes that participate in a successful write operation
var MIN_READS int // Minimum no. of nodes that participate in a successful read operation
var	NUM_VIRTUAL_NODES int // No. of virtual nodes per physical node

func LoadEnvFile() error  {
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

	return nil
}