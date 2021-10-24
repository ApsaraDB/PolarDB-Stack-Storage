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

package stage

import (
	"fmt"
	"polardb-sms/pkg/common"
	"polardb-sms/pkg/device"
	"polardb-sms/pkg/manager/config"
	"polardb-sms/pkg/network/message"
)

const (
	DmExecTimeout = 5 //time.second
)

type DmCreateRunner struct {
	*Stage
	DmDevice *device.DmDevice
}

func (s *DmCreateRunner) Run(ctx common.TraceContext) *StageExecResult {
	msg, err := message.NewMessage(message.SmsMessageHead_CMD_DM_CREAT_REQ, s.Content, ctx)
	if err != nil {
		return StageExecFail(err.Error())
	}
	ret := sendToAllParallel(msg, DmExecTimeout, config.AvailableNodes())
	s.Result = ret
	return ret
}

func (s *DmCreateRunner) Rollback(ctx common.TraceContext) *StageExecResult {
	return StageExecFail(fmt.Errorf("umimplement fs expand rollback").Error())
}

func NewDmExecStage(name, content string) *DmCreateRunner {
	return &DmCreateRunner{
		Stage: &Stage{
			SType:     DmExecStage,
			StartTime: 0,
			Result:    nil,
			Content:   &message.DmExecCommand{},
		},
	}
}

type DmExecStageConstructor struct {
}

func (c *DmExecStageConstructor) Construct() interface{} {
	return &DmCreateRunner{
		Stage: &Stage{
			Content: &message.DmExecCommand{},
		},
	}
}
