package watcher

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/shirou/gopsutil/process"
	chart "github.com/wcharczuk/go-chart"
)

type Records struct {
	Time []time.Time `json:"times"`
	CPU  []float64   `json:"cpu"`
	Mem  []float64   `json:"mem"`
}

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
			ValueFormatter: chart.TimeHourValueFormatter,
		},
		YAxis: chart.YAxis{
			Name:           "Metric in %",
			Style:          chart.StyleShow(),
			ValueFormatter: chart.FloatValueFormatter,
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

func getProcessStats(p *process.Process, noChildren bool) (float64, float64) {
	cpu, err := p.CPUPercent()

	if err != nil {
		fmt.Fprintf(os.Stderr, "Couldn't get process cpu usage: %v\n", err)
		os.Exit(1)
	}

	mem, err := p.MemoryPercent()

	if err != nil {
		fmt.Fprintf(os.Stderr, "Couldn't marshal process memory usage: %v\n", err)
		os.Exit(1)
	}

	mem64 := float64(mem)

	if !noChildren {
		childs, err := p.Children()

		if err != nil {
			fmt.Fprintf(os.Stderr, "Couldn't get process children: %v\n", err)
			os.Exit(1)
		}

		for _, child := range childs {
			c, _ := child.CPUPercent()
			m, _ := child.MemoryPercent()
			cpu += c
			mem64 += float64(m)
		}
	}

	return cpu, mem64
}

func save(rec Records, output string, json bool) {
	if json {
		saveRecords(rec, output)
	} else {
		plotRecords(rec, output)
	}
	os.Exit(0)
}

func Watch(pid int, interval time.Duration, duration time.Duration, noChildren bool, output string, json bool) {

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGINT)

	proc, err := process.NewProcess(int32(pid))

	if err != nil {
		fmt.Fprintf(os.Stderr, "Couldn't find process for pid %v: %v\n", pid, err)
		os.Exit(1)
	}

	rand.Seed(int64(0))

	ticker := time.NewTicker(interval)

	times := []time.Time{}
	cpuRecords := []float64{}
	memRecords := []float64{}
	go func() {

		for {
			select {
			case now := <-ticker.C:
				cpu, mem := getProcessStats(proc, noChildren)

				times = append(times, now)
				cpuRecords = append(cpuRecords, cpu)
				memRecords = append(memRecords, mem)
			case <-sigs:
				rec := Records{times, cpuRecords, memRecords}
				save(rec, output, json)
				defer os.Exit(0)
			}

		}
	}()

	time.Sleep(duration)
	ticker.Stop()
	rec := Records{times, cpuRecords, memRecords}
	save(rec, output, json)
}
