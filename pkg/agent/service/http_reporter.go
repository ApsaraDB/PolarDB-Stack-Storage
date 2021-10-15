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
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"polardb-sms/pkg/agent/msgserver"
	smslog "polardb-sms/pkg/log"
	"polardb-sms/pkg/protocol"
)

type HttpEventReporter struct {
	NodeIp     string
	ReportPort string
}

var _httpClient = &http.Client{
	Transport: &http.Transport{
		DisableKeepAlives: true,
	},
}

func NewHttpEventReporter(nodeIp string, reportPort string) EventReporter {
	return &HttpEventReporter{
		NodeIp:     nodeIp,
		ReportPort: reportPort,
	}
}

func (r *HttpEventReporter) pickServerEndpoint() string {
	serverIp := msgserver.PickOneClientIp()
	if serverIp == "" {
		serverIp = r.NodeIp
	}
	return fmt.Sprintf("%s:%s", serverIp, r.ReportPort)
}

func (r *HttpEventReporter) endpoint() string {
	return fmt.Sprintf("http://%s/events", r.pickServerEndpoint())
}

func (r *HttpEventReporter) batchEndpoint() string {
	return fmt.Sprintf("http://%s/events/batch", r.pickServerEndpoint())
}

func (r *HttpEventReporter) heartbeatEndpoint() string {
	return fmt.Sprintf("http://%s/agent/heartbeat", r.pickServerEndpoint())
}

func (r *HttpEventReporter) Report(event *protocol.Event) error {
	reqBytes, err := json.Marshal(event)
	if err != nil {
		smslog.Infof(fmt.Errorf("marshal request body %v: %v", event, err).Error())
	}
	return send(reqBytes, r.endpoint())
}

func (r *HttpEventReporter) BatchReport(batchEvent *protocol.BatchEvent) error {
	reqBytes, err := json.Marshal(batchEvent)
	if err != nil {
		smslog.Infof(fmt.Errorf("marshal request body %v: %v", batchEvent, err).Error())
	}
	return send(reqBytes, r.batchEndpoint())
}

func (r *HttpEventReporter) Heartbeat(heartbeat *HeartbeatRequest) error {
	reqBytes, err := json.Marshal(heartbeat)
	if err != nil {
		smslog.Infof(fmt.Errorf("marshal request body %v: %v", heartbeat, err).Error())
	}
	return send(reqBytes, r.heartbeatEndpoint())
}

func send(body []byte, url string) error {
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		err := fmt.Errorf("new request: %v", err)
		smslog.Infof(err.Error())
		return err
	}
	resp, err := _httpClient.Do(req)
	if err != nil || resp.StatusCode != 200 {
		smslog.Errorf("send report event to %s, response failed, resp: %v err: %v", url, resp, err)
	}
	defer func() {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}()
	smslog.Debugf("send report event response finished: %v", resp)
	return nil
}
