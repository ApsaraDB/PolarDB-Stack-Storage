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

package nvme

import (
	"encoding/json"
	"fmt"
	"polardb-sms/pkg/agent/device/reservation/mpathpersist"
	"polardb-sms/pkg/agent/utils"
)

type NvmeReservationInfo struct {
	Generation                   int64                   `json:"gen"`
	ReservationType              int                     `json:"rtype"`
	ControllersCount             int                     `json:"regctl"`
	PersistThroughPowerLossState int                     `json:"ptpls"`
	Controllers                  []*RegisteredController `json:"regctlext"`
}

type RegisteredController struct {
	ControllerId      int    `json:"cntlid"`
	ReservationStatus int    `json:"rcsts"`
	ReservationKey    int64  `json:"rkey"`
	HostId            string `json:"hostid"`
}

func NewNvmeReservationInfo() *NvmeReservationInfo {
	return &NvmeReservationInfo{}
}

func GetPrInfo(device string) (*mpathpersist.PersistentReserve, error) {
	var (
		cmd string
		pr  = mpathpersist.NewPersistentReserve()
	)
	cmd = fmt.Sprintf("nvme resv-report %s -n 1 -c 0x1 -o json", device)
	keyOut, keyErr, err := utils.ExecCommand(cmd, mpathpersist.DefaultTimeout)
	if err != nil || keyErr != "" {
		return nil, fmt.Errorf("failed to query nvme pr key, stdout: %s, stderr: %s, err: %s", keyOut, keyErr, err)
	}
	nvmeResvInfo := NewNvmeReservationInfo()
	if err := json.Unmarshal([]byte(keyOut), &nvmeResvInfo); err != nil {
		return nil, fmt.Errorf("invalid nvme reservation info, err: %s", err)
	}

	for _, controller := range nvmeResvInfo.Controllers {
		count := 1
		key := fmt.Sprintf("0x%x", controller.ReservationKey)
		if tmp, ok := pr.Keys[key]; ok {
			count = tmp
		}
		pr.Keys[key] = count
	}

	return pr, err
}
