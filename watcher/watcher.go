package watcher

import (
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/shirou/gopsutil/process"
)

type Records struct {
	Time []time.Time `json:"times"`
	CPU  []float64   `json:"cpu"`
	Mem  []float64   `json:"mem"`
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

	return cpu / 100, mem64 / 100
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
