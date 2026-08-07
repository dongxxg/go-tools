package charts

import (
	_ "embed"
	"errors"
	"fmt"
	"math"
	"os"
	"sync"

	"github.com/vicanso/go-charts/v2"
)

// 常用的图表绘制库:
// go-echarts 优点：可以完美绘制前端echarts的图表, 缺点：只能生成html, html转png需要使用chromedp加载网页并截图, 效率较低, 代价较大
// go-charts 优点：绘制效率高, 可以直接生成图片,  缺点：自定义能力较少, 绘制相对不够美观
// 最终选用go-charts, 实现简单, 效率较高

// 基于 go-charts 的简单封装, 目前只实现了柱状图的绘制
// go-charts 的文档: https://github.com/vicanso/go-charts/blob/main/README_zh.md
var (
	initFlag bool
	once     = sync.Once{}
	//go:embed simhei.ttf
	font                       []byte
	ErrXAndValuesNotMatch      = errors.New("x and values not match")
	ErrLegendAndValuesNotMatch = errors.New("legend and values not match")
	ErrChartsHelperIsNil       = errors.New("charts helper is nil")
)

// InitCharsHelper 主要是初始化默认字体, 使用once保证只初始化一次
func initCharsHelper() {
	if initFlag {
		return
	}
	once.Do(func() {
		err := charts.InstallFont("simhei", font)
		if err != nil {
			panic(err)
		}
		font, _ := charts.GetFont("simhei")
		charts.SetDefaultFont(font)
		initFlag = true
	})
}

type BarChartsConfig struct {
	Title     string      // 图表的标题
	X         []string    // x轴的选项
	Values    [][]float64 // x轴选项的值, 一个x可能有多个值, 如每个月份的降雨和蒸发量,每个指标需要放在一组, 例如 {{1,2,3},{,2,3,4}}, 第一组是每个月份的降雨量, 第二组是每个月份的蒸发量, 详情查看示例 TestBarChartsHelper_Save2Img
	Legend    []string    // 图例的名称, 如果一个x轴的选项有多个值, 需要和图例的顺序保持一致, 如果设置了图例, 那么需要和每组的值数量保持一致
	BarWidth  int         // 单个柱状图的宽度
	IsInteger bool        // y轴是否是整数
}

// BarChartsHelper 柱状图图表辅助类
type BarChartsHelper struct {
	*BarChartsConfig
}

func (c *BarChartsConfig) preCheck() error {
	// 图例有值的情况下, 检查图例的数量和值数量是否一致
	if len(c.Legend) != 0 && len(c.Legend) != len(c.Values) {
		return ErrLegendAndValuesNotMatch
	}
	// 检查x轴选项的数量和值数量是否一致
	for _, value := range c.Values {
		if len(value) != len(c.X) {
			return ErrXAndValuesNotMatch
		}
	}
	return nil
}

func NewBarChartsHelper(
	barConfig *BarChartsConfig,
) (*BarChartsHelper, error) {

	initCharsHelper()

	err := barConfig.preCheck()

	if err != nil {
		return nil, err
	}

	return &BarChartsHelper{
		barConfig,
	}, nil
}

// Save2Img 保存图表为图片, 默认情况下仅支持png格式
func (c *BarChartsHelper) Save2Img(imgPath string) error {
	if c == nil {
		return ErrChartsHelperIsNil
	}

	var formater = func(f float64) string {
		return fmt.Sprintf("%.2f", f)
	}
	var yOpt charts.YAxisOption
	if c.IsInteger {
		var maxValue float64
		var divideCount int

		for i := range c.Values {
			for j := range c.Values[i] {
				maxValue = math.Max(maxValue, c.Values[i][j])
			}
		}
		maxValue, divideCount = findMaxAndDivideCount(int(maxValue))

		yOpt = charts.YAxisOption{
			Max:         &maxValue,
			DivideCount: divideCount,
		}
		formater = func(f float64) string {
			return fmt.Sprintf("%.0f", f)
		}
	}

	var opts []charts.OptionFunc
	opts = append(opts,
		charts.XAxisDataOptionFunc(c.X),
		charts.TitleTextOptionFunc(c.Title),
		charts.LegendLabelsOptionFunc(c.Legend, charts.PositionRight),
		charts.YAxisOptionFunc(yOpt),
	)

	seriesList := charts.NewSeriesListDataFromValues(c.Values, charts.ChartTypeBar)
	p, err := charts.Render(charts.ChartOption{
		SeriesList:     seriesList,
		BarWidth:       c.BarWidth, // 设置柱状图的宽度
		ValueFormatter: formater,   // 设置y轴的格式化函数
	}, opts...)

	if err != nil {
		return err
	}

	data, err := p.Bytes()
	if err != nil {
		return err
	}
	return os.WriteFile(imgPath, data, os.ModePerm)

}

// findMaxAndDivideCount 寻找y轴maxValue和分割y轴的数量
// 这里是为了保证: 数据只有整数的情况, 使用工具默认的绘制方案, y轴坐标可能会出现浮点数, 所以这里计算出y轴的DivideCount, 保证y轴一定是整数
func findMaxAndDivideCount(maxValue int) (max float64, divideCount int) {

	if maxValue <= 3 {
		return 3, 3
	}

	// 从高到低遍历9到4，寻找能整除maxValue的数
	for n := 9; n > 3; n-- {
		if maxValue%n == 0 {
			divideCount = n
			max = float64(maxValue)
		}
	}
	if max == 0 {
		// 如果没有找到能整除maxValue的数, 把maxValue+1, 递归寻找
		return findMaxAndDivideCount(maxValue + 1)

	}
	return max, divideCount
}
