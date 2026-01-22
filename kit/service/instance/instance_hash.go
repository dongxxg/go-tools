package instance

import (
	"encoding/json"
	"github.com/gogf/gf/v2/container/gset"
	"github.com/pkg/errors"
	"hash/crc32"
	"io"
	"sort"
	"strconv"
	"strings"
	"time"
	"unitechs.com/cno/kit/db/kredis"
	"unitechs.com/cno/kit/kctx"
	"unitechs.com/cno/kit/service/client"
	"unitechs.com/cno/kit/util"
)

// 实例hash

/*
分区消费实现逻辑
实现功能： 定时轮询线程， 超时处理线程等， 当开取多实例的 时，相同的数据 ， 不能每一个实例都处理一次。  每一个实例只处理一部分，  所有实例相加， 则是数据的全部。
以实例个数， 实现分区
1.  程序启动时随机一个id
2. 每 5s, 程序向 redis 写入心跳， 维持当前实例的健康
3. 每 10s， 从 redis 中获取节点信息， 得到健康节点的数量,  并指定实例序号
	实例序号生成方法， 对实例ID 一致性 hash, 得到数字， 从小到大排序。 索引号为实例序号
	一致性 hash 需要的参数  实例个数，  当前实例的序号。
4.  轮询 或 超时线程处理时， 根据 id 做一致性 hash 做负载均衡， 判断当前实现是否消费该数据
*/

type InsRepo interface {
	GetHeartbeat(key string) ([]map[string]string, error)
	Hmset(key string, vals map[string]string) error
	DelInstance(key string, id string) error
	Get(key string) (string, error)
	Set(key, val string) error
}

type InstanceHashStatus string

const (
	InstanceHashStatusRunning InstanceHashStatus = "running"
	InstanceHashStatusStopped InstanceHashStatus = "stopped"
)

type InstanceHash struct {
	Id                      string // cacheKey + id 当前实例的ID号，每一次启动都不一样
	repo                    InsRepo
	cacheKey                string // 存储实例信息的的key值 hash, 默认redis
	instanceHashName        string
	heartBeatIntervalSecond int                // 每 5s, 程序向 redis 写入心跳， 维持当前实例的健康
	getInterval             int                // 每 10s， 从 redis 中获取节点信息， 得到健康节点的数量,  并指定实例序号
	deadTimes               int                // 判定失联/复联的次数 默认3次
	allNumOfInstance        int                // 实例总数
	instanceNum             int                // 当前实例 序号
	status                  InstanceHashStatus // running  stopped
	isMaster                bool
	ctx                     kctx.Ctx
}

var instanceHash *InstanceHash

// IsDealWith  当前告警id 是否当前实例处理
func IsDealWith(ctx kctx.Ctx, id string) bool {
	return instanceHash.IsDeal(ctx, id)
}

func IsMaster() bool {
	return instanceHash.IsMaster()
}

func InitInstanceHash(ctx kctx.Ctx, repo InsRepo, cacheKey string, heartbeatSecond, getSecond, deadTimes int) *InstanceHash {
	ins := NewInstanceHash(ctx, repo, cacheKey, heartbeatSecond, getSecond, deadTimes)
	ins.Run()
	instanceHash = ins
	return ins
}

// NewInstanceHashWith
/*
# 实例hash信息缓存的key 值， 这里不同的程序设置不同的值, 示例如下
InstanceCacheKey = "unios-alarm:subscribe:instance"
# 心跳间隔
InstanceHeartbeatSecond = 3
# 计算实例的时间间隔
InstanceCalcIntervalSecond = 3
# 心中多少次失败算实例已死
InstanceDeadTimes = 3
*/
func NewInstanceHashWith(ctx kctx.Ctx, getter client.IConfigGetter, redisRepoPrefix, instancePrefix string) *InstanceHash {
	rdb := client.NewRedisCli(ctx, getter, redisRepoPrefix)
	repo := kredis.NewRedisRepoWithCli(rdb)

	return InitInstanceHash(ctx, repo,
		getter.GetString(instancePrefix+".InstanceCacheKey"),
		getter.GetIntWithDefault(instancePrefix+".InstanceHeartbeatSecond", 3),
		getter.GetIntWithDefault(instancePrefix+".InstanceCalcIntervalSecond", 3),
		getter.GetIntWithDefault(instancePrefix+".InstanceDeadTimes", 3),
	)
}

