package main

import (
	"fmt"
	"regexp"
)

func main() {
	//pattern := "a[a-z]b*"
	//s := "acddb"
	//compile, err := regexp.Compile(pattern)
	//if err != nil {
	//	fmt.Printf("Fail to compile pattern:%v", err)
	//}
	//res := compile.FindAllString(s, 4)
	//fmt.Println(res)

	text := "MENAME:... \r\nurl = https://45.138.136.4:31943/fileserver/v1/web/download?filepath=/EXP_MML/.../EXP_MML_SOCC-CQ-NMS-SCN08VI1-AMF08_20260312095922.zip&DirId=13已完成 100%\\r\\n(结果个数 = 1)\r\n共有6个报告\r\n---   END\r\n"
	//match, err := regexp.Compile("\\r\\n共有\\d+个报告\\r\\n---\\s*END\\r\\n")
	//match, err := regexp.Compile("\\r\\nurl\\s*=\\s*https?://[^\\r\\n]+\\r\\n.*\\r\\n---\\s+END\\r\\n")
	//match, err := regexp.Compile("(?s)\\r\\nurl\\s*=\\s*https?://[^\\r\\n]+\\r\\n.*?\\r\\n---\\s+END\\r\\n")
	//match, err := regexp.Compile("\\r?\\nurl\\s*=\\s*https?://[^\\r\\n]+\\r?\\n.*\\r?\\n---\\s+END\\r?\\n")
	//match, err := regexp.Compile("url\\s*=\\s*[\\s\\S]*?---\\s+END\\r?\\n")
	//match, err := regexp.Compile("(?s)url\\s*=.*?---\\s*END")
	match, err := regexp.Compile("url[\\s\\S]*END")
	if err != nil {
		fmt.Printf("%s", err.Error())
		return
	}
	re := match.MatchString(text)
	fmt.Println(re)
}