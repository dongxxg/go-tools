package cache

import (
	"encoding/json"
	"sync"
	"unitechs.com/cno/kit/kctx"
	"unitechs.com/cno/kit/template/model"
)

// ConvertTemplate 转换模板
type ConvertTemplate struct {
	data map[string]model.ConvertRule
	mu   sync.RWMutex
}

func NewConvertTemplate() *ConvertTemplate {
	return &ConvertTemplate{data: make(map[string]model.ConvertRule)}
}

func (c *ConvertTemplate) Get(key string) (model.ConvertRule, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, ok := c.data[key]
	return val, ok
}

func (c *ConvertTemplate) Set(key string, value model.ConvertRule) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = value
}
func (c *ConvertTemplate) SetByMap(ctx kctx.Ctx, m map[string]string) {
	for k, v := range m {
		temp := model.ConvertRule{}
		if err := json.Unmarshal([]byte(v), &temp); err != nil {
			ctx.Errorf("failed to unmarshal to rule %s, err: %v", v, err)
			continue
		}
		c.Set(k, temp)
	}
}

func (c *ConvertTemplate) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.data, key)
}
