package media

import (
	"fmt"
	"testing"
)

func TestNewM3u8Parser(t *testing.T) {
	parser := NewM3u8Parser([]byte("\nbc\n1111\n12341234\n"), nil)
	it, hasNext := parser.GetLineIterator()
	for hasNext {
		var item string
		item, hasNext = it()
		fmt.Println(item)
	}

}
func TestM3u8Parser_GetTsInfoList(t *testing.T) {
	m3u8 := "#EXTM3U\n#EXT-X-VERSION:3\n#EXT-X-MEDIA-SEQUENCE:0\n#EXT-X-ALLOW-CACHE:YES\n#EXT-X-TARGETDURATION:17\n#EXTINF:16.000000,\ns1/9be3441078466b4be7d339b0b1a94372_R_plaso_04967163-879a-48df-ba95-124539d3a5ca_20230726085204295.ts\n#EXTINF:16.000000,\ns1/9be3441078466b4be7d339b0b1a94372_R_plaso_04967163-879a-48df-ba95-124539d3a5ca_20230726085220298.ts"
	parser := NewM3u8Parser([]byte(m3u8), ParseStartTimeFromTsPath4AGORA)
	res := parser.GetTsInfoList()
	fmt.Println(res)
	fmt.Println(parser.GetDuration())
}
func TestParseExtXProgramDateTime(t *testing.T) {
	line := "#EXT-X-PROGRAM-DATE-TIME:2023-09-14T01:35:18.601+00:00"
	time, e := ParseExtXProgramDateTime(line)
	fmt.Println(time, e)
}
