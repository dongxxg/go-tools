package rule

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unitechs.com/cno/kit/kctx"
	"unitechs.com/cno/kit/template/model"
	util2 "unitechs.com/cno/kit/util"
)

// JudgeMethod 属性值字段不存在时，判断处理方法
type JudgeMethod int

const (
	JudgeMethodTrue  JudgeMethod = iota // 如果属性字段不存在， 判断为true
	JudgeMethodFalse                    // 如果属性字段不存在， 判断为false
)

// IsMatch 数据是否匹配表达式
func IsMatch(ctx kctx.Ctx, exp string, data model.IMsg, method JudgeMethod) (bool, error) {
	f, err := MakeFilter(ctx, exp, method)
	if err != nil {
		return false, err
	}
	return f(data), nil
}

// DoMatch 为filterLst 匹配返回规则列表
func DoMatch(ctx kctx.Ctx, filters model.FilterList, data model.IMsg, method JudgeMethod) model.FilterList {
	for k, f := range filters {
		matched, err := IsMatch(ctx, f.FilterExpr, data, method)
		if err != nil {
			ctx.Errorf("[DoMatch] match for expr <%s> failed, err: %v",
				f.FilterExpr, err)
			continue
		}
		filters[k].IsMatched = matched
	}
	return filters
}

func qg(s string, sep string) (flag bool, err error, field string, value string, valueInt int, isInt bool) {
	if strings.Contains(s, sep) { // ==
		flag = true
		infos := strings.Split(s, sep)
		field, value = infos[0], infos[1]
		field = strings.Trim(field, " ")
		value = strings.Trim(value, " ")
		if !strings.Contains(value, "'") {
			isInt = true
			valueInt, err = strconv.Atoi(value)
			if err != nil {
				err = errors.New("value must be int :" + value) //+ " expression:" + s
			}
		} else {
			value = strings.Trim(value, "'")
		}
	}
	return
}

var parenthesReg = regexp.MustCompile(`\([^()]+\)`)

type subSTotal struct {
	count int
	m     map[string]func(m model.IMsg) bool
}

func getStrAttr(ex model.IMsg, name string) (string, bool) {
	val, e := ex.Get(name)
	if !e {
		return "", false
	}
	switch value := val.(type) {
	case string:
		return value, true
	case time.Time:
		return value.Format("2006-01-02 15:04:05"), true
	default:
		return fmt.Sprint(value), true
	}
}
func getIntAttr(ex model.IMsg, name string) (int, bool) {
	val, e := ex.Get(name)
	if !e {
		return 0, false
	}
	switch value := val.(type) {
	case string:
		atoi, err := strconv.Atoi(value)
		if err != nil {
			kctx.Error(err.Error())
			return 0, false
		}
		return atoi, true
	case time.Time:
		kctx.Error("can not check time to number")
		return 0, false
	case int:
		return value, true
	case float64: //ext 转换可能会变成这个
		return int(value), true

	default:
		n, err := util2.ConvertToInt(val)
		if err != nil {
			kctx.Error(fmt.Sprintf("can not parse to int  %v err: %v", val, err))
			return 0, false
		}

		return int(n), true
	}
}

var seps = []string{">=", ">", "<=", "<", "==", "!="}

func bjStr(ctx kctx.Ctx, left, right, sep string) bool {
	// 规则为时间
	t2, err := util2.ParseToTime(right)
	if err == nil {
		if left == "" {
			switch sep {
			case ">", ">=":
				return true
			case "=", "<", "<=":
				return false
			}
		}
		t1, err := util2.ParseToTime(left)
		if err == nil {
			switch sep {
			case ">", ">=":
				return t1.After(t2)
			case "<", "<=":
				return t1.Before(t2)
			case "==":
				return t1.Equal(t2)
			}
		} else {
			ctx.Errorf("failed to parse string to time err: %v", err)
		}
	} else {
		ctx.Debugf("right str is not time format %s", right)
	}

	switch sep {
	case ">":
		return left > right
	case ">=":
		return left > right
	case "<":
		return left < right
	case "<=":
		return left <= right
	case "==":
		return left == right
	case "!=":
		return left != right
	}
	return false
}

func bjInt(left, right int, sep string) bool {
	switch sep {
	case ">":
		return left > right
	case ">=":
		return left > right
	case "<":
		return left < right
	case "<=":
		return left <= right
	case "==":
		return left == right
	case "!=":
		return left != right
	}
	return false
}

