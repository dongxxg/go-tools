/**
  @Author: dongxx
  @Since: 2026/4/8 13:53
  @Desc: //TODO
**/

package tests

import (
	"encoding/json"
	"github.com/pkg/errors"
	"log"
	"strings"
	"testing"
	"time"
	"unitechs.com/unios-dice/uni-base/core/config"
	"unitechs.com/unios-dice/uni-base/lib/stores/redis"
)

type ResList struct {
	NodeCode string `json:"nodeCode"`
	NodeName string `json:"nodeName"`
	ResId    string `json:"resId"`
	ResName  string `json:"resName"`
	ResPara  string `json:"resPara"`
}

type EngiRule struct { // 规则主表
	CreateUserId string    `json:"createUserId"`
	DutyMan      string    `json:"dutyMan"`
	EndTime      string    `json:"endTime"`
	EngiDesc     string    `json:"engiDesc"`
	EngiName     string    `json:"engiName"`
	EngiRes      []ResList `json:"engiRes"`
	EngiType     int       `json:"engiType"`
	EngiID       string    `json:"engiid"`
	NodeCode     string    `json:"nodeCode"`
	NodeName     string    `json:"nodeName"`
	Specialty    string    `json:"specialty"`
	StartTime    string    `json:"startTime"`
	Tid          int       `json:"tid"`
	TimeSeg      string    `json:"timeSeg"`
	SubResMap    map[string]map[string]int
}

type EngiInfo struct {
	ID      string    `json:"id"`
	OutTime time.Time `json:"outime"`
}

// MakeAssoRuleByString json解析
func MakeEngiRuleByString(s string) (*EngiRule, error) {
	var res EngiRule
	err := json.Unmarshal([]byte(s), &res)
	if err != nil {
		return nil, errors.New(" make engi rule error:" + err.Error())
	}
	err = InitEngiRule(&res)
	return &res, err
}
func InitEngiRule(res *EngiRule) error {
	res.SubResMap = make(map[string]map[string]int)
	for i := range res.EngiRes {
		if _, ok := res.SubResMap[res.EngiRes[i].ResName]; !ok {
			res.SubResMap[res.EngiRes[i].ResName] = make(map[string]int)
		}
		res.SubResMap[res.EngiRes[i].ResName][res.EngiRes[i].ResPara] = 1
	}
	return nil
}

func (s EngiInfo) Marshal() string {
	b2, _ := json.Marshal(s)
	return string(b2)
}

func GetEngiTimeFlag(engirule *EngiRule, currentTime time.Time) bool {
	//设置天粒度时间范围
	//"startTime": "2023-08-08","timeSeg": "00:04—17:04"
	SegArray := strings.Split(engirule.TimeSeg, ",")
	EngiFlag := false
	for index, _ := range SegArray {
		TimeSegArray := strings.Split(SegArray[index], "—")
		StartTime := engirule.StartTime + " " + TimeSegArray[0] + ":00"
		EndTime := engirule.EndTime + " " + TimeSegArray[1] + ":00"
		workStartTime, _ := time.ParseInLocation("2006-01-02 15:04:05", StartTime, time.Local)
		workEndTime, _ := time.ParseInLocation("2006-01-02 15:04:05", EndTime, time.Local)
		log.Printf("GetEngiTimeFlag workStartTime %v,workEndTime %v", workStartTime, workEndTime)
		currentTime = currentTime.Truncate(time.Second)
		// 判断当前时间是否在工作时间范围内
		if currentTime.After(workStartTime) && currentTime.Before(workEndTime) {
			EngiFlag = true
			break
		} else {
			log.Printf("GetEngiTimeFlag outtime currentTime %v,workStartTime %v,workEndTime %v", currentTime, workStartTime, workEndTime)
		}
	}
	return EngiFlag
}

var RedisConn *redis.Redis

func init() {
	if RedisConn != nil {
		return
	}
	conf := redis.RedisConf{
		Host: "192.168.201.59:30009",
		Type: "node",
		Pass: "",
		Db:   0,
	}
	RedisConn = conf.NewRedis()

	if RedisConn == nil {
		log.Println("failed to connect to redis....")
		return
	}
	log.Println("Successfully connected to redis")
}

func TestEngi(t *testing.T) {
	//	str := `{
	//    "createUserId": "hxw管理员",
	//    "dutyMan": "hxw管理员",
	//    "endTime": "2026-04-09",
	//    "engiDesc": "",
	//    "engiName": "测试工程打标",
	//    "engiRes": [
	//        {
	//            "nodeName": "重庆运控中心",
	//            "resId": "SubNetwork=1,ManagedElement=1",
	//            "resName": "AMF",
	//            "resPara": "SubNetwork=1,ManagedElement=1"
	//        },
	//        {
	//            "nodeName": "重庆运控中心",
	//            "resId": "SCXWZZZMAM00000001",
	//            "resName": "XWSAMF01",
	//            "resPara": "XWSAMF01"
	//        }
	//    ],
	//    "engiType": 100,
	//    "engiid": "a3e0c159-eff5-4554-ab09-f3e279af7d65",
	//    "nodeCode": "503",
	//    "nodeName": "重庆运控中心",
	//    "specialty": "HX",
	//    "startTime": "2026-04-08",
	//    "tid": 0,
	//    "timeSeg": "11:00—19:00"
	//}`
	data, err := RedisConn.Hgetall(config.GetString("Redis.Key.EngiKey"))
	if err != nil {
		log.Printf("load engi_rule keys error:%s", err.Error())
		return
	}

	for _, valueStr := range data {
		t.Logf("valueStr %s", valueStr)
		r, err := MakeEngiRuleByString(valueStr)
		if err != nil {
			t.Logf("MakeEngiRuleByString failed %v", err)
			return
		}
		t.Logf("SubResMap %v", r.SubResMap)
		if r.EngiType == 100 { //设备级别
			if r.SubResMap["XWSAMF01"]["XWSAMF01"] == 1 {
				//判断是否满足工程预约时间范围
				currentTime := time.Now()
				EngiFlag := GetEngiTimeFlag(r, currentTime)
				if !EngiFlag {
					t.Log("resname: XWSAMF01 not in engirule timerange")
					return
				}
				t.Logf("r %v", r)
			}
		}
	}
}