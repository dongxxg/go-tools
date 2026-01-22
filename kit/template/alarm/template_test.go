// 任务号    日期     开发者      描述

package alarm

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"
)

func TestGenTransferInfo(t *testing.T) {
	b, _ := os.ReadFile("template_ab.json")
	out := make(map[string]interface{})
	json.Unmarshal(b, &out)
	b, _ = json.Marshal(out)
	s := string(b)
	s = "hset unios-alarm:transformTempl  alarm-attribute \"" + strings.ReplaceAll(s, "\"", "\\\"") + "\""
	// hget unios-alarm:transformTempl  alarm-attribute
	fmt.Println(s)
}

func TestReg(t *testing.T) {
	//s := "http://RESOURCE-JAVA-SERVICE/api/ala/getCircuitNo?resid=${k_RESID}"
	s := "http://127.0.0.1:8000/api/ala/getCircuitNo?resid=${k_RESID}"
	reg, _ := regexp.Compile(`://(.*?)/`)
	lst := reg.FindStringSubmatch(s)
	fmt.Println(lst)
}
