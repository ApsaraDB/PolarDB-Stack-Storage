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


package mpathpersist

import (
	"fmt"
	"polardb-sms/pkg/common"
	smslog "polardb-sms/pkg/log"
	"time"

	"polardb-sms/pkg/agent/device/reservation"
	"polardb-sms/pkg/network/message"
)

type PrReserveExecutor struct {
}

func NewPrReserveExecutor() reservation.PrCmdExecutor {
	return &PrReserveExecutor{}
}

func (e *PrReserveExecutor) GetCmdString(cmd *message.PrCmd) (string, error) {
	param := cmd.CmdParam.(*message.PrReserveCmdParam)
	devicePath, err := common.GetDevicePath(cmd.VolumeId)
	if err != nil {
		return "", err
	}
	cmdStr := fmt.Sprintf("mpathpersist -v %d --out --reserve --param-rk=%s --prout-type=%d %s", MPLogLevel, param.RegisterKey, param.ReserveType, devicePath)
	return cmdStr, nil
}

func (e *PrReserveExecutor) PrCmdExec(cmd *message.PrCmd, timeout time.Duration) (int, error) {
	cmdStr, err := e.GetCmdString(cmd)
	if err != nil {
		return 0, err
	}
	return reservation.PrCmdExec(cmdStr, timeout)
}

func (e *PrReserveExecutor) NoNeedExec(cmd *message.PrCmd) bool {
	param := cmd.CmdParam.(*message.PrReserveCmdParam)

	devicePath, err := common.GetDevicePath(cmd.VolumeId)
	if err != nil {
		return true
	}
	prKey, err := GetPrInfo(devicePath)
	if err != nil {
		smslog.Infof("get pr info err, exec it")
		return false
	}
	if len(prKey.Keys) == 0 {
		_, ok := prKey.Keys[param.RegisterKey]
		if ok {
			return true
		}
	}
	return false
}
