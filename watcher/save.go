package watcher

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	chart "github.com/wcharczuk/go-chart"
)

func saveRecords(rec Records, name string) {
	records, err := json.Marshal(rec)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Couldn't marshal records to json: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(name)
	err = ioutil.WriteFile(fmt.Sprintf("%s.json", name), records, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Couldn't save to file: %v\n", err)
		os.Exit(1)
	}
}

func plotRecords(rec Records, name string) {
	graph := chart.Chart{
		XAxis: chart.XAxis{
			Name:           "Time",
			Style:          chart.StyleShow(),
			ValueFormatter: chart.TimeMinuteValueFormatter,
		},
		YAxis: chart.YAxis{
			Name:           "Metric in %",
			Style:          chart.StyleShow(),
			ValueFormatter: chart.PercentValueFormatter,
		},
		Series: []chart.Series{
			chart.TimeSeries{
				XValues: rec.Time,
				YValues: rec.CPU,
				Name:    "CPU",
				Style: chart.Style{
					Show:        true,
					StrokeColor: chart.ColorBlue,
					FillColor:   chart.ColorBlue.WithAlpha(100),
				},
			},
			chart.TimeSeries{
				XValues: rec.Time,
				YValues: rec.Mem,
				Name:    "Memory",
				Style: chart.Style{
					Show:        true,
					StrokeColor: chart.ColorRed,
					FillColor:   chart.ColorRed.WithAlpha(100),
				},
			},
		},
	}

	graph.Elements = []chart.Renderable{
		chart.Legend(&graph),
	}

	img := bytes.NewBuffer([]byte{})
	err := graph.Render(chart.PNG, img)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Couldn't render graph: %v\n", err)
		os.Exit(1)
	}
	err = ioutil.WriteFile(fmt.Sprintf("%s.png", name), img.Bytes(), 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Couldn't save graph to file: %v\n", err)
		os.Exit(1)
	}
}

func save(rec Records, output string, json bool) {
	if json {
		saveRecords(rec, output)
	} else {
		plotRecords(rec, output)
	}
	os.Exit(0)
}
