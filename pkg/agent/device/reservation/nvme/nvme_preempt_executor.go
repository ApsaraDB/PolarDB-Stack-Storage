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
	"fmt"
	"polardb-sms/pkg/common"
	smslog "polardb-sms/pkg/log"
	"time"

	"polardb-sms/pkg/agent/device/reservation"
	"polardb-sms/pkg/network/message"
)

type NvmePreemptExecutor struct {
}

func NewNvmePreemptExecutor() reservation.PrCmdExecutor {
	return &NvmePreemptExecutor{}
}

func (e *NvmePreemptExecutor) GetCmdString(cmd *message.PrCmd) (string, error) {
	param := cmd.CmdParam.(*message.PrPreemptCmdParam)
	cmdStr := fmt.Sprintf("nvme resv-acquire /dev/mapper/%s -n 1 -c %s -p %s -t 5 -a 1", cmd.VolumeId, param.RegisterKey, param.PreemptedKey)
	return cmdStr, nil
}

func (e *NvmePreemptExecutor) PrCmdExec(cmd *message.PrCmd, timeout time.Duration) (int, error) {
	devicePath, err := common.GetDevicePath(cmd.VolumeId)
	if err != nil {
		return 0, err
	}
	prKey, err := GetPrInfo(devicePath)
	smslog.Debugf("prKey: %v", prKey)
	if err != nil {
		smslog.Infof("get nvme pr info err %s, exec it", err.Error())
		return 0, err
	}
	param := cmd.CmdParam.(*message.PrPreemptCmdParam)
	smslog.Debugf("param: %v", param)
	_, existRegisterKey := prKey.Keys[param.RegisterKey]
	if !existRegisterKey {
		//register
		cmdStr := fmt.Sprintf("nvme resv-register %s -n 1 -k %s", devicePath, param.RegisterKey)
		smslog.Debugf("register cmd: %s", cmdStr)
		_, err = reservation.PrCmdExec(cmdStr, timeout)
		if err != nil {
			return 0, err
		}
	}
	cmdStr, err := e.GetCmdString(cmd)
	if err != nil {
		return 0, err
	}
	return reservation.PrCmdExec(cmdStr, timeout)
}
