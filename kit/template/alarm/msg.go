// 任务号    日期     开发者      描述

package alarm

import (
	"github.com/spf13/cast"
	"sync"
	"unitechs.com/cno/kit/kctx"
	"unitechs.com/cno/kit/util"
)

type MsgMap struct {
	m map[string]interface{}
	sync.Mutex
	condKey string // 条件key 值
}

func (m *MsgMap) GetStr(key string) string {
	n, ok := m.Get(key)
	if !ok {
		return ""
	}

	return cast.ToString(n)
}

func NewMsgMap(condKey string) *MsgMap {
	return &MsgMap{
		m:       make(map[string]interface{}),
		condKey: condKey,
	}
}
func NewMsgMapWithMap(condKey string, d map[string]interface{}) *MsgMap {
	return &MsgMap{
		m:       d,
		condKey: condKey,
	}
}
func (m *MsgMap) Set(d map[string]interface{}) {
	m.Lock()
	defer m.Unlock()
	m.m = d
}

func (m *MsgMap) Get(k string) (interface{}, bool) {
	m.Lock()
	defer m.Unlock()
	v, ok := m.m[k]
	return v, ok
}

func (m *MsgMap) GetCondKey() (key, val string, exist bool) {
	if m.condKey == "" {
		return "", "", false
	}
	v, ok := m.Get(m.condKey)
	if !ok {
		return m.condKey, "", false
	}

	return m.condKey, util.ConvertToString(v), true
}

type SerialNoDemo struct {
	n int64
}

func (s *SerialNoDemo) Get(ctx kctx.Ctx, ruleId, templateId, dstField string, start, interval int64) (int64, error) {
	s.n += 1
	return s.n, nil
}