func mkFilter(ctx kctx.Ctx, s string, subS *subSTotal, method JudgeMethod) (func(m model.IMsg) bool, error) {
	var filterFuncs []func(m model.IMsg) bool
	formulas := strings.Split(s, "||")
	for _, formula := range formulas {
		if formula == "" {
			continue
		}

		subformulas := strings.Split(formula, "&&")
		var subFilterFuncs []func(m model.IMsg) bool
		for _, subformula := range subformulas {
			subformula = strings.Trim(subformula, " ")

			if subformula == "" {
				continue
			}

			var ret bool
			for _, sep := range seps {
				flag, err, fieldName, value, valInt, isInt := qg(subformula, sep)
				if flag {
					if err != nil {
						return nil, errors.New(err.Error() + " expression:" + s)
					}
					if isInt {
						subFilterFuncs = append(subFilterFuncs, func(m model.IMsg) bool {
							val, ok := getIntAttr(m, fieldName)
							if !ok || !bjInt(val, valInt, sep) {
								ctx.Debugf(" isMatch: false, left: %d, operator:%s, right: %d",
									val, sep, valInt)
								return false
							}
							return true
						})
					} else {
						subFilterFuncs = append(subFilterFuncs, func(m model.IMsg) bool {
							val, ok := getStrAttr(m, fieldName)
							if !ok || !bjStr(ctx, val, value, sep) {
								ctx.Debugf("isMatch: false, left: %s, operator:%s, right: %s",
									val, sep, value)
								return false
							}
							return true
						})
					}
					ret = true
					break
				}
			}
			if ret {
				continue
			}

			if strings.Contains(subformula, " not in [") {
				infos := strings.Split(subformula, " not in [")
				field, value := infos[0], infos[1]
				field = strings.Trim(field, " ")
				value = strings.Trim(value, " ")
				value = value[0 : len(value)-1]
				if !strings.Contains(value, "'") {
					// 说明是int 类型
					var valNums []int
					for _, numStr := range strings.Split(value, ",") {
						numStr = strings.Trim(numStr, " ")
						valueNum, err := strconv.Atoi(numStr)
						if err != nil {
							return nil, errors.New("value must be in :" + value + " expression:" + s)
						}
						valNums = append(valNums, valueNum)
					}
					subFilterFuncs = append(subFilterFuncs, func(m model.IMsg) bool {
						val, e := getStrAttr(m, field)
						if !e {
							ctx.Errorf("get field %s failed, %v", field, e)

							if method == JudgeMethodTrue {
								return true
							}
							return false
						}
						n, err := strconv.ParseInt(val, 10, 64)
						if err != nil {
							ctx.Errorf("convert %s=%v to int failed %v", field, val, err)
							return false
						}
						for _, valNum := range valNums {
							if valNum == int(n) {
								return false
							}
						}
						return true
					})
				} else {
					valStrList := strings.Split(value, ",")
					var valList []string
					for _, valStr := range valStrList {
						valStr = strings.Trim(valStr, " ")
						valStr = strings.Trim(valStr, "'")
						valList = append(valList, valStr)
					}
					subFilterFuncs = append(subFilterFuncs, func(m model.IMsg) bool {
						val, ok := getStrAttr(m, field)
						if !ok {
							ctx.Errorf("get field %s failed", field)
							if method == JudgeMethodTrue {
								return true
							}
							return true
						}

						for _, v := range valList {
							if v == val {
								return false
							}
						}
						return true
					})
				}
			} else if strings.Contains(subformula, " in [") { // ['a','b','c']  [1,2,3]
				infos := strings.Split(subformula, " in [")
				field, value := infos[0], infos[1]
				value = strings.Trim(value, " ")
				value = value[0 : len(value)-1]
				if !strings.Contains(value, "'") {
					var valNums []int
					for _, numStr := range strings.Split(value, ",") {
						numStr = strings.Trim(numStr, " ")
						valueNum, err := strconv.Atoi(numStr)
						if err != nil {
							return nil, errors.New("value must be in :" + value + " expression:" + s)
						}
						valNums = append(valNums, valueNum)
					}
					subFilterFuncs = append(subFilterFuncs, func(m model.IMsg) bool {
						v, ok := getIntAttr(m, field)
						if !ok {
							ctx.Errorf("get field %s failed", field)
							if method == JudgeMethodTrue {
								return true
							}
							return false
						}

						for _, valNum := range valNums {
							if valNum == v {
								return true
							}
						}
						return false
					})
				} else {
					valStrList := strings.Split(value, ",")
					var valList []string
					for _, valStr := range valStrList {
						valStr = strings.Trim(valStr, " ")
						valStr = strings.Trim(valStr, "'")
						valList = append(valList, valStr)
					}
					subFilterFuncs = append(subFilterFuncs, func(m model.IMsg) bool {
						val, ok := getStrAttr(m, field)
						if !ok {
							ctx.Errorf("get field %s failed, %v", field, ok)
							if method == JudgeMethodTrue {
								return true
							}
							return false
						}
						for _, v := range valList {
							if v == val {
								return true
							}
						}
						return false
					})
				}
			} else if strings.Contains(subformula, "matches") {
				infos := strings.Split(subformula, " matches")
				fieldName, regStr := infos[0], infos[1]
				regStr = strings.Trim(regStr, " ")
				regStr = strings.Trim(regStr, "'")
				reg, err := regexp.Compile(regStr)
				if err != nil {
					return nil, errors.New("compile reg str error:" + regStr)
				}
				subFilterFuncs = append(subFilterFuncs, func(m model.IMsg) bool {
					value, ok := getStrAttr(m, fieldName)
					if !ok {
						ctx.Errorf("get field %s failed", fieldName)
						if method == JudgeMethodTrue {
							return true
						}
						return true
					}
					return reg.MatchString(value)
				})
			} else if strings.HasPrefix(subformula, "$") {
				subFilterFuncs = append(subFilterFuncs, subS.m[subformula])
			} else if strings.Contains(subformula, "include") {
				infos := strings.Split(subformula, "include")
				if len(infos) < 2 {
					return nil, fmt.Errorf("invalid expr %s", subformula)
				}
				field, val := strings.Trim(infos[0], " "), strings.Trim(infos[1], " ")
				val = strings.Trim(val, "'")
				subFilterFuncs = append(subFilterFuncs, func(m model.IMsg) bool {
					v, e := getStrAttr(m, field)
					if !e {
						ctx.Errorf("get field %s failed", field)
						if method == JudgeMethodTrue {
							return true
						}
						return true
					}
					b := strings.Contains(v, val)
					return b
				})
			} else if strings.Contains(subformula, "exclude") {
				infos := strings.Split(subformula, "exclude")
				if len(infos) < 2 {
					return nil, fmt.Errorf("invalid expr %s", subformula)
				}
				field, val := strings.Trim(infos[0], " "), strings.Trim(infos[1], " ")
				val = strings.Trim(val, "'")
				subFilterFuncs = append(subFilterFuncs, func(m model.IMsg) bool {
					v, e := getStrAttr(m, field)
					if !e {
						ctx.Errorf("get field %s failed, %v", field, e)
						if method == JudgeMethodTrue {
							return true
						}
						return true
					}
					return !strings.Contains(v, val)
				})
			} else {

				return nil, errors.New("got error expression：" + subformula)
			}
		}
		if len(subformulas) > 0 {
			filterFuncs = append(filterFuncs, func(m model.IMsg) bool {
				for _, fx := range subFilterFuncs {
					if !fx(m) {
						return false
					}
				}
				return true
			})
		}
	}
	return func(m model.IMsg) bool {
		for _, fx := range filterFuncs {
			if fx(m) {
				return true
			}
		}
		return false
	}, nil
}

