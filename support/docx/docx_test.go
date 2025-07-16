package docx

import (
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/fumiama/go-docx"
)

func TestDocx(t *testing.T) {
	w := docx.New().WithDefaultTheme()
	// add new paragraph
	// add text

	size := 1

	for i := 0; i < size; i++ {

		// 创建一个新段落
		paragraph := w.AddParagraph().Justification("start")
		// 添加文字并指定大小
		paragraph.AddText("老师" + strconv.Itoa(i) + "：").Size("54")
		paragraph1 := w.AddParagraph().Justification("start")
		paragraph1.AddText("这是一段长点评这是一段长点评这是一段长点评这是一段长点评这是一段长点评这是一段长点评这是一段长点评这是一段长点评这是一段长点评这是一段长点评").
			Size("44")
		paragraph1.AddText("这是一段长点评这是一段长点评这是一段长点评这是一段长点评这是一段长点评这是一段长点评这是一段长点评这是一段长点评这是一段长点评这是一段长点评").
			Size("44")

		// 创建一个和文字排布在一起的图片, 而不是浮于文字之上的, 浮于文字之上的使用 AddAnchorDrawingFrom
		img2, _ := paragraph1.AddInlineDrawingFrom("./hollow_knight.jpg")

		// cx,cy的单位是 English Metric Units, 等于 1/914400 英寸, 所以 1 英寸=2.54 CM=914400 English Metric Units
		// 所以 1 CM=360000 English Metric Units😓, 秦始皇呢, 统一度量衡的时候怎么把大英漏了
		fmt.Println("w:", img2.Children[0].(*docx.Drawing).Inline.Extent.CX)
		fmt.Println("w:", img2.Children[0].(*docx.Drawing).Inline.Extent.CX/360000, "cm")
		fmt.Println("h:", img2.Children[0].(*docx.Drawing).Inline.Extent.CY)
		fmt.Println("h:", img2.Children[0].(*docx.Drawing).Inline.Extent.CY/360000, "cm")
		// 指定图片的大小, 单位是这个 English Metric Units
		img2.Children[0].(*docx.Drawing).Inline.Size(
			img2.Children[0].(*docx.Drawing).Inline.Extent.CX,
			img2.Children[0].(*docx.Drawing).Inline.Extent.CY,
		)

		// 创建一个和文字排布在一起的图片, 而不是浮于文字之上的, 浮于文字之上的使用 AddAnchorDrawingFrom
		img3, _ := paragraph1.AddInlineDrawingFrom("./hollow_knight.jpg")

		// cx,cy的单位是 English Metric Units, 等于 1/914400 英寸, 所以 1 英寸=2.54 CM=914400 English Metric Units
		// 所以 1 CM=360000 English Metric Units😓, 秦始皇呢, 统一度量衡的时候怎么把大英漏了
		fmt.Println("w:", img3.Children[0].(*docx.Drawing).Inline.Extent.CX)
		fmt.Println("h:", img3.Children[0].(*docx.Drawing).Inline.Extent.CY)
		// 指定图片的大小, 单位是这个 English Metric Units
		img3.Children[0].(*docx.Drawing).Inline.Size(
			360000*2,
			360000*2,
		)

		paragraph.AddText("这是一段长点评这是一段长点评这是一段长点评这是一段长点评这是一段长点评这是一段长点评这是一段长点评这是一段长点评这是一段长点评这是一段长点评").
			Size(FontSizeSmallFour).Font("黑体", "黑体", "黑体", "eastAsia")

	}
	w.AddParagraph()
	w.AddParagraph()

	// 创建一个表格
	tbl1 := w.AddTable(3, 3, 8250, nil).
		Justification("center")

	tbl1.TableRows[0].Justification("center")
	tbl1.TableRows[0].TableCells[0].TableCellProperties.VAlign = &docx.WVerticalAlignment{
		Val: "center",
	}
	tbl1.TableRows[0].TableCells[0].AddParagraph().Justification("center").AddText("文件名称").Bold()
	tbl1.TableRows[0].TableCells[1].AddParagraph().Justification("center").AddText("上传人").Bold()
	tbl1.TableRows[0].TableCells[2].AddParagraph().Justification("center").AddText("上传大小").Bold()

	tbl1.TableRows[1].Justification("center")
	tbl1.TableRows[1].TableCells[0].TableCellProperties.VAlign = &docx.WVerticalAlignment{
		Val: "center",
	}
	tbl1.TableRows[1].TableCells[0].TableCellProperties.TableBorders = &docx.WTableBorders{
		Left: &docx.WTableBorder{
			Val: "nil",
		},
	}
	tbl1.TableRows[1].TableCells[0].AddParagraph().
		Justification("center").
		AddText("第2节+第1课时+电解质的电离（同步课件）.pptx")
	tbl1.TableRows[1].TableCells[1].AddParagraph().Justification("center").AddText("詹老师")
	tbl1.TableRows[1].TableCells[2].AddParagraph().Justification("center").AddText("50M")

	tbl1.TableRows[2].Justification("center")
	tbl1.TableRows[2].TableCells[0].TableCellProperties.VAlign = &docx.WVerticalAlignment{
		Val: "center",
	}
	tbl1.TableRows[2].TableCells[0].AddParagraph().
		Justification("center").
		AddText("电解质的电离（同步课件）.pptx")
	tbl1.TableRows[2].TableCells[1].AddParagraph().Justification("center").AddText("张老师")
	tbl1.TableRows[2].TableCells[2].AddParagraph().Justification("center").AddText("1M")

	w.AddParagraph()

	// 创建一个表格
	tbl2 := w.AddTable(8, 8, 8100, nil).
		Justification("center")
	for i, r := range tbl2.TableRows {
		r.Justification("center")
		for j, c := range r.TableCells {
			c.TableCellProperties.VAlign = &docx.WVerticalAlignment{Val: "center"}
			if i == 0 {
				c.AddParagraph().Justification("center").AddText("坐标").Bold()

			} else {
				c.AddParagraph().Justification("center").AddText(fmt.Sprintf("(%d, %d)", i, j))
			}
		}
	}
	tbl2.TableRows[0].TableCells[0].Shade("clear", "auto", "E7E6E6")

	f, err := os.Create("generated.docx")
	// save to file
	if err != nil {
		panic(err)
	}
	_, err = w.WriteTo(f)
	if err != nil {
		panic(err)
	}
	err = f.Close()
	if err != nil {
		panic(err)
	}

}

