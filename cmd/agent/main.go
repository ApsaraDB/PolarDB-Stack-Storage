/*
*Copyright (c) 2019-2021, Alibaba Group Holding Limited;
*Licensed under the Apache License, Version 2.0 (the "License");
*you may not use this file except in compliance with the License.
*You may obtain a copy of the License at

*   http://www.apache.org/licenses/LICENSE-2.0

*Unless required by applicable law or agreed to in writing, software
*distributed under the License is distributed on an "AS IS" BASIS,
*WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
*See the License for the specific language governing permissions and
*limitations under the License.
 */

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"net"
	"os"
	"polardb-sms/pkg/agent/device/dmhelper"
	"polardb-sms/pkg/agent/meta"
	smslog "polardb-sms/pkg/log"
	"polardb-sms/pkg/version"
	"runtime"
	"strings"
	"time"

	"k8s.io/sample-controller/pkg/signals"
	"polardb-sms/pkg/agent/msgserver"
	"polardb-sms/pkg/agent/service"
	"polardb-sms/pkg/agent/utils"
)

var (
	// agent server args
	port       = flag.String("port", "12345", "The port for the sms agent to serve on")
	reportPort = flag.String("report-port", "2002", "The port for the report port on remote server")
	address    = flag.String("address", "0.0.0.0", "The IP address for the sms agent to serve on")
	dataDir    = flag.String("data-dir", "/var/lib/sms-agent/", "The agent data directory")
	whiteIPs   = flag.String("white-ips", "0.0.0.0/0", "The white ip list that allowed to connect to sms agent")
	// monitor report args
	rules              = flag.String("rules", "", "The udev rules that sms agent listening, AND is separated by comma, OR is separated by |, e.g. 'SUBSYSTEM=net|SUBSYSTEM=block'")
	nodeId             = flag.String("node-id", "", "The node id, default is hostname")
	nodeIp             = flag.String("node-ip", "", "The node ip")
	disableUdevEvent   = flag.Bool("disable-udev-event", false, "Disable udev event")
	disableKernelEvent = flag.Bool("disable-kernel-event", false, "Disable kernel event")
	logLevel           = flag.String("smslog-level", "info", "The smslog level, can be debug,info,warning,error")
	logDir             = flag.String("logDir", "/var/log/polardb-box/polardb-sms/agent/", "sms smslog dir")
	localDiskDir       = flag.String("local-disk-dir", "/dev/local-disks/", "sms local disk or vg dir")

	pidFile = "/var/run/polar-sms-agent.pid"
)

const (
	LogFile = "polardb-sms-agent.smslog"
)

type Config struct {
	EventReporterConfig *service.EventReporterConfig
}

func (c *Config) String() string {
	out, err := json.Marshal(c)
	if err != nil {
		smslog.Error("failed to marshal config: %+v, err: %s", c, err)
		return ""
	}
	return string(out)
}

func init() {
	flag.Parse()
	smslog.InitLogger(*logDir, LogFile, smslog.LogLevel(*logLevel))
}

func buildConfig() (*Config, error) {
	var cfg = Config{
		EventReporterConfig: &service.EventReporterConfig{
			Port:       *port,
			ReportPort: *reportPort,
		},
	}

	if ip := net.ParseIP(*address); ip == nil {
		return nil, fmt.Errorf("invalid address: %s", *address)
	} else {
		cfg.EventReporterConfig.Address = *address
	}

	ips := strings.Split(*whiteIPs, ",")
	for _, ip := range ips {
		if !strings.Contains(ip, "/") {
			ip += "/32"
		}

		_, netAddr, err := net.ParseCIDR(ip)
		if err != nil {
			return nil, fmt.Errorf("failed to parse white ip: %s", err)
		}

		cfg.EventReporterConfig.WhiteIPs = append(cfg.EventReporterConfig.WhiteIPs, netAddr)
	}

	if *nodeId == "" {
		hostname, err := os.Hostname()
		if err != nil {
			return nil, fmt.Errorf("failed to get hostname: %s", err)
		}
		nodeId = &hostname
	}
	cfg.EventReporterConfig.NodeId = *nodeId
	cfg.EventReporterConfig.NodeIp = *nodeIp

	return &cfg, nil
}

func main() {
	defer smslog.LogPanic()
	version.LogVersion()
	dmhelper.InitBlacklist()
	runtime.GOMAXPROCS(2)

	smslog.Infof("Starting SMS Agent")
	rand.Seed(time.Now().UnixNano())

	cfg, err := buildConfig()
	if err != nil {
		smslog.Fatal(err)
	}
	smslog.Infof("config: %s", cfg.String())

	// 只允许启动一个实例，避免多个agent同时变更造成dm元数据损坏
	lockFile, err := utils.LockPath(pidFile)
	if err != nil {
		smslog.Fatal(err)
	}
	defer utils.UnlockPath(lockFile)

	// 准备工作目录
	if _, err := os.Stat(*dataDir); err != nil {
		if !os.IsNotExist(err) {
			smslog.Fatalf("check dm table path failed: %s", err)
		}

		if err = os.MkdirAll(*dataDir, 0x700); err != nil {
			smslog.Fatalf("make dm table path failed: %s", err)
		}
		smslog.Infof("successfully create data dir %s", *dataDir)
	} else {
		smslog.Infof("find existing data directory %s", *dataDir)
	}

	if err = meta.CreateDmStore(*dataDir); err != nil {
		smslog.Fatalf("create meta store failed: %s", err.Error())
	}

	stopCh := signals.SetupSignalHandler()

	msgServer := msgserver.NewMessageServer(*nodeId, *nodeIp, *port)
	go msgServer.Run()
	defer msgServer.Shutdown()

	server := service.NewEventReporterServer(cfg.EventReporterConfig)
	server.Run(stopCh)

	exitCode := 0

	smslog.Infof("Exit SMS Agent")
	os.Exit(exitCode)
}
