package gocodefuncs

import (
	"fmt"
	"github.com/LubyRuffy/goflow/utils"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/mitchellh/mapstructure"
	"github.com/tidwall/gjson"
	"os"
	"path/filepath"
)

// BarChart 生成bar类型报表
func BarChart(p Runner, params map[string]interface{}) *FuncResult {
	var err error
	var options pieChartParams
	if err = mapstructure.Decode(params, &options); err != nil {
		panic(err)
	}

	barData := make(map[string]int64)
	err = utils.EachLineWithContext(p.GetContext(), p.GetLastFile(), func(line string) error {
		name := gjson.Get(line, options.Name)
		if !name.Exists() {
			return nil
			//return fmt.Errorf(`pie chart data is invalid: %s is needed`, options.Name)
		}

		if options.Value != "count()" {
			value := gjson.Get(line, options.Value)
			if !value.Exists() {
				return fmt.Errorf(`pie chart data is invalid: %s is needed`, options.Value)
			}
			barData[name.String()] = value.Int() + barData[name.String()]
		} else {
			barData[name.String()] = barData[name.String()] + 1
		}

		return nil
	})
	if err != nil {
		panic(err)
	}

	var barItems []opts.BarData
	var keys []string
	for _, i := range utils.TopMapByValue(barData, options.Size) {
		keys = append(keys, i.Name)
		barItems = append(barItems, opts.BarData{Name: i.Name, Value: i.Value})
	}

	chart := charts.NewBar()
	chart.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: options.Title, Left: "center"}),
		charts.WithTooltipOpts(opts.Tooltip{Show: true}),
		charts.WithInitializationOpts(opts.Initialization{
			Width:      "640px",
			Height:     "480px",
			PageTitle:  options.Title,
			AssetsHost: ChartAssetsHost,
		}),
		//charts.WithInitializationOpts(opts.Initialization{AssetsHost: ChartAssetsHost}),
	)
	chart.SetXAxis(keys).AddSeries(options.Name, barItems)

	f, err := utils.WriteTempFile(".html", func(f *os.File) error {
		return chart.Render(f)
	})

	if err != nil {
		panic(fmt.Errorf("generateChart error: %w", err))
	}

	return &FuncResult{
		Artifacts: []*Artifact{{
			FilePath: f,
			FileName: filepath.Base(f),
			FileType: "chart_html",
		}},
	}
}

/*
- 如果name，value字段都存在，但是name对应多条记录，value不一样，是取哪一条合适？还是加起来合适？目前选择加起来
*/
