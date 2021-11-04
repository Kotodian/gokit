package redis

import (
	rdlib "github.com/gomodule/redigo/redis"
	"os"
	"strconv"
	"sync"
	"time"
)

var _workerIdLock sync.Mutex

var _workerIdList []int32 // 当前已注册的WorkerId
var _loopCount = 0        // 循环数量
var _lifeIndex = -1       // WorkerId本地生命时序（本地多次注册时，生命时序会不同）

var _WorkerIdLifeTimeSeconds = 42    // IdGen:WorkerId:Value:xx 的值在 redis 中的有效期（单位秒，最好是3的整数倍）
var _MaxLoopCount = 10               // 最大循环次数（无可用WorkerId时循环查找）
var _SleepMillisecondEveryLoop = 200 // 每次循环后，暂停时间
var _MaxWorkerId int32 = 0           // 最大WorkerId值，超过此值从0开始

const _WorkerIdIndexKey string = "IdGen:WorkerId:Index"        // redis 中的key
const _WorkerIdValueKeyPrefix string = "IdGen:WorkerId:Value:" // redis 中的key
const _WorkerIdFlag = "Y"                                      // IdGen:WorkerId:Value:xx 的值（将来可用 _token 替代）

func Validate(workerId int32) int32 {
	for _, value := range _workerIdList {
		if value == workerId {
			return 1
		}
	}

	return 0

}

func UnRegister() {
	_workerIdLock.Lock()

	_lifeIndex = -1
	for _, value := range _workerIdList {
		if value > -1 {
			_client := GetRedis()
			_client.Do("del", _WorkerIdValueKeyPrefix+strconv.Itoa(int(value)))
		}
	}
	_workerIdList = []int32{}

	_workerIdLock.Unlock()
}

func autoUnRegister() {
	// 如果当前已注册过 WorkerId，则先注销，并终止先前的自动续期线程
	if len(_workerIdList) > 0 {
		UnRegister()
	}
}

func RegisterOne(maxWorkerId int32) (int32, error) {
	if maxWorkerId < 0 {
		return -2, nil
	}

	autoUnRegister()

	_MaxWorkerId = maxWorkerId
	_loopCount = 0
	_client := GetRedis()
	defer func() {
		if _client != nil {
			_ = _client.Close()
		}
	}()

	_lifeIndex++
	var id, err = register(_client, _lifeIndex)
	if id > -1 {
		_workerIdList = []int32{id}
		go extendLifeTime(_lifeIndex)
	}

	return id, err
}

func register(_client rdlib.Conn, lifeTime int) (int32, error) {
	_loopCount = 0
	return getNextWorkerId(_client, lifeTime)
}

func getNextWorkerId(_client rdlib.Conn, lifeTime int) (int32, error) {
	// 获取当前 WorkerIdIndex
	r, err := rdlib.Int(_client.Do("incr", _WorkerIdIndexKey))
	if err != nil {
		return -1, err
	}

	candidateId := int32(r)

	// 如果 candidateId 大于最大值，则重置
	if candidateId > _MaxWorkerId {
		if canReset(_client) {
			// 当前应用获得重置 WorkerIdIndex 的权限
			err = setWorkerIdIndex(_client, 0)
			if err != nil {
				return -1, err
			}
			err = endReset(_client) // 此步有可能不被执行？
			if err != nil {
				return -1, err
			}
			_loopCount++

			// 超过一定次数，直接终止操作
			if _loopCount > _MaxLoopCount {
				_loopCount = 0
				// 返回错误
				return -1, err
			}

			// 每次一个大循环后，暂停一些时间
			time.Sleep(time.Duration(_SleepMillisecondEveryLoop*_loopCount) * time.Millisecond)

			return getNextWorkerId(_client, lifeTime)
		} else {
			// 如果有其它应用正在编辑，则本应用暂停200ms后，再继续
			time.Sleep(time.Duration(200) * time.Millisecond)

			return getNextWorkerId(_client, lifeTime)
		}
	}

	if isAvailable(_client, candidateId) {

		// 最新获得的 WorkerIdIndex，在 redis 中是可用状态
		err = setWorkerIdFlag(_client, candidateId)
		if err != nil {
			return -1, err
		}
		_loopCount = 0

		// 获取到可用 WorkerId 后，启用新线程，每隔 1/3个 _WorkerIdLifeTimeSeconds 时间，向服务器续期（延长一次 LifeTime）
		// go extendWorkerIdLifeTime(lifeTime, candidateId)

		return candidateId, nil
	} else {
		// 最新获得的 WorkerIdIndex，在 redis 中是不可用状态，则继续下一个 WorkerIdIndex
		return getNextWorkerId(_client, lifeTime)
	}
}

