package config

import "time"

const (
	NUM_VIRTUAL_NODES      int           = 3
	REPLICATION_FACTOR     int           = 3                      // N value
	RING_BACKUP_FILEFORMAT string        = "store/node%dring.bak" // %d format is compulsory
	RING_BACKUP_INTERVAL   time.Duration = 1 * time.Minute
	HINT_CHECK_INTERVAL    time.Duration = 5 * time.Minute
	HINT_BUCKETNAME        string        = "hints"
)
