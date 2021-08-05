/*
github.com/twitter/snowflake in golang

id =>  timestamp sequence type  pos   retained
		  30        10	   4     10       10

timestamp   时间戳，从2018年开始
sequence    序列号
type        类型  0 设备 1 枪头 2 地锁 3 车位传感器
pos   		位置
device


*/

package id

import (
	"fmt"
	"sync"
	"time"
)

const (
	BitsTimestamp = 30
	BitsSequence  = 10
	BitsType      = 4
	BitsPos       = 10
	BitsRetain    = 10
	MaxSequence   = 1<<BitsSequence - 1
	MaxType       = 1<<BitsType - 1
	MaxPos        = 1<<BitsPos - 1
	MaxRetain     = 1<<BitsRetain - 1
	MaskType      = MaxType << (BitsPos + BitsRetain)
	MaskPos       = MaxPos << BitsRetain
	MaskRetain    = MaxRetain
)

const (
	KindTypeEvse = iota
	KindTypeConnector
	KindTypeParkLock
	KindTypeParkSensor
	KindTypeGroup = 15
)

var (
	Since     int64 = time.Date(2018, 3, 29, 0, 0, 0, 0, time.Local).Unix()
	snowFlake *SnowFlake
)

func init() {
	snowFlake = NewSnowFlake()
}

type SnowFlake struct {
	lastTimestamp uint64
	sequence      uint32
	lock          sync.Mutex
}

func (sf *SnowFlake) uint64() uint64 {
	return (sf.lastTimestamp << (BitsSequence + BitsType + BitsPos + BitsRetain)) |
		(uint64(sf.sequence) << (BitsType + BitsPos + BitsRetain))
}

func (sf *SnowFlake) next() (uint64, error) {
	sf.lock.Lock()
	defer sf.lock.Unlock()

	ts := timestamp()
	if ts == sf.lastTimestamp {
		sf.sequence = (sf.sequence + 1) & MaxSequence
		if sf.sequence == 0 {
			ts = tilNextSecond(ts)
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

func NewSnowFlake() *SnowFlake {
	return &SnowFlake{}
}

/**
 * seconds from 2016-01-01
 */
func timestamp() uint64 {
	return uint64(time.Now().Unix() - Since)
}

func tilNextSecond(ts uint64) uint64 {
	i := timestamp()
	for i <= ts {
		i = timestamp()
	}
	return i
}

func NewEvseID() (uint64, error) {
	id, err := snowFlake.next()
	if err != nil {
		return 0, err
	}
	return id, nil
}

func GetPosID(evseID uint64, kindType int, pos int32) (uint64, error) {
	if pos > MaxPos {
		//TODO: 要警告，需要注意超过范围
		return 0, fmt.Errorf("group pos %v is invalid", pos)
	}
	return (evseID | uint64(kindType)<<(BitsPos+BitsRetain) | uint64(pos)<<(BitsRetain)), nil
}

func GetPosByID(hardwareID uint64) uint64 {
	return hardwareID & MaskPos >> BitsRetain
}

func GetTypeByID(hardwareID uint64) uint64 {
	return hardwareID & MaskType >> (BitsPos + BitsRetain)
}
