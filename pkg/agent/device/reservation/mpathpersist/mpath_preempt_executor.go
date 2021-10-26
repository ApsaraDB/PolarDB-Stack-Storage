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
	"polardb-sms/pkg/agent/device/reservation"
	"polardb-sms/pkg/common"
	smslog "polardb-sms/pkg/log"
	"polardb-sms/pkg/network/message"
	"time"
)

type PrPreemptExecutor struct {
}

func NewPrPreemptExecutor() reservation.PrCmdExecutor {
	return &PrPreemptExecutor{}
}

func (e *PrPreemptExecutor) GetCmdString(cmd *message.PrCmd) (string, error) {
	param := cmd.CmdParam.(*message.PrPreemptCmdParam)
	devicePath, err := common.GetDevicePath(cmd.VolumeId)
	if err != nil {
		return "", err
	}
	cmdStr := fmt.Sprintf("sg_persist --out --preempt --param-sark=%s --param-rk=%s --prout-type=%d %s",
		param.PreemptedKey, param.RegisterKey, param.ReserveType, devicePath)
	return cmdStr, nil
}

//todo refact this
func (e *PrPreemptExecutor) PrCmdExec(cmd *message.PrCmd, timeout time.Duration) (int, error) {
	if e.NoNeedExec(cmd) {
		return 0, nil
	}
	devicePath, err := common.GetDevicePath(cmd.VolumeId)
	if err != nil {
		return 0, err
	}
	prKey, err := GetPrInfo(devicePath)
	if err != nil {
		smslog.Infof("get pr info err %s, exec it", err.Error())
		return 0, err
	}

	param := cmd.CmdParam.(*message.PrPreemptCmdParam)
	_, existPreemptedKey := prKey.Keys[param.PreemptedKey]
	_, existRegisterKey := prKey.Keys[param.RegisterKey]
	if !existRegisterKey {
		//register
		cmdStr := fmt.Sprintf("mpathpersist -v %d --out --register --param-sark=%s %s", MPLogLevel, param.RegisterKey, devicePath)
		_, err = reservation.PrCmdExec(cmdStr, timeout)
		if err != nil {
			return 0, err
		}
	}
	if !existPreemptedKey {
		//reserve
		cmdStr := fmt.Sprintf("mpathpersist -v %d --out --reserve --param-rk=%s --prout-type=%d %s", MPLogLevel, param.RegisterKey, param.ReserveType, devicePath)
		_, err = reservation.PrCmdExec(cmdStr, timeout)
		if err != nil {
			return 0, err
		}
		return 0, nil
	}
	cmdStr := fmt.Sprintf("sg_persist --out --preempt --param-sark=%s --param-rk=%s --prout-type=%d %s",
		param.PreemptedKey, param.RegisterKey, param.ReserveType, devicePath)
	return reservation.PrCmdExec(cmdStr, timeout)
}

func (e *PrPreemptExecutor) NoNeedExec(cmd *message.PrCmd) bool {
	param := cmd.CmdParam.(*message.PrPreemptCmdParam)
	devicePath, err := common.GetDevicePath(cmd.VolumeId)
	if err != nil {
		return false
	}
	prKey, err := GetPrInfo(devicePath)
	if err != nil {
		smslog.Infof("get pr info err %s, exec it", err.Error())
		return false
	}
	if prKey.ReservationKey != "" && prKey.ReservationType == PRC_WR_EX {
		if len(prKey.Keys) == 1 {
			_, ok := prKey.Keys[param.RegisterKey]
			if ok {
				return true
			}
		}
	}
	return false
}
