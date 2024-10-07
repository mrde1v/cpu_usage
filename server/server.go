package server

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/shirou/gopsutil/process"
)

type ProcessInfo struct {
	PID    int32
	Name   string
	MaxCPU float64
}

func Start() {
	usageMap := make(map[int32]float64)
	var mu sync.Mutex
	done := make(chan bool)

	go func() {
		var input string
		for {
			fmt.Scanln(&input)
			if input == "q" {
				done <- true
				return
			}
		}
	}()

	fmt.Println("Monitoring CPU usage... Press 'q' to stop.")

	ticker := time.NewTicker(100 * time.Millisecond) // adjust to reduce CPU usage
	defer ticker.Stop()

	for {
		select {
		case <-done:
			// stop monitoring
			goto report
		case <-ticker.C:
			procs, err := process.Processes()
			if err != nil {
				fmt.Println("Error fetching processes:", err)
				return
			}

			mu.Lock()
			for _, proc := range procs {
				cpuPercent, err := proc.CPUPercent()
				if err != nil {
					continue
				}

				if currentMax, exists := usageMap[proc.Pid]; !exists || cpuPercent > currentMax {
					usageMap[proc.Pid] = cpuPercent
				}
			}
			mu.Unlock()
		}
	}

report:
	var processList []ProcessInfo
	mu.Lock()
	for pid, maxCPU := range usageMap {
		proc, err := process.NewProcess(pid)
		if err == nil {
			name, _ := proc.Name()
			processList = append(processList, ProcessInfo{PID: pid, Name: name, MaxCPU: maxCPU})
		}
	}
	mu.Unlock()

	sort.Slice(processList, func(i, j int) bool {
		return processList[i].MaxCPU > processList[j].MaxCPU
	})

	fmt.Printf("\nTop 10 processes by peak CPU usage:\n")
	for i, proc := range processList {
		if i >= 10 {
			break
		}
		fmt.Printf("PID: %d, Name: %s, Peak CPU Usage: %.2f%%\n", proc.PID, proc.Name, proc.MaxCPU)
	}
}
