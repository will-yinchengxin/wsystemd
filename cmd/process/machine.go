package process

import (
	"fmt"
	"github.com/go-kit/kit/log/level"
	"net"
	"os"
	"wsystemd/cmd/log"
)

func GetHostName() (string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		level.Error(log.Logger).Log("err", fmt.Sprintf("无法获取主机名: %s", err.Error()))
		return "", err
	}
	return hostname, nil
}

func GetServerIp() (map[string]struct{}, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		level.Error(log.Logger).Log("err", fmt.Sprintf("无法获取 IP 地址: %s", err.Error()))
		return nil, err
	}
	ips := make(map[string]struct{}, 0)
	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				level.Info(log.Logger).Log("info", fmt.Sprintf("IPv4 地址: %s", ipNet.IP.String()))
				ips[ipNet.IP.String()] = struct{}{}
			}
		}
	}
	return ips, nil
}

func GetServerIpV6() (map[string]struct{}, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		level.Error(log.Logger).Log("err", fmt.Sprintf("无法获取 IP 地址: %s", err.Error()))
		return nil, err
	}
	ips := make(map[string]struct{}, 0)
	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To16() != nil {
				level.Info(log.Logger).Log("info", fmt.Sprintf("IPv6 地址: %s", ipNet.IP.String()))
				ips[ipNet.IP.String()] = struct{}{}
			}
		}
	}
	return ips, nil
}

func HNContail(hs string) (bool, error) {
	name, err := GetHostName()
	if err != nil {
		return false, err
	}
	return name == hs, nil
}

func IpContail(ip string) (bool, error) {
	ips, err := GetServerIp()
	if err != nil {
		return false, err
	}
	_, ok := ips[ip]
	return ok, nil
}

func IpV6Contail(ip string) (bool, error) {
	ips, err := GetServerIpV6()
	if err != nil {
		return false, err
	}
	_, ok := ips[ip]
	return ok, nil
}