func TestDocxHelper(t *testing.T) {
	helper := NewDocxHelper()
	helper.AddTextListParagraph(
		[]*TextConfig{
			{
				FontSize: FontSizeSmallFour,
				Text:     "老师1：",
				Italic:   true,
				Bold:     true,
			},
			{
				FontSize: FontSizeFive,
				Text:     "加粗点评内容",
				Bold:     true,
			},
			{
				Text: "普通点评内容",
			},
			{
				Text:   "斜体点评内容",
				Italic: true,
			},
		}, FontSizeFive, true)

	helper.AddImageParagraph(
		[]string{"./hollow_knight.jpg", "./hollow_knight.jpg", "./hollow_knight.jpg"},
		4.6,
		2.6,
		false)

	helper.AddTable(
		[]string{"文件名称", "上传人", "上传大小"},
		&TextConfig{
			Italic:     true,
			Bold:       true,
			ShadeColor: "E7E6E6",
		},
		[][]string{
			{"课件1", "老师1", "40M"},
			{"课件2", "老师2", "60M"},
		},
		FontSizeSmallFour,
		true,
		[]float64{},
		"BFBFBF",
	)

	helper.AddTextListParagraph(
		[]*TextConfig{
			{
				Text: "老师1：\t234234234234234234\tfsfdkgjkjkf\n加粗点评内容\tv此方法几块几块十几块的\t厂家反馈什么方面更加",
			},
		}, FontSizeFive, true)

	helper.WriteDocx("generatedHelper.docx")
}
