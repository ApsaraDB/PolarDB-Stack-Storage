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

package reservation

import (
	"fmt"
	"time"

	"polardb-sms/pkg/agent/utils"
	"polardb-sms/pkg/network/message"
)

const (
	DefaultTimeout = 5 * time.Second
)

type PrCmdExecutor interface {
	PrCmdExec(cmd *message.PrCmd, timeout time.Duration) (int, error)
}

func ExecuteCmd(e PrCmdExecutor, cmd *message.PrCmd) (*message.PrCheckCmdResult, error) {
	var (
		result = 0
		err    error
	)
	result, err = e.PrCmdExec(cmd, DefaultTimeout)
	if err != nil {
		return nil, err
	}
	return &message.PrCheckCmdResult{
		CheckType:   cmd.CmdType,
		CheckResult: result,
		VolumeType:  cmd.VolumeType,
		Name:        cmd.VolumeId,
	}, nil
}

func PrCmdExec(cmd string, timeout time.Duration) (int, error) {
	outInfo, errInfo, err := utils.ExecCommand(cmd, timeout)
	if err != nil {
		return 0, fmt.Errorf("cmd %s :: stdout: %s, stderr: %s, err: %s", cmd, outInfo, errInfo, err)
	}
	return 0, nil
}
