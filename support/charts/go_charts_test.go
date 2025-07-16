package charts

import (
	"os"
	"testing"

	"github.com/vicanso/go-charts/v2"
)

func TestBar(t *testing.T) {
	values := [][]float64{
		{
			2.0,
			4.9,
			7.0,
			23.2,
			25.6,
			76.7,
			135.6,
			32.6,
			20.0,
			6.4,
			162.2,
			3.3,
		},
		//{
		//	1.0,
		//	1.0,
		//	8.0,
		//	0.0,
		//	0.0,
		//	0.0,
		//	0.0,
		//	0.0,
		//	8.0,
		//	8.0,
		//	1.0,
		//	1.0,
		//},
	}

	buf, err := os.ReadFile("./simhei.ttf")
	if err != nil {
		panic(err)
	}
	err = charts.InstallFont("simhei", buf)
	if err != nil {
		panic(err)
	}
	font, _ := charts.GetFont("simhei")
	charts.SetDefaultFont(font)

	p, err := charts.BarRender(
		values,
		charts.XAxisDataOptionFunc([]string{
			"Jan",
			"Feb",
			"Mar",
			"Apr",
			"May",
			"Jun",
			"Jul",
			"Aug",
			"Sep",
			"Oct",
			"Nov",
			"Dec",
		}),
		charts.TitleTextOptionFunc("降雨量统计"),
		//charts.YAxisDataOptionFunc([]string{
		//	"ml",
		//}),
		charts.LegendLabelsOptionFunc([]string{
			"降雨",
			//"蒸发",
		}, charts.PositionRight),
		//charts.MarkLineOptionFunc(0, charts.SeriesMarkDataTypeAverage),
		//charts.MarkPointOptionFunc(0, charts.SeriesMarkDataTypeMax,
		//	charts.SeriesMarkDataTypeMin),
		//func(opt *charts.ChartOption) {
		//	opt.SeriesList[1].MarkPoint = charts.NewMarkPoint(
		//		charts.SeriesMarkDataTypeMax,
		//		charts.SeriesMarkDataTypeMin,
		//	)
		//	opt.SeriesList[1].MarkLine = charts.NewMarkLine(
		//		charts.SeriesMarkDataTypeAverage,
		//	)
		//},
	)
	if err != nil {
		panic(err)
	}

	data, err := p.Bytes()
	if err != nil {
		panic(err)
	}
	// snip...
	os.WriteFile("bar.png", data, os.ModePerm)

}

func TestBar1(t *testing.T) {
	values := [][]float64{
		{
			40,
			7,
			96,
			36,
			28,
			17,
			9,
		},
	}

	buf, err := os.ReadFile("./simhei.ttf")
	if err != nil {
		panic(err)
	}
	err = charts.InstallFont("simhei", buf)
	if err != nil {
		panic(err)
	}
	font, _ := charts.GetFont("simhei")
	charts.SetDefaultFont(font)

	xAxisOption := charts.NewXAxisOption(
		[]string{
			"文件",
			"批注",
			"便签",
			"图形",
			"文本",
			"图片",
			"网页",
		},
	)
	xAxisOption.StrokeColor = charts.Color{
		R: 255,
		G: 255,
		B: 255,
		A: 255,
	}

	p, err := charts.BarRender(
		values,
		charts.XAxisOptionFunc(xAxisOption),
		charts.TitleTextOptionFunc("常用元素统计"),
		charts.WidthOptionFunc(500),
		charts.HeightOptionFunc(300),
	)
	if err != nil {
		panic(err)
	}

	data, err := p.Bytes()
	if err != nil {
		panic(err)
	}
	// snip...
	os.WriteFile("bar.png", data, os.ModePerm)

}

func TestBarChartsHelper_Save2Img(t *testing.T) {
	barChartsHelper, err := NewBarChartsHelper(&BarChartsConfig{
		Title: "内容统计",
		X: []string{
			"文本",
			"元素",
			"便签",
			"网页",
		},
		Values: [][]float64{
			{4, 1, 1, 1}, // 这是x轴上每个选项指标1的值
			{2, 1, 1, 2}, // 这是x轴上每个选项指标2的值
		},
		Legend:    []string{"指标1", "指标2"},
		BarWidth:  20,
		IsInteger: true,
	})
	if err != nil {
		panic(err)
	}
	barChartsHelper.Save2Img("testBar.png")
}
