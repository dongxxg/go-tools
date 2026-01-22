package util

import (
	"encoding/json"
	"github.com/gogf/gf/v2/container/gset"
	"github.com/spf13/viper"
	"os"
	"strings"
	"unitechs.com/cno/kit/kctx"
)

// NewEnv 将配置文件 生成环境变量， 以文件的形式存储在路径中
// 默认 path ./conf/env.json
func NewEnv(ctx kctx.Ctx, path string) {
	if path == "" {
		path = "./conf/env.toml"
	}
	out := make(map[string]interface{})
	viper.GetViper().Unmarshal(&out)

	ctx.Debugf("out %v", out)
	bb, _ := json.Marshal(out)
	lst, _ := ConvertJsonString(string(bb), "KeyValue")

	ctx.Debugf("lst %v", lst)
	fName := "./conf/env_filter"
	b, err := os.ReadFile(fName)
	if err != nil {
		ctx.Errorf("read %s failed: %v", fName, err)
	}
	filterList := gset.NewStrSetFrom(strings.Split(string(b), "\n"))
	//writeToExcel(ctx, strings.Join(lst, "\n"), filterList)
	var bbb string
	for _, v := range lst {
		temp := strings.Split(v, "=")
		if len(temp) < 2 {
			continue
		}
		key := strings.Trim(temp[0], "KeyValue.")
		if !filterList.Contains(key) {
			continue
		}
		bbb += v + "\n"
	}

	_ = os.WriteFile(path, []byte(bbb), 0644)
}

//func writeToExcel(ctx kctx.Ctx, s string, filterList *gset.StrSet) {
//	lst := strings.Split(s, "\n")
//
//	f := excelize.NewFile()
//	sheetName := "simple"
//	sheetNameAll := "all"
//	f.NewSheet(sheetName)
//	f.NewSheet(sheetNameAll)
//	fmt.Println(filterList)
//	commitMap := getCommit(ctx, "./conf/configuration.toml")
//
//	n := 1
//	nAll := 1
//	for _, v := range lst {
//		lst2 := strings.Split(v, "=")
//		if len(lst2) < 2 {
//			continue
//		}
//		temp := []string{"keyValue"}
//
//		k := strings.Trim(lst2[0], "KeyValue.")
//		temp = append(temp, k, strings.Trim(ConvertToString(strings.Join(lst2[1:], "=")), "\""))
//		if c, ok := commitMap[k]; ok {
//			temp = append(temp, c)
//		}
//
//		if err := f.SetSheetRow(sheetNameAll, fmt.Sprintf("A%d", nAll), &temp); err != nil {
//			ctx.Errorf("set sheet %s row failed: %v", sheetNameAll, err)
//		}
//		nAll++
//		if !filterList.Contains(k) {
//			continue
//		}
//		if err := f.SetSheetRow(sheetName, fmt.Sprintf("A%d", n), &temp); err != nil {
//			ctx.Errorf("set sheet %s row failed: %v", sheetNameAll, err)
//		}
//		n++
//	}
//
//	fileName := "./conf/env.xlsx"
//	if err := f.SaveAs(fileName); err != nil {
//		ctx.Errorf("save %s  failed: %v", fileName, err)
//	}
//}

func getCommit(ctx kctx.Ctx, fileName string) map[string]string {
	b, err := os.ReadFile(fileName)
	if err != nil {
		ctx.Errorf("read configFile failed: %v", err)
		return nil
	}

	lines := strings.Split(string(b), "\n")
	out := make(map[string]string)
	/*
		# 是否注册到nacos, 如果为false, 需要手动改为true
		位置1  s1
		Enabled = true
		# 注册到nacos 的地址， 多网卡时需要配置， 单网卡时可为空
		Host = "" # 手动配置注册到nacos 的地址  位置2 s2
	*/
	var s1, s2, group string

	for _, v := range lines {
		if v == "" {
			continue
		}
		if strings.HasPrefix(v, "#") && strings.Contains(v, "=") {
			continue
		}
		if strings.Contains(v, "[") {
			v = strings.Trim(v, "[")
			v = strings.Trim(v, "\r")
			v = strings.Trim(v, "]")
			group = v
			continue
		}

		if strings.Contains(v, "#") && !strings.Contains(v, "=") {
			// 位置1
			s1 = strings.Split(v, "#")[1]
		}
		if strings.Contains(v, "#") && strings.Contains(v, "=") {
			// 位置1
			s2 = strings.Split(v, "#")[1]
		}
		if strings.Contains(v, "=") {
			temp := strings.Split(v, "=")
			if len(temp) < 2 {
				continue
			}
			key := strings.Trim(temp[0], " ")
			key = group + "_" + key
			key = strings.ToUpper(strings.ReplaceAll(key, ".", "_"))

			if s1 != "" {
				out[key] = s1
				s1 = ""
				s2 = ""
			}
			if s2 != "" {
				out[key] = s2
				s1 = ""
				s2 = ""
			}
		}
	}
	return out
}
