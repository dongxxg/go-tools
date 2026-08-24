package media

import "strings"

type M3u8Info struct {
	//duration int // 单位毫秒
	content []byte
}

func (m *M3u8Info) GetContent() []byte {
	return m.content
}

type Generator interface {
	Begin()
	AppendLine([]byte)
	End() M3u8Info
}

var M3u8HeadText = []string{"#EXTM3U",
	"#EXT-X-VERSION:3",
	"#EXT-X-MEDIA-SEQUENCE:0",
	"#EXT-X-ALLOW-CACHE:YES",
	"#EXT-X-TARGETDURATION:17",
}

const status_init = 1
const status_start = 2
const status_end = 3

var M3u8TailText = "#EXT-X-ENDLIST"

type M3u8Generator struct {
	m3u8   *M3u8Info
	status int
}

func NewM3u8Generator() *M3u8Generator {
	return &M3u8Generator{
		m3u8: &M3u8Info{
			//duration: 0,
			content: make([]byte, 0, 1000),
		},
		status: status_init,
	}
}

func (m *M3u8Info) writeLine(line string) {
	m.writeInfo(line)
	m.content = append(m.content, '\n')
}
func (m *M3u8Info) writeInfo(line string) {
	m.content = append(m.content, line...)
}

func (m *M3u8Info) writeHead() {
	for _, headText := range M3u8HeadText {
		m.writeLine(headText)
	}
}
func (m *M3u8Info) writeTail() {
	m.writeInfo(M3u8TailText)
}

func (m *M3u8Info) writeContent(line string, tsPathPrefix string) {
	line = strings.TrimSpace(line)
	if strings.HasPrefix(line, "#EXTINF:") {
		m.writeLine(line)
		return
	}
	if strings.HasSuffix(line, ".ts") {
		m.writeLine(tsPathPrefix + "/" + line)
		return
	}
}

func (m *M3u8Generator) appendLineInfo(m3u8LineInfo string, tsPathPrefix string) {
	switch m.status {
	case status_init:
		m.m3u8.writeHead()
		m.m3u8.writeContent(m3u8LineInfo, tsPathPrefix)
		m.status = status_start
	case status_start:
		m.m3u8.writeContent(m3u8LineInfo, tsPathPrefix)
	case status_end:
		return
	default:
		break
	}
	// 统计时长
	// #EXTINF:11.946000,
	//if strings.HasPrefix(info, "#EXTINF:") {
	//	extInf := strings.Split(info, ":")
	//	originDuration := extInf[1] // 单位秒
	//	parseDuartion, _ := strconv.ParseFloat(originDuration, 64)
	//	m.m3u8.duration += int(parseDuartion) * 1000
	//} else {
	//	// a2/859ebe26fe49b63145a79c820e27f522_1366a0b3-ddee-4609-98b2-cbc83e7f3872_1688119919005_dev-s_20230630101907270.ts
	//}
}

// AppendM3u8
//
//	@Description:  直接追加完整的m3u8信息
//	@receiver m m3u8生成器
//	@param m3u8Content 完整的m3u8内容
//	@param pathPrefix ts路径前缀
func (m *M3u8Generator) AppendM3u8(m3u8Content string, tsPathPrefix string) {
	m3u8Parser := NewM3u8Parser([]byte(m3u8Content), nil)
	it, hasNext := m3u8Parser.GetLineIterator()
	for hasNext {
		var m3u8Line string
		m3u8Line, hasNext = it()
		m.appendLineInfo(m3u8Line, tsPathPrefix)
	}
}

func (m *M3u8Generator) End() *M3u8Info {
	switch m.status {
	case status_init:
		m.m3u8.writeHead()
		fallthrough
	case status_start:
		m.m3u8.writeTail()
		m.status = status_end
	case status_end:
		//不处理
	default:
		//不处理
	}
	return m.m3u8
}
