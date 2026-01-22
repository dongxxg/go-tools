// 任务号    日期     开发者      描述

package kctx

import (
	"context"
	"testing"
)

func TestCtx(t *testing.T) {
	c := NewContext(context.TODO(), "aaaaaaa")
	c.Infof("aaa")
	c = c.WithField("name", "wiken")
	c.Infof("bbb")
}
