package docx

import (
	"os"

	"github.com/fumiama/go-docx"
)

type Font int

const (
	Song Font = iota
	Hei
)

type Helper struct {
	DefaultTextColor string
	DefaultFont      Font
	docx             *docx.Docx
}

func NewDocxHelper() *Helper {
	w := docx.New().WithDefaultTheme()
	return &Helper{
		docx:        w,
		DefaultFont: Song,
	}
}

// setGlobalSetting 应用全局样式
func (h *Helper) setGlobalSetting(run *docx.Run) *docx.Run {
	h.setFont(run)
	h.setColor(run)
	return run
}

// SetGlobalDefaultFont 设置全局默认字体, 目前只支持宋体和黑体
func (h *Helper) SetGlobalDefaultFont(font Font) {
	h.DefaultFont = font
}

// setFont 设置字体
func (h *Helper) setFont(run *docx.Run) *docx.Run {
	switch h.DefaultFont {
	case Song:
		run.Font("宋体", "宋体", "宋体", "eastAsia")
	case Hei:
		run.Font("黑体", "黑体", "黑体", "eastAsia")
	default:
		run.Font("宋体", "宋体", "宋体", "eastAsia")
	}
	return run
}

// SetGlobalDefaultColor 设置全局文字默认颜色
func (h *Helper) SetGlobalDefaultColor(rgb string) {
	h.DefaultTextColor = rgb
}

func (h *Helper) setColor(run *docx.Run) *docx.Run {
	if h.DefaultTextColor != "" {
		run.Color(h.DefaultTextColor)
	}
	return run
}

// AddSimpleTextLine 添加一行文本, 仅指定字号(换行)
// defaultFontSize 默认字体大小
func (h *Helper) AddSimpleTextLine(
	text string,
	defaultFontSize string,
) error {
	p := newParagraph(ParagraphTypeText).init([]*TextConfig{{Text: text}}, defaultFontSize, false)
	return p.add(h)
}

// AddEmptyLine 添加一行空行
// defaultFontSize 默认字体大小
func (h *Helper) AddEmptyLine(
	defaultFontSize string,
) error {
	p := newParagraph(ParagraphTypeText).init([]*TextConfig{}, defaultFontSize, false)
	return p.add(h)
}

// AddTextParagraphWithConfig 添加文本段落, 支持设置字体简单样式, 字体大小, 斜体, 加粗等
// textWithConfig *TextConfig 文本数组, 支持设置字体简单样式, 字体大小, 斜体, 加粗等
func (h *Helper) AddTextParagraphWithConfig(
	textWithConfig *TextConfig,
) error {
	p := newParagraph(
		ParagraphTypeText,
	).init([]*TextConfig{textWithConfig}, textWithConfig.FontSize, false)
	return p.add(h)
}

// AddTextListParagraph 添加文本段落, 支持数组
// textWithConfig []*TextConfig 文本数组, 支持设置字体简单样式, 字体大小, 斜体, 加粗等
// defaultFontSize 默认字体大小
// withLineBreak 表示数组之间会不会换行
func (h *Helper) AddTextListParagraph(
	textWithConfig []*TextConfig,
	defaultFontSize string,
	withLineBreak bool,
) error {
	p := newParagraph(ParagraphTypeText).init(textWithConfig, defaultFontSize, withLineBreak)
	return p.add(h)
}

// AddImageParagraph 添加图片段落, 支持string数组, 将所有图片按照数组顺序添加到docx中, 图片会按照maxW,maxH进行等比缩放
// maxW, maxH 为图片最大宽高, 任意一个为0表示不缩放
// withLineBreak 表示图片之间会不会换行
func (h *Helper) AddImageParagraph(path []string, maxW, maxH float64, withLineBreak bool) error {
	p := newParagraph(ParagraphTypeImage).init(path, maxW, maxH, withLineBreak)
	return p.add(h)
}

// AddTable 添加表格, 支持设置标题, 标题样式, 内容, 默认字体大小
// title []string 标题数组, 支持设置字体简单样式, 字体大小, 斜体, 加粗等
// titleConfig *TextConfig 标题样式, 支持设置字体简单样式, 字体大小, 斜体, 加粗等
// content [][]string 内容数组, 默认字体, 不设置样式
// defaultFontSize 默认字体大小
// isTransparent 表示表格边框是否透明
// widths []float64 表示表格列宽, 百分比
// borderColor 表示表格边框颜色
func (h *Helper) AddTable(
	title []string,
	titleConfig *TextConfig,
	content [][]string,
	defaultFontSize string,
	isTransparent bool,
	widths []float64,
	borderColor string,
) error {
	p := newParagraph(
		ParagraphTypeTable,
	).init(title, titleConfig, content, defaultFontSize, isTransparent, widths, borderColor)
	return p.add(h)
}

func (h *Helper) WriteDocx(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = h.docx.WriteTo(f)
	if err != nil {
		return err
	}
	return nil
}