func extendLifeTime(lifeIndex int) {
	// 获取到可用 WorkerId 后，启用新线程，每隔 1/3个 _WorkerIdLifeTimeSeconds 时间，向服务器续期（延长一次 LifeTime）
	var myLifeIndex = lifeIndex

	// 循环操作：间隔一定时间，刷新 WorkerId 在 redis 中的有效时间。
	for {
		time.Sleep(time.Duration(_WorkerIdLifeTimeSeconds/3) * time.Second)

		// 上锁操作，防止跟 UnRegister 操作重叠
		_workerIdLock.Lock()

		// 如果临时变量 myLifeIndex 不等于 全局变量 _lifeIndex，表明全局状态被修改，当前线程可终止，不应继续操作 redis
		if myLifeIndex != _lifeIndex {
			break
		}

		// 已经被注销，则终止（此步是上一步的二次验证）
		if len(_workerIdList) < 1 {
			break
		}

		// 延长 redis 数据有效期
		for _, value := range _workerIdList {
			if value > -1 {
				extendWorkerIdFlag(value)
			}
		}

		_workerIdLock.Unlock()
	}
}

func extendWorkerIdLifeTime(lifeIndex int, workerId int32) {
	var myLifeIndex = lifeIndex
	var myWorkerId = workerId

	// 循环操作：间隔一定时间，刷新 WorkerId 在 redis 中的有效时间。
	for {
		time.Sleep(time.Duration(_WorkerIdLifeTimeSeconds/3) * time.Second)

		// 上锁操作，防止跟 UnRegister 操作重叠
		_workerIdLock.Lock()

		// 如果临时变量 myLifeIndex 不等于 全局变量 _lifeIndex，表明全局状态被修改，当前线程可终止，不应继续操作 redis
		if myLifeIndex != _lifeIndex {
			break
		}

		// 已经被注销，则终止（此步是上一步的二次验证）
		//if _usingWorkerId < 0 {
		//	break
		//}

		// 延长 redis 数据有效期
		extendWorkerIdFlag(myWorkerId)

		_workerIdLock.Unlock()
	}
}

func setWorkerIdIndex(_client rdlib.Conn, val int) error {
	_, err := _client.Do("set", _WorkerIdIndexKey, val)
	return err
}

func setWorkerIdFlag(_client rdlib.Conn, workerId int32) error {
	hostname, _ := os.Hostname()
	_, err := _client.Do("set", _WorkerIdValueKeyPrefix+strconv.Itoa(int(workerId)), hostname, "EX", "15")
	return err
}

func extendWorkerIdFlag(workerId int32) {
	var client = GetRedis()
	if client == nil {
		return
	}
	defer func() {
		if client != nil {
			_ = client.Close()
		}
	}()

	client.Do("expire", _WorkerIdValueKeyPrefix+strconv.Itoa(int(workerId)), time.Duration(_WorkerIdLifeTimeSeconds)*time.Second)
}

func canReset(_client rdlib.Conn) bool {
	r, err := rdlib.Int(_client.Do("incr", _WorkerIdValueKeyPrefix+"Edit"))
	if err != nil && err != rdlib.ErrNil {
		return false
	}

	return r != 1
}

func endReset(_client rdlib.Conn) error {
	//_client.Set(_WorkerIdValueKeyPrefix+"Edit", 0, time.Duration(2)*time.Second)
	_, err := _client.Do("set", _WorkerIdValueKeyPrefix+"Edit", 0, "EX", "2")
	return err
}

func isAvailable(_client rdlib.Conn, workerId int32) bool {
	r, err := rdlib.String(_client.Do("get", _WorkerIdValueKeyPrefix+strconv.Itoa(int(workerId))))
	if err != nil {
		if err == rdlib.ErrNil {
			return true
		}
		return false
	}

	return r != _WorkerIdFlag
}
