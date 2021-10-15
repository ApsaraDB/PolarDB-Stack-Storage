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

type LvCreateStageRunner struct {
	*Stage
}

func (s *LvCreateStageRunner) Run(ctx common.TraceContext) *StageExecResult {
	msg, err := message.NewMessage(message.SmsMessageHead_CMD_DM_CREAT_REQ, s.Content, ctx)
	if err != nil {
		return StageExecFail(err.Error())
	}
	ret := sendToAllParallel(msg, BaseTimeout, config.AvailableNodes())
	s.Result = ret
	return ret
}

func (s *LvCreateStageRunner) Rollback(ctx common.TraceContext) *StageExecResult {
	return StageExecFail(fmt.Errorf("umimplement fs expand rollback").Error())
}

func NewLvCreateStage(core *device.DmDeviceCore) *LvCreateStageRunner {
	return &LvCreateStageRunner{
		Stage: &Stage{
			Content: &message.DmExecCommand{
				CommandType: message.Create,
				DeviceName:  core.VolumeId,
				Device:      core,
			},
			SType:     LvCreateStage,
			StartTime: 0,
			Result:    nil,
		},
	}
}

type LvCreateStageConstructor struct {
}

func (c *LvCreateStageConstructor) Construct() interface{} {
	return &LvCreateStageRunner{}
}

type LvDeleteStageRunner struct {
	*Stage
}

func (s *LvDeleteStageRunner) Run(ctx common.TraceContext) *StageExecResult {
	msg, err := message.NewMessage(message.SmsMessageHead_CMD_DM_DELETE_REQ, s.Content, ctx)
	if err != nil {
		return StageExecFail(err.Error())
	}
	ret := sendToAllParallel(msg, BaseTimeout, config.AvailableNodes())
	s.Result = ret
	return ret
}

func (s *LvDeleteStageRunner) Rollback(ctx common.TraceContext) *StageExecResult {
	return StageExecFail(fmt.Errorf("umimplement fs expand rollback").Error())
}

//todo fix this
func NewLvDeleteStage(core *device.DmDeviceCore) *LvDeleteStageRunner {
	return &LvDeleteStageRunner{
		Stage: &Stage{
			Content: &message.DmExecCommand{
				CommandType: message.Delete,
				DeviceName:  core.VolumeId,
				Device:      core,
			},
			SType:     LvDeleteStage,
			StartTime: 0,
			Result:    nil,
		},
	}
}

type LvDeleteStageConstructor struct {
}

func (c *LvDeleteStageConstructor) Construct() interface{} {
	return &LvDeleteStageRunner{}
}

type LvExpandStageRunner struct {
	*Stage
}

func (s *LvExpandStageRunner) Run(ctx common.TraceContext) *StageExecResult {
	msg, err := message.NewMessage(message.SmsMessageHead_CMD_DM_UPDATE_REQ, s.Content, ctx)
	if err != nil {
		return StageExecFail(err.Error())
	}
	ret := sendToAllParallel(msg, s.timeout(), config.AvailableNodes())
	s.Result = ret
	return ret
}

func (s *LvExpandStageRunner) Rollback(ctx common.TraceContext) *StageExecResult {
	return StageExecFail(fmt.Errorf("umimplement fs expand rollback").Error())
}

func (s *LvExpandStageRunner) timeout() int64 {
	var (
		dmCommand *message.DmExecCommand
	)
	command := s.Content.(map[string]interface{})
	if err := common.MapToStruct(command, &dmCommand); err != nil {
		return BaseTimeout
	}
	reqSizeIn100GiB := dmCommand.Device.SectorNum * int64(dmCommand.Device.SectorSize) / (100 * 1024 * 1024 * 1024)
	return BaseTimeout + TimeoutPer100G*reqSizeIn100GiB
}

//todo fix this
func NewLvExpandStage(core *device.DmDeviceCore) *LvExpandStageRunner {
	return &LvExpandStageRunner{
		Stage: &Stage{
			Content: &message.DmExecCommand{
				CommandType: message.Expand,
				DeviceName:  core.VolumeId,
				Device:      core,
			},
			SType:     LvExpandStage,
			StartTime: 0,
			Result:    nil,
		},
	}
}

type LvExpandStageConstructor struct {
}

func (c *LvExpandStageConstructor) Construct() interface{} {
	return &LvExpandStageRunner{}
}
