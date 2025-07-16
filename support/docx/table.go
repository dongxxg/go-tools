package docx

import (
	"github.com/fumiama/go-docx"
)

var (
	// 透明边框相关
	transparent = &docx.WTableBorder{
		Val: "nil",
	}
	TransparentBorders = &docx.WTableBorders{
		Top:     transparent,
		Left:    transparent,
		Bottom:  transparent,
		Right:   transparent,
		InsideH: transparent,
		InsideV: transparent,
	}
)

type TableParagraph struct {
	Title           []string    // 表头
	TitleConfig     *TextConfig // 表头样式
	Content         [][]string  // 内容
	DefaultFontSize string      // 默认字体大小
	IsTransparent   bool        // 透明边框
	Widths          []float64   // 列宽的比例
	BorderColor     string      // 边框颜色
}

func (t *TableParagraph) init(options ...any) Paragraph {
	t.Title = options[0].([]string)
	t.TitleConfig = options[1].(*TextConfig)
	t.Content = options[2].([][]string)
	t.DefaultFontSize = options[3].(string)
	t.IsTransparent = options[4].(bool)
	t.Widths = options[5].([]float64)
	t.BorderColor = options[6].(string)

	if t.DefaultFontSize == "" {
		t.DefaultFontSize = FontSizeSmallFour
	}
	return t
}

func (t *TableParagraph) add(helper *Helper) error {
	word := helper.docx
	var colCount, rowCount int
	if t.Title == nil {
		rowCount = len(t.Content)
		if len(t.Content) > 0 {
			colCount = len(t.Content[0])
		}
	} else {
		rowCount = len(t.Content) + 1
		colCount = len(t.Title)
	}

	rowHeights := make([]int64, rowCount)
	colWidths := make([]int64, colCount)

	if len(t.Widths) > 0 {
		for i, width := range t.Widths {
			colWidths[i] = int64(width * 8100)
		}
	}
	for i := range rowHeights {
		rowHeights[i] = 360 // 360是试出来的, 目前可以在小五字体下, 表格的宽度近似于1.5倍行距, 因为没有找到可以直接设置行距的方法
	}

	// 8100宽度是试出来的, 基本上可以撑满a4纸的宽度, 目前写死
	table := word.AddTableTwips(rowHeights, colWidths, 8100, nil).
		Justification("center")

	var defaultBorderColor *docx.WTableBorders
	if t.BorderColor != "" {
		defaultBorderColor = &docx.WTableBorders{
			Top:     &docx.WTableBorder{Val: "single", Color: t.BorderColor},
			Left:    &docx.WTableBorder{Val: "single", Color: t.BorderColor},
			Bottom:  &docx.WTableBorder{Val: "single", Color: t.BorderColor},
			Right:   &docx.WTableBorder{Val: "single", Color: t.BorderColor},
			InsideH: &docx.WTableBorder{Val: "single", Color: t.BorderColor},
			InsideV: &docx.WTableBorder{Val: "single", Color: t.BorderColor},
		}
	}

	rowIndex := 0

	// 填充标题
	for i, titleText := range t.Title {
		titleColParagraph := table.TableRows[rowIndex].TableCells[i].AddParagraph().
			AddText(titleText)
		helper.setGlobalSetting(titleColParagraph)

		if t.TitleConfig.FontSize != "" {
			titleColParagraph.Size(t.TitleConfig.FontSize)
		} else {
			titleColParagraph.Size(t.DefaultFontSize)
		}

		if t.TitleConfig.Italic {
			titleColParagraph.Italic()
		}
		if t.TitleConfig.Bold {
			titleColParagraph.Bold()
		}
		if t.TitleConfig.Color != "" {
			titleColParagraph.Color(t.TitleConfig.Color)
		}
		if t.TitleConfig.ShadeColor != "" {
			table.TableRows[rowIndex].TableCells[i].Shade("clear", "auto", t.TitleConfig.ShadeColor)
		}

		if defaultBorderColor != nil {
			table.TableRows[rowIndex].TableCells[i].TableCellProperties.TableBorders = defaultBorderColor
		}
		if t.IsTransparent {
			table.TableRows[rowIndex].TableCells[i].TableCellProperties.TableBorders = TransparentBorders
		}
	}
	if len(t.Title) > 0 {
		rowIndex++
	}

	// 填充内容
	for _, ContentCol := range t.Content {
		for colIndex, ContentText := range ContentCol {
			run := table.TableRows[rowIndex].TableCells[colIndex].AddParagraph().
				AddText(ContentText).
				Size(t.DefaultFontSize)
			helper.setGlobalSetting(run)
			if defaultBorderColor != nil {
				table.TableRows[rowIndex].TableCells[colIndex].TableCellProperties.TableBorders = defaultBorderColor
			}
			if t.IsTransparent {
				table.TableRows[rowIndex].TableCells[colIndex].TableCellProperties.TableBorders = TransparentBorders
			}
		}
		rowIndex++
	}
	return nil
}
