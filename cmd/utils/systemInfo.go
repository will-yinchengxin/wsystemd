package utils

import (
	"fmt"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
	"log"
	"net"
	"os"
	"time"
)

func GetHostName() (string, error) {
	return os.Hostname()
}

func GetLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}
	return "", fmt.Errorf("no valid local IP found")
}

func GetCPUUsage() (float64, error) {
	cpuPercent, err := cpu.Percent(time.Second, false)
	if err != nil {
		return 0, err
	}
	if len(cpuPercent) == 0 {
		return 0, fmt.Errorf("no valid cpu usage found")
	}
	sum := 0.0
	for _, v := range cpuPercent {
		sum += v
	}
	avgCPUUsage := sum / float64(len(cpuPercent))
	return avgCPUUsage, nil
}

func GetMemoryUsage() (float64, error) {
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		log.Fatal(err)
		return 0, err
	}

	return memInfo.UsedPercent, nil
}

func GetLoadAverage() (float64, error) {
	loadAvg, err := load.Avg()
	if err != nil {
		log.Fatal(err)
		return 0, err
	}
	return loadAvg.Load5, nil
}
