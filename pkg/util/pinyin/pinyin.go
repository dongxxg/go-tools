package pinyin

import (
	"regexp"

	"github.com/mozillazg/go-pinyin"
)

func GetPinyin(str string) string {
	// 防止特殊字符
	reg := regexp.MustCompile(`[^\p{L}]`)
	str = reg.ReplaceAllString(str, "")
	a := pinyin.NewArgs()
	a.Separator = ""
	a.Fallback = func(r rune, a pinyin.Args) []string {
		return []string{string(r)}
	}
	result := pinyin.Slug(str, a)
	return result
}
