package id

/*
github.com/twitter/snowflake in golang

id =>  timestamp retain center worker sequence
          40      4       5      5      10

*/

import (
	"github.com/Kotodian/gokit/datasource"
	"github.com/yitter/idgenerator-go/idgen"
)

// var partnerId uint32

func Init() {
	// 创建 IdGeneratorOptions 对象，请在构造函数中输入 WorkerId：
	var options = idgen.NewIdGeneratorOptions(1)
	// options.WorkerIdBitLength = 10 // WorkerIdBitLength 默认值6，支持的 WorkerId 最大值为2^6-1，若 WorkerId 超过64，可设置更大的 WorkerIdBitLength
	// ...... 其它参数设置参考 IdGeneratorOptions 定义，一般来说，只要再设置 WorkerIdBitLength （决定 WorkerId 的最大值）。

	// 保存参数（必须的操作，否则以上设置都不能生效）：
	idgen.SetIdGenerator(options)
	// 以上初始化过程只需全局一次，且必须在第2步之前设置。
}

func Next() datasource.UUID {
	newID := idgen.NextId()
	return datasource.UUID(newID)
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
