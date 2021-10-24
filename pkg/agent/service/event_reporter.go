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

package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	_ "net/http/pprof"
	"polardb-sms/pkg/agent/device/devicemapper"
	"strings"
	"time"

	"polardb-sms/pkg/agent/device/dmhelper"
	"polardb-sms/pkg/agent/meta"
	"polardb-sms/pkg/device"
	"polardb-sms/pkg/protocol"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
	smslog "polardb-sms/pkg/log"
)

//go:generate protoc -I ./protocol ./protocol/services.proto --go_out=plugins=grpc:protocol

var (
	ReportQueue = make(chan *EventReport, 1000)
)

const (
	FullReportInterval = 10 * time.Minute
	HeartbeatInterval  = 15 * time.Second
)

type EventReporterConfig struct {
	NodeId     string
	NodeIp     string
	Port       string
	ReportPort string
	Address    string
	WhiteIPs   []*net.IPNet
}
type EventReporter interface {
	Report(event *protocol.Event) error
	BatchReport(batchEvent *protocol.BatchEvent) error
	Heartbeat(heartbeat *HeartbeatRequest) error
}

type EventReporterServer struct {
	cfg      *EventReporterConfig
	ts       meta.DMTableStore
	reporter EventReporter
}

func NewEventReporterServer(cfg *EventReporterConfig) *EventReporterServer {
	return &EventReporterServer{
		cfg:      cfg,
		ts:       meta.GetDmStore(),
		reporter: NewHttpEventReporter(cfg.NodeIp, cfg.ReportPort),
	}
}

func (s *EventReporterServer) Run(stopCh <-chan struct{}) {
	smslog.Infof("Starting Agent Server")
	defer smslog.LogPanic()
	deviceMap, err := s.LoadLocalTables()
	if err != nil {
		smslog.Errorf("LoadLocalTables err %s", err)
	}
	smslog.Info("LoadLocalTables finished, batch report")
	go s.batchReport(deviceMap)
	go s.FullReportLoop(stopCh)

	go s.DeltaReportLoop(stopCh)
	go s.Heartbeat(stopCh)

	<-stopCh
	smslog.Infof("Shutting down Agent Server")
}

func (s *EventReporterServer) batchReport(deviceMap map[string]*device.DmDevice) {
	defer smslog.LogPanic()
	smslog.Debugf("startup batch report %v", deviceMap)
	batchEvent := &protocol.BatchEvent{
		Events: make([]*protocol.Event, 0),
		NodeId: s.cfg.NodeId,
	}
	for _, d := range deviceMap {
		//todo fixme
		var event interface{}
		if d.DeviceType == "multipath" {
			event = s.getTransformer(d.DeviceType).Transform(d, protocol.LunAdd)
		} else {
			event = s.getTransformer(d.DeviceType).Transform(d, protocol.LvAdd)
		}
		if event != nil {
			batchEvent.Events = append(batchEvent.Events, event.(*protocol.Event))
		}
	}
	_ = s.reporter.BatchReport(batchEvent)
}

func (s *EventReporterServer) simpleAuth(ctx context.Context) (context.Context, error) {
	if client, ok := peer.FromContext(ctx); ok {
		ip := strings.SplitN(client.Addr.String(), ":", 2)[0]

		clientIP := net.ParseIP(ip)
		if clientIP == nil {
			return ctx, status.Error(codes.InvalidArgument, fmt.Sprintf("invalid client ip %s", ip))
		}

		for _, item := range s.cfg.WhiteIPs {
			if item.Contains(clientIP) {
				return ctx, nil
			}
		}

		return ctx, status.Error(codes.PermissionDenied, fmt.Sprintf("client %s is not in white ip list", ip))
	}
	return ctx, status.Error(codes.InvalidArgument, "empty metadata error")
}

func (s *EventReporterServer) Heartbeat(stopCh <-chan struct{}) {
	smslog.Debugf("start to heartbeat")
	defer smslog.LogPanic()
	for {
		select {
		case <-stopCh:
			smslog.Infof("heartbeat stopped")
			return
		case <-time.After(HeartbeatInterval):
			heartbeat := &HeartbeatRequest{
				AgentId:   s.cfg.NodeId,
				NodeId:    s.cfg.NodeId,
				NodeIp:    s.cfg.NodeIp,
				Port:      s.cfg.Port,
				Timestamp: time.Now().Format("2006-01-02 15:04:05"),
			}
			_ = s.reporter.Heartbeat(heartbeat)
		}
	}
}

