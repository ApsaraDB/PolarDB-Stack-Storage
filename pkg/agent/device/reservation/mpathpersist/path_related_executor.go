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
	"strconv"
	"strings"
	"time"

	"polardb-sms/pkg/agent/device/reservation"
	"polardb-sms/pkg/agent/utils"
	"polardb-sms/pkg/network/message"
)

type PrPathNumExecutor struct {
}

func NewPrPathNumExecutor() reservation.PrCmdExecutor {
	return &PrPathNumExecutor{}
}

func (e *PrPathNumExecutor) GetCmdString(cmd *message.PrCmd) (string, error) {
	devicePath, err := common.GetDevicePath(cmd.VolumeId)
	if err != nil {
		return "", err
	}
	cmdStr := fmt.Sprintf("multipath -ll %s | grep -E '[0-9]+:[0-9]+:[0-9]+:[0-9]+'|wc -l", devicePath)
	return cmdStr, nil
}

func (e *PrPathNumExecutor) PrCmdExec(cmd *message.PrCmd, timeout time.Duration) (int, error) {
	cmdStr, err := e.GetCmdString(cmd)
	if err != nil {
		return 0, err
	}
	outInfo, errInfo, err := utils.ExecCommand(cmdStr, timeout)
	if err != nil {
		return 0, fmt.Errorf("stdout: %s, stderr: %s, err: %s", outInfo, errInfo, err)
	}
	cnt, err := strconv.Atoi(strings.Trim(outInfo, "\n"))
	if err != nil {
		smslog.Infof("%s convert to int err", outInfo)
		return 0, nil
	}
	return cnt, nil
}

type PrPathCanWriteExecutor struct {
}

func NewPrPathCanWriteExecutor() reservation.PrCmdExecutor {
	return &PrPathCanWriteExecutor{}
}

func (e *PrPathCanWriteExecutor) GetCmdString(cmd *message.PrCmd) (string, error) {
	devicePath, err := common.GetDevicePath(cmd.VolumeId)
	if err != nil {
		return "", err
	}
	cmdStr := fmt.Sprintf("dd if=/dev/zero of=%s seek=512 bs=512 count=1 oflag=direct", devicePath)
	return cmdStr, nil
}

func (e *PrPathCanWriteExecutor) PrCmdExec(cmd *message.PrCmd, timeout time.Duration) (int, error) {
	cmdStr, err := e.GetCmdString(cmd)
	if err != nil {
		return 0, err
	}
	return reservation.PrCmdExec(cmdStr, timeout)
}

type PrPathCannotWriteExecutor struct {
}

func NewPrPathCannotWriteExecutor() reservation.PrCmdExecutor {
	return &PrPathCannotWriteExecutor{}
}

func (e *PrPathCannotWriteExecutor) GetCmdString(cmd *message.PrCmd) (string, error) {
	devicePath, err := common.GetDevicePath(cmd.VolumeId)
	if err != nil {
		return "", err
	}
	cmdStr := fmt.Sprintf("dd if=/dev/zero of=%s bs=4k count=1", devicePath)
	return cmdStr, nil
}

func (e *PrPathCannotWriteExecutor) PrCmdExec(cmd *message.PrCmd, timeout time.Duration) (int, error) {
	cmdStr, err := e.GetCmdString(cmd)
	if err != nil {
		return 0, err
	}
	return reservation.PrCmdExec(cmdStr, timeout)
}