func NewInstanceHash(ctx kctx.Ctx, repo InsRepo, cacheKey string, heartbeatSecond, getSecond, deadTimes int) *InstanceHash {
	if heartbeatSecond == 0 {
		heartbeatSecond = 5
	}
	if getSecond == 0 {
		getSecond = 10
	}
	if deadTimes == 0 {
		deadTimes = 3
	}
	ins := &InstanceHash{
		Id:                      util.NewUUIDStr(),
		deadTimes:               deadTimes,
		repo:                    repo,
		cacheKey:                cacheKey,
		heartBeatIntervalSecond: heartbeatSecond,
		getInterval:             getSecond,
		ctx:                     ctx,
	}

	return ins
}

func (i *InstanceHash) Run() {
	i.instanceHashName = i.genInstanceKey()
	go func() {
		ticker := time.NewTicker(time.Second * time.Duration(i.heartBeatIntervalSecond))
		defer ticker.Stop()
		_ = i.pushHeartbeat()
		for {
			select {
			case <-ticker.C:
				if err := i.pushHeartbeat(); err != nil {
					i.ctx.Error("push heartbeat failed, err: ", err)
				}
			}
		}
	}()

	go func() {
		ticker := time.NewTicker(time.Second * time.Duration(i.getInterval))
		defer ticker.Stop()
		_ = i.DealHeartbeat()
		for {
			select {
			case <-ticker.C:
				if err := i.DealHeartbeat(); err != nil {
					i.ctx.Error("heartbeat deal heartbeat failed, err: ", err)
				}
			}
		}
	}()
}

func (i *InstanceHash) SetDeadTimes(n int) {
	i.deadTimes = n
}

// IsMaster  master 节点， 只是名义上的， 人为的指定一个, 没有实际意义
// 使用场景： 当某些操作，只有单进程操作的时候， 用于操作控制
// 这里使用实例hash 排序后的每一个作为master 节点
func (i *InstanceHash) IsMaster() bool {
	return i.isMaster
}

// IsDeal  是否当前实例处理
func (i *InstanceHash) IsDeal(ctx kctx.Ctx, id string) bool {
	if i.allNumOfInstance == 0 {
		return true
	}
	idHash := i.genHash(id)
	n := int(idHash) % i.allNumOfInstance

	//i.ctx.Debugf("compute %d, cur index: %d", n, i.instanceNum)
	ok := n == i.instanceNum
	if !ok {
		//ctx.Infof("IsDeal false allNumOfInstance %d, hash %d, curIndex %d, n %d",
		//	i.allNumOfInstance,
		//	idHash,
		//	i.instanceNum,
		//	n,
		//)
	}

	return ok
}

func (i *InstanceHash) AllNumOfInstance() int {
	return i.allNumOfInstance
}

func (i *InstanceHash) InstanceNum() int {
	return i.instanceNum
}

// ShutDown 手动关闭该实例
//func (i *InstanceHash) ShutDown(key, instanceId string) error {
//	return i.repo.ShutDown(key, instanceId)
//}

const (
	statusRunning = iota + 1
	statusStopped
)

const (
	instanceFieldLatestTime = "latest_time"
	instanceFieldStatus     = "status"    // 1: running, 2: stopped
	instanceFieldDeadTime   = "dead_time" // 1: running, 2: stopped
)

type redisInstance struct {
	LatestTime  time.Time `json:"latest_time"`
	Id          string    `json:"id"`
	DeadTime    string    `json:"dead_time"`
	DeadTimeInt int
}