// 定时全量汇报，补偿运行过程中复杂环境变化带来的各种偏差
func (s *EventReporterServer) FullReportLoop(stopCh <-chan struct{}) {
	smslog.Infof("full report starting")
	defer smslog.LogPanic()
	for {
		select {
		case <-stopCh:
			smslog.Infof("full report stopped")
			return
		case <-time.After(FullReportInterval):
		}

		start := time.Now()
		dmDevices, err := dmhelper.QueryDMDevices()
		if err != nil {
			smslog.Errorf("full report failed to query dm devices: %s", err)
			continue
		}
		s.batchReport(dmDevices)
		smslog.Infof("full report finished, cost: %ds", time.Now().Sub(start).Milliseconds())
	}
}

// 实时增量汇报
func (s *EventReporterServer) DeltaReportLoop(stopCh <-chan struct{}) {
	smslog.Infof("delta report starting")
	defer smslog.LogPanic()
	for {
		select {
		case <-stopCh:
			smslog.Infof("delta report stopped")
			return
		case eventReport, ok := <-ReportQueue:
			if !ok {
				smslog.Infof("delta report queue closed")
				return
			}

			if err := s.report(eventReport); err != nil {
				smslog.Errorf("delta report failed: %v", err)
			}
		}
	}
}

// 启动的时候加载本地table，恢复盘的状态
func (s *EventReporterServer) LoadLocalTables() (map[string]*device.DmDevice, error) {
	devices, err := dmhelper.QueryDMDevices()
	if err != nil {
		return nil, fmt.Errorf("LoadLocalTables: failed to query all dm devices, err: %s", err)
	}
	tables, err := s.ts.List()
	if err != nil {
		return nil, fmt.Errorf("failed to load table files, %s", err)
	}
	smslog.Debug("LoadLocalTables:  start update table")
	var dm = devicemapper.GetDeviceMapper()
	for name, table := range tables {
		actualDevice, ok := devices[name]
		if !ok {
			if err := dm.DmSetupCreate(name, "lv", table.Data); err != nil {
				smslog.Errorf("failed to create device %s by load local table %s,err: %v", name, table, err)
			} else {
				smslog.Infof("successfully create device %s by load local table %s", name, table)
			}
			continue
		}

		expectDevice, err := dmhelper.ParseDMDevice(name, table.Data)
		if err != nil {
			smslog.Errorf("failed to parse local table for device %s, table %s, err: %s", name, table, err)
			continue
		}

		if !actualDevice.Compare(expectDevice) {
			smslog.Errorf("device %s already exists, but not equal with local table %s", name, table)
		}
	}

	smslog.Debug("LoadLocalTables:  finished load table")
	return devices, nil
}

func (s *EventReporterServer) getTransformer(deviceType device.DmDeviceType) Transformer {
	return getTransformer(deviceType, s.cfg.NodeId, s.cfg.NodeIp)
}

func (s *EventReporterServer) report(report *EventReport) error {
	smslog.Infof("[report] received new report: [%s::%v]", report.Action, *report.Device)
	var eventType protocol.EventType

	switch report.Action {
	case "add":
		eventType = protocol.LunAdd
	case "change":
		eventType = protocol.LvUpdate
	case "remove":
		eventType = protocol.LvRemove
		smslog.Infof("do not support remove action")
		return nil
	default:
		return fmt.Errorf("event lv - (%s) action [%s] not define", device.Linear, report.Action)
	}
	event := s.getTransformer(report.Device.DeviceType).Transform(report.Device, eventType)
	return s.reporter.Report(event.(*protocol.Event))
}

type EventReport struct {
	Action string
	Device *device.DmDevice
}

func (e *EventReport) String() string {
	out, err := json.Marshal(e)
	if err != nil {
		smslog.Errorf("failed to marshal event report %v: %s", e, err)
		return ""
	}
	return string(out)
}
