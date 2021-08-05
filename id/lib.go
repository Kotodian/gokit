package id

/*
github.com/twitter/snowflake in golang

id =>  timestamp retain center worker sequence
          40      4       5      5      10

*/

import (
	"fmt"
	"sync"
	"time"
)

// var partnerId uint32
var IDGenr *SnowFlake

func init() {
	IDGenr, _ = NewSnowFlake(0, 0)
}

const (
	nano = 1000 * 1000
)

const (
	TimestampBits = 40                         // timestamp
	Maxtimestamp  = -1 ^ (-1 << TimestampBits) // timestamp mask
	RetainedBits  = 4
	MaxRetain     = -1 ^ (-1 << RetainedBits)
	CenterBits    = 5
	MaxCenter     = -1 ^ (-1 << CenterBits) // center mask
	WorkerBits    = 5
	MaxWorker     = -1 ^ (-1 << WorkerBits)   // worker mask
	SequenceBits  = 10                        // sequence
	MaxSequence   = -1 ^ (-1 << SequenceBits) // sequence mask
)

var (
	Since  int64                 = time.Date(2017, 5, 1, 0, 0, 0, 0, time.Local).UnixNano() / nano
	poolMu sync.RWMutex          = sync.RWMutex{}
	pool   map[uint64]*SnowFlake = make(map[uint64]*SnowFlake)
)

type SnowFlake struct {
	lastTimestamp uint64
	retain        uint32
	center        uint32
	worker        uint32
	sequence      uint32
	lock          sync.Mutex
}

func (sf *SnowFlake) uint64() uint64 {
	return (sf.lastTimestamp << (RetainedBits + CenterBits + WorkerBits + SequenceBits)) |
		(uint64(sf.retain) << (CenterBits + WorkerBits + SequenceBits)) |
		(uint64(sf.center) << (WorkerBits + SequenceBits)) |
		(uint64(sf.worker) << SequenceBits) |
		uint64(sf.sequence)
	// (uint64(sf.partnerId) << BusinessBits) |
	// (uint64(sf.businessId))
}

func (sf *SnowFlake) Next() (uint64, error) {
	sf.lock.Lock()
	defer sf.lock.Unlock()

	ts := timestamp()
	if ts == sf.lastTimestamp {
		sf.sequence = (sf.sequence + 1) & MaxSequence
		if sf.sequence == 0 {
			ts = tilNextMillis(ts)
		}
	} else {
		sf.sequence = 0
	}

	if ts < sf.lastTimestamp {
		return 0, fmt.Errorf("Invalid timestamp: %v - precedes %v", ts, sf)
	}
	sf.lastTimestamp = ts
	return sf.uint64(), nil
}

func NewSnowFlake(centerID uint32, workerID uint32) (*SnowFlake, error) {
	if centerID > MaxCenter {
		return nil, fmt.Errorf("CenterID %v is invalid", centerID)
	} else if workerID > MaxWorker {
		return nil, fmt.Errorf("WorkerID %v is invalid", workerID)
	}
	return &SnowFlake{
		worker: workerID,
		center: centerID,
	}, nil
}

func timestamp() uint64 {
	return uint64(time.Now().UnixNano()/nano - Since)
}

func tilNextMillis(ts uint64) uint64 {
	i := timestamp()
	for i <= ts {
		i = timestamp()
	}
	return i
}

// func GetSnowFlake(businessId uint32) (*SnowFlake, error) {
// 	var key uint64 = uint64(businessId) << SequenceBits
// 	var sf *SnowFlake
// 	var exist bool
// 	var err error
// 	poolMu.RLock()
// 	if sf, exist = pool[key]; !exist {
// 		poolMu.RUnlock()
// 		poolMu.Lock()
// 		// double check
// 		if sf, exist = pool[key]; !exist {
// 			sf, err = NewSnowFlake(businessId)
// 			if err != nil {
// 				poolMu.Unlock()
// 				return nil, err
// 			}
// 			pool[key] = sf
// 		}
// 		poolMu.Unlock()
// 	} else {
// 		poolMu.RUnlock()
// 	}
// 	return sf, err
// }
