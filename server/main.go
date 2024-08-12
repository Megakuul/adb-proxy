package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/megakuul/adb-proxy/server/api"
	"github.com/megakuul/adb-proxy/server/discover"
	"github.com/megakuul/adb-proxy/server/proxy"
)

func parsePortRange(arg string) (int, int, error) {
	argSlice := strings.Split(arg, "-")
	if len(argSlice) != 2 {
		return 0, 0, fmt.Errorf("expected 2 ports separated by '-'")
	}
	firstPort, err := strconv.Atoi(argSlice[0])
	if err!=nil {
		return 0, 0, err
	}
	lastPort, err := strconv.Atoi(argSlice[1])
	if err!=nil {
		return 0, 0, err
	}
	return firstPort, lastPort, nil
}

func main() {
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors: true,
		FullTimestamp: true,
	})
	
	if err := run(); err!=nil {
		logrus.Errorf("%v\n", err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) != 4 {
		return fmt.Errorf(
			"expected 3 arguments:\n" +
				"1. http api port (e.g. '7000')\n" +
				"2. proxy discovery port (e.g. '6775')\n" +
				"3. proxy port range (e.g. '8990-9000')\n")
	}
	apiPort, err := strconv.Atoi(os.Args[1])
	if err!=nil {
		return fmt.Errorf("failed to parse http api port")
	}
	proxyPort, err := strconv.Atoi(os.Args[2])
	if err!=nil {
		return fmt.Errorf("failed to parse http api port")
	}
	firstPort, lastPort, err := parsePortRange(os.Args[3])
	if err!=nil {
		return fmt.Errorf("failed to parse argument: %v", err)
	}
	
	deviceController := proxy.NewDeviceController(uint16(firstPort), uint16(lastPort))

	go func() {
		logrus.Infof("starting http api listener on %d\n", apiPort)
		if err := api.StartApiListener(uint16(apiPort), deviceController); err!=nil {
			logrus.Errorf("failed to run api listener: %v", err)
		}
	}()

	logrus.Infof("starting discovery listener on %d\n", proxyPort)
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", proxyPort))
	if err!=nil {
		return fmt.Errorf("failed to listen: %v", err)
	}
	discover.StartDiscoverListener(listener, deviceController)

	return nil
}