// DealHeartbeat
// 推送心跳
// hash
// hashName: unios-alarm:alarm:instanceHash:id
// 心跳存储格式
//
//			type redisInstance struct {
//				LatestTime time.Time `json:"latest_time"`//
//		        Id         string    `json: "id"`
//	         deadTime   int   //失败的次数
//			}
func (i *InstanceHash) DealHeartbeat() error {
	valList, err := i.repo.GetHeartbeat(i.cacheKey)
	if err != nil {
		return errors.Wrapf(err, "get heartbeat failed")
	}

	num := make([]hashTemp, 0, len(valList))
	for _, v := range valList {
		b, _ := json.Marshal(v)
		info := redisInstance{}
		if err := json.Unmarshal(b, &info); err != nil {
			i.ctx.Error("heartbeat unmarshal instanceInfo failed, err: %v", err)
			continue
		}
		info.DeadTimeInt, _ = strconv.Atoi(info.DeadTime)

		if info.DeadTimeInt >= i.deadTimes {
			if err := i.repo.DelInstance(i.cacheKey, info.Id); err != nil {
				i.ctx.Error("heartbeat del instance failed, err: ", err)
			}
			continue
		}

		if int(time.Now().Sub(info.LatestTime).Seconds()) > i.heartBeatIntervalSecond {
			if err := i.setDeadTime(info.Id, v, "add"); err != nil {
				i.ctx.Error("set dead time failed, err: ", err)
				continue
			}
		} else {
			if err := i.setDeadTime(info.Id, v, "remove"); err != nil {
				i.ctx.Error("set dead time failed, err: ", err)
				continue
			}
		}

		num = append(num, hashTemp{
			Id:     info.Id,
			hashNo: i.genHash(info.Id),
		})
	}

	// 排序
	sort.Slice(num, func(i, j int) bool {
		return num[i].hashNo < num[j].hashNo
	})
	for index, v := range num {
		if v.Id == i.Id {
			i.instanceNum = index
			break
		}
	}
	i.isMaster = false
	if len(num) != 0 {
		if num[0].Id == i.Id {
			i.isMaster = true
		}
	}

	i.allNumOfInstance = len(num)
	return nil
}
func (i *InstanceHash) setDeadTime(hashName string, m map[string]string, Type string) error {
	switch Type {
	case "add":
		dt, ok := m[instanceFieldDeadTime]
		if !ok {
			m[instanceFieldDeadTime] = "1"
		} else {
			if dt == "" {
				m[instanceFieldDeadTime] = "1"
			} else {
				n, _ := strconv.Atoi(dt)
				if n < i.deadTimes {
					m[instanceFieldDeadTime] = strconv.Itoa(n + 1)
				}
			}
		}
	case "remove":
		dt, ok := m[instanceFieldDeadTime]
		if !ok {
			m[instanceFieldDeadTime] = "0"
		} else {
			if dt == "" {
				m[instanceFieldDeadTime] = "0"
			} else {
				n, _ := strconv.Atoi(dt)
				if n != 0 {
					m[instanceFieldDeadTime] = strconv.Itoa(n - 1)
				}
			}
		}
	}

	return i.repo.Hmset(i.cacheKey+":"+hashName, m)
}

func (i *InstanceHash) pushHeartbeat() error {

	insHashName := i.genInstanceKey()
	if err := i.repo.Hmset(insHashName, map[string]string{
		instanceFieldLatestTime: time.Now().Format(time.RFC3339),
	}); err != nil {
		return errors.Wrapf(err, "set  heartbeat failed")
	}
	if err := i.UpdateHashList(); err != nil {
		return errors.Wrapf(err, "update hash list")
	}

	return nil
}
func (i *InstanceHash) UpdateHashList() error {
	val, err := i.repo.Get(i.cacheKey)
	if err != nil {
		if !strings.Contains(err.Error(), "redis: nil") {
			return errors.Wrapf(err, "get hash list get failed")
		}
	}
	lst := gset.NewStrSetFrom(strings.Split(val, ","))
	lst.Add(i.Id)
	if err = i.repo.Set(i.cacheKey, strings.Join(lst.Slice(), ",")); err != nil {
		return errors.Wrapf(err, "set hash list set failed")
	}

	return nil
}

func (i *InstanceHash) genInstanceKey() string {
	return i.cacheKey + ":" + i.Id
}

func (i *InstanceHash) dealHeartbeat(h map[string]string) {
	n := time.Now()

	num := make([]hashTemp, 0, len(h))

	for id, timeStr := range h {
		t, err := time.Parse(time.RFC3339, timeStr)
		if err != nil {
			i.ctx.Error("get instance hash.dealHeartbeat parse time failed err: ", err)
			continue
		}
		temp := int(n.Sub(t).Seconds())
		if temp >= i.getInterval {
			// 超过时间， 删除
			//if err := i.repo.DelInstance(i.cacheKey, id); err != nil {
			//	i.ctx.Error("del instance failed, err: ", err)
			//	continue
			//}
			continue
		}
		num = append(num, hashTemp{
			Id:     id,
			hashNo: i.genHash(id),
		})
	}
	// 排序
	sort.Slice(num, func(i, j int) bool {
		return num[i].hashNo < num[j].hashNo
	})
	for index, v := range num {
		if v.Id == i.Id {
			i.instanceNum = index
			break
		}
	}
	i.allNumOfInstance = len(num)
}

type hashTemp struct {
	Id     string
	hashNo uint32
}

// genHash 计算 一致性hash
func (i *InstanceHash) genHash(id string) uint32 {
	ieee := crc32.NewIEEE()
	_, _ = io.WriteString(ieee, id)
	return ieee.Sum32()
}
