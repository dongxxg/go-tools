package cache

import (
	"encoding/json"
	"github.com/gogf/gf/v2/container/gset"
	"sync"
	"unitechs.com/cno/kit/kctx"
	"unitechs.com/cno/kit/template/model"
)

type RuleCache struct {
	data map[string]*model.Rule
	mu   sync.RWMutex
	ctx  kctx.Ctx
}

func (c *RuleCache) Range(f func(rule2 *model.Rule) bool) {

	for _, v := range c.data {
		if v == nil {
			continue
		}
		if v.RuleStatus != 1 {
			continue
		}
		if !f(v) {
			break
		}
	}
}

// InitReceiver 初始化接收器
func (c *RuleCache) InitReceiver() {

}

var rule *RuleCache

func NewRuleCache(ctx kctx.Ctx) *RuleCache {
	rule = &RuleCache{
		data: make(map[string]*model.Rule),
		ctx:  ctx,
	}
	return rule
}

func (c *RuleCache) Set(key string, value model.Rule) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if r, ok := c.data[key]; ok {
		if r.Receiver != nil {
			value.Receiver = r.Receiver
			value.HandleFunc = r.HandleFunc
			if err := r.Receiver.Close(); err != nil {
				c.ctx.Errorf("rule %s  ruleType: %s close failed: %v", r.RuleId, r.RuleType, err)
			} else {
				c.ctx.Infof("rule %s  ruleType: %s close success", r.RuleId, r.RuleType)
			}
		}
	}

	c.data[key] = &value
}

func (c *RuleCache) set(key string, value model.Rule) {
	c.data[key] = &value
}

func (c *RuleCache) SetWithoutLock(key string, value model.Rule) {
	c.data[key] = &value
}

// GetTmplId 获取模板id
func (c *RuleCache) GetTmplId() []string {
	l := gset.NewStrSet()
	for _, v := range c.data {
		l.Add(v.TemplId)
	}
	return l.Slice()
}

func (c *RuleCache) SetByMap(ctx kctx.Ctx, m map[string]string) (ruleIds []string) {
	g := gset.NewStrSet()

	for k, v := range m {
		temp := model.Rule{}
		if err := json.Unmarshal([]byte(v), &temp); err != nil {
			ctx.Errorf("failed to unmarshal to rule %s, err: %v", v, err)
			continue
		}
		g.Add(temp.RuleId)
		c.Set(k, temp)
	}
	return g.Slice()
}

func (c *RuleCache) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.data)
}

func (c *RuleCache) Get(key string) (model.Rule, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	value, ok := c.data[key]
	if !ok {
		return model.Rule{}, false
	}
	return *value, ok
}

func (c *RuleCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.data, key)
}
