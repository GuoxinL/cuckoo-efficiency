/*
   Created by guoxin in 2022/8/31 3:21 PM
*/
package report

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/types"
)

// Report Used to generate charts, This method can be used to produce beautiful diagrams for stress testing
func Report(title, subtitle string, xAxis interface{}, series ...Series) *charts.Line {
	// create a new bar instance
	line := charts.NewLine()
	line.SetGlobalOptions(charts.WithLegendOpts(opts.Legend{
		Bottom: "5px",
		TextStyle: &opts.TextStyle{
			Color: "#eee",
		},
		SelectedMode: "multiple",
	}),
		charts.WithToolboxOpts(opts.Toolbox{Show: true}),
		charts.WithLegendOpts(opts.Legend{
			Show:          true,
			Left:          "",
			Top:           "",
			Right:         "",
			Bottom:        "",
			Data:          nil,
			Orient:        "",
			InactiveColor: "",
			Selected:      nil,
			SelectedMode:  "",
			Padding:       5,
			ItemWidth:     0,
			ItemHeight:    0,
			X:             "",
			Y:             "",
			Width:         "",
			Height:        "",
			Align:         "",
			TextStyle:     nil,
		}),
		charts.WithInitializationOpts(opts.Initialization{Theme: types.ThemeWesteros}),
		charts.WithTitleOpts(opts.Title{
			Title:    title,
			Subtitle: subtitle,
		}),
	)

	// set some global options like Title/Legend/ToolTip or anything else
	line.SetGlobalOptions()

	axis := line.SetXAxis(xAxis)
	for i := range series {
		var data []opts.LineData
		for j := 0; j < len(series[i].Data); j++ {
			data = append(data, opts.LineData{Value: series[i].Data[j]})
		}
		axis.AddSeries(series[i].Name, data, series[i].Options...)
	}
	axis.SetSeriesOptions(charts.WithLineChartOpts(opts.LineChart{Smooth: true}))

	return axis
}

type Series struct {
	Name    string
	Data    []interface{}
	Options []charts.SeriesOpts
}

func getFilename(title, subtitle string) string {
	filename := title + " " + subtitle
	filename = strings.Trim(filename, " ")
	filename = strings.ToLower(filename)
	filename = strings.Replace(filename, " ", "_", -1)
	filename = strings.Replace(filename, ",", "_", -1)
	filename = strings.Replace(filename, ":", "_", -1)
	return filename
}

func GeneratePages(title, path string, charts ...components.Charter) {
	page := components.NewPage()
	page.PageTitle = title
	page.AddCharts(charts...)
	page.Layout = components.PageFlexLayout
	err := createDirIfNotExist(path)
	if err != nil {
		panic(err)
	}
	path = filepath.Join(path, getFilename(title, "")+".html")
	f, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	_ = page.Render(io.MultiWriter(f))
}

func createDirIfNotExist(path string) error {
	_, err := os.Stat(path)
	if err == nil {
		return nil
	}
	if os.IsNotExist(err) {
		// create dir
		err := os.MkdirAll(path, os.ModePerm)
		if err != nil {
			return err
		}
	}
	return nil
}
