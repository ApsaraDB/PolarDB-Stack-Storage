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
	"time"

	"polardb-sms/pkg/agent/device/reservation"
	"polardb-sms/pkg/network/message"
)

type PrReleaseExecutor struct {
}

func NewPrReleaseExecutor() reservation.PrCmdExecutor {
	return &PrReleaseExecutor{}
}

func (e *PrReleaseExecutor) GetCmdString(cmd *message.PrCmd) (string, error) {
	devicePath, err := common.GetDevicePath(cmd.VolumeId)
	if err != nil {
		return "", err
	}
	cmdStr := fmt.Sprintf("mpathpersist -v %d --out --register-ignore %s", MPLogLevel, devicePath)
	return cmdStr, nil
}

func (e *PrReleaseExecutor) PrCmdExec(cmd *message.PrCmd, timeout time.Duration) (int, error) {
	cmdStr, err := e.GetCmdString(cmd)
	if err != nil {
		return 0, err
	}
	return reservation.PrCmdExec(cmdStr, timeout)
}

func (e *PrReleaseExecutor) NoNeedExec(cmd *message.PrCmd) bool {
	return false
}
