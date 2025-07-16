package docx

import (
	"github.com/fumiama/go-docx"
)

const (
	ParagraphTypeText  = 1 // 文本
	ParagraphTypeImage = 2 // 图片
	ParagraphTypeTable = 3 // 表格
)

const (
	FontSizeSmallSix   = "13" // 小六
	FontSizeSix        = "15" // 六号
	FontSizeSmallFive  = "18" // 小五
	FontSizeFive       = "21" // 五号
	FontSizeSmallFour  = "24" // 小四
	FontSizeFour       = "28" // 四号
	FontSizeSmallThree = "30" // 小三
	FontSizeThree      = "32" // 三号
	FontSizeSmallTwo   = "36" // 小二
	FontSizeTwo        = "44" // 二号
	FontSizeSmallOne   = "48" // 小一
	FontSizeOne        = "52" // 一号
	FontSizeSmallFirst = "72" // 小初
	FontSizeFirst      = "84" // 初号
)

const (
	JustificationStart      = "start"      // 左对齐
	JustificationCenter     = "center"     // 居中对齐
	JustificationEnd        = "end"        // 右对齐
	JustificationBoth       = "both"       // 两端对齐
	JustificationDistribute = "distribute" // 分散对齐
)

type Paragraph interface {
	init(options ...any) Paragraph
	add(helper *Helper) error
}

func newParagraph(t int) Paragraph {
	switch t {
	case ParagraphTypeText:
		return &TextParagraph{}
	case ParagraphTypeImage:
		return &ImageParagraph{}
	case ParagraphTypeTable:
		return &TableParagraph{}

	default:
		return nil
	}
}

// TextConfig 文字配置
type TextConfig struct {
	FontSize      string // 字体大小
	Text          string // 文字内容
	Italic        bool   // 斜体
	Bold          bool   // 加粗
	Color         string // 颜色 rgb 000000
	ShadeColor    string // 底纹颜色
	Justification string // 对齐方式
}

type TextParagraph struct {
	withLineBreak   bool // TextList 之间是否换行
	TextList        []*TextConfig
	DefaultFontSize string // 默认字体大小
}

func (t *TextParagraph) init(options ...any) Paragraph {
	t.TextList = options[0].([]*TextConfig)
	t.DefaultFontSize = options[1].(string)
	t.withLineBreak = options[2].(bool)
	return t
}

func (t *TextParagraph) add(helper *Helper) error {
	word := helper.docx
	var paragraph *docx.Paragraph
	// 如果无需换行, 使用同一个段落对象
	if !t.withLineBreak {
		var justification string
		if len(t.TextList) > 0 && len(t.TextList[0].Justification) > 0 {
			justification = t.TextList[0].Justification
		} else {
			justification = JustificationStart
		}
		paragraph = word.AddParagraph().Justification(justification)
	}
	for _, text := range t.TextList {
		// 如果需要换行, 每个text使用一个段落对象
		if t.withLineBreak || paragraph == nil {
			var justification string
			if len(text.Justification) > 0 {
				justification = text.Justification
			} else {
				justification = JustificationStart
			}
			paragraph = word.AddParagraph().Justification(justification)
		}
		fontSize := text.FontSize
		if fontSize == "" {
			fontSize = t.DefaultFontSize
		}
		// 添加文字并指定大小
		line := paragraph.AddText(text.Text).Size(fontSize)
		helper.setGlobalSetting(line)
		if text.Italic {
			line.Italic()
		}
		if text.Bold {
			line.Bold()
		}
		if text.Color != "" {
			line.Color(text.Color)
		}
		if text.ShadeColor != "" {
			line.Shade("clear", "auto", text.ShadeColor)
		}
	}
	return nil
}

type ImageParagraph struct {
	withLineBreak bool     // 图片之间是否换行
	ImagePath     []string // 图片地址
	w, h          int64    // 图片最大的宽高 单位从cm 转换到 English Metric Units, 1cm = 360000 English Metric Units, 如果某个图片超过了指定的大小, 则等比例缩放
}

func (i *ImageParagraph) init(options ...any) Paragraph {
	i.ImagePath = options[0].([]string)
	w := options[1].(float64)
	h := options[2].(float64)
	i.w = int64(w * 360000)
	i.h = int64(h * 360000)
	return i
}

func (i *ImageParagraph) add(helper *Helper) error {
	word := helper.docx
	var paragraph *docx.Paragraph
	// 如果无需换行, 使用同一个段落对象
	if !i.withLineBreak {
		paragraph = word.AddParagraph().Justification(JustificationStart)
	}
	// 添加图片
	for _, imgPath := range i.ImagePath {
		// 如果需要换行, 每个图片使用一个段落对象
		if i.withLineBreak || paragraph == nil {
			paragraph = word.AddParagraph().Justification(JustificationStart)
		}
		img, err := paragraph.AddInlineDrawingFrom(imgPath)
		if err != nil {
			return err
		}

		if i.w == 0 || i.h == 0 {
			continue
		}

		imgW := img.Children[0].(*docx.Drawing).Inline.Extent.CX
		imgY := img.Children[0].(*docx.Drawing).Inline.Extent.CY

		// 等比例缩放
		if imgW > i.w || imgY > i.h {
			ratio := float64(imgW) / float64(imgY)
			if imgW > i.w {
				imgW = i.w
				imgY = int64(float64(i.w) / ratio)
			}
			if imgY > i.h {
				imgY = i.h
				imgW = int64(float64(i.h) * ratio)
			}
			img.Children[0].(*docx.Drawing).Inline.Extent.CX = imgW
			img.Children[0].(*docx.Drawing).Inline.Extent.CY = imgY
		}

	}

	return nil
}
