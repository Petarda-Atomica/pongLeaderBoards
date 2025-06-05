package main

import (
	"fmt"
	"net"
	"time"
)

func getLocalIP() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String(), nil
}

func broadcastLocalIP(port int, interval time.Duration) error {
	localIP, err := getLocalIP()
	if err != nil {
		return err
	}

	broadcastAddr := fmt.Sprintf("255.255.255.255:%d", port)
	udpAddr, err := net.ResolveUDPAddr("udp", broadcastAddr)
	if err != nil {
		return err
	}

	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		return err
	}
	defer conn.Close()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		_, err := conn.Write([]byte(localIP))
		if err != nil {
			return err
		}
	}

	return nil
}
