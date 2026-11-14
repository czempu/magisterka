package charts

import (
    "github.com/go-echarts/go-echarts/v2/charts"
    "github.com/go-echarts/go-echarts/v2/opts"

)

func PieEmployees(total, active int) *charts.Pie {
    inactive := total - active

    pie := charts.NewPie()
    pie.SetGlobalOptions(
        charts.WithTitleOpts(opts.Title{
            Title:    "Pracownicy",
            Subtitle: "Obecni / Nieobecni",
        }),
    )

    pie.AddSeries("Pracownicy", []opts.PieData{
        {Name: "Obecni", Value: active},
        {Name: "Nieobecni", Value: inactive},
    }).SetSeriesOptions(
        charts.WithLabelOpts(opts.Label{
            Show:      opts.Bool(true),
	    Formatter: "{b}: {c}",
        }),
        charts.WithPieChartOpts(opts.PieChart{
            Radius: []string{"40%", "70%"},
        }),
    )

    return pie
}

