// 任务号    日期     开发者      描述

package kerr

type ErrKeyNotExit struct {
	Msg string
}

func (e ErrKeyNotExit) Error() string {
	return "key not exist %s" + e.Msg
}