// MakeFilter 根据字符串转换为过滤函数
func MakeFilter(ctx kctx.Ctx, s string, method JudgeMethod) (func(m model.IMsg) bool, error) {
	subS := subSTotal{m: map[string]func(m model.IMsg) bool{}}
	var e []error
	//for {
	//	if !strings.Contains(s, "(") {
	//		break
	//	}
	//	s = parenthesReg.ReplaceAllStringFunc(s, func(subStr string) string {
	//		subS.count += 1
	//		subName := "$" + strconv.Itoa(subS.count)
	//		f, err := mkFilter(ctx, subStr[1:len(subStr)-1], &subS, 0)
	//		if err != nil {
	//			e = append(e, err)
	//			subS.m[subName] = nil
	//		} else {
	//			subS.m[subName] = f
	//		}
	//		return subName
	//	})
	//}
	res, err := mkFilter(ctx, s, &subS, method)
	if err != nil {
		e = append(e, err)
	}
	if len(e) > 0 {
		errInfo := ""
		for i := range e {
			errInfo += e[i].Error()
		}
		return nil, errors.New(errInfo)
	}
	return func(m model.IMsg) bool {
		defer func() {
			err := recover()
			if err != nil {
				ctx.Errorf("filter func error: %+v", err)
			}
		}()
		return res(m)
	}, err
}
