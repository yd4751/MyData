package snowflake

import (
	"fmt"
	"sync"
	"time"
)

const (
	workerIDBits   = 10
	sequenceBits   = 12
	workerIDShift  = sequenceBits
	timestampShift = workerIDBits + sequenceBits
	maxWorkerID    = -1 ^ (-1 << workerIDBits)
	maxSequence    = -1 ^ (-1 << sequenceBits)
)

var (
	epoch         = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC).UnixNano() / 1e6
	mu            sync.Mutex
	workerID      int64
	lastTimeStamp int64
	sequence      int64
)

func Init(worker int64) error {
	if worker > maxWorkerID || worker < 0 {
		return ErrInvalidWorkerID
	}
	workerID = worker
	return nil
}

func Generate() int64 {
	mu.Lock()
	defer mu.Unlock()

	now := time.Now().UnixNano() / 1e6
	if now < lastTimeStamp {
		panic("clock moved backwards")
	}

	if now == lastTimeStamp {
		sequence = (sequence + 1) & maxSequence
		if sequence == 0 {
			for now <= lastTimeStamp {
				now = time.Now().UnixNano() / 1e6
			}
		}
	} else {
		sequence = 0
	}

	lastTimeStamp = now
	return (now-epoch)<<timestampShift | workerID<<workerIDShift | sequence
}

func GenerateID() int64 {
	return Generate()
}

func GenerateIDString() string {
	return fmt.Sprintf("%d", Generate())
}

var ErrInvalidWorkerID = &snowflakeError{"invalid worker ID"}

type snowflakeError struct {
	msg string
}

func (e *snowflakeError) Error() string {
	return e.msg
}
