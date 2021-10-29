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
	"polardb-sms/pkg/manager/config"
	"polardb-sms/pkg/network/message"
)

type PvcCreateStageRunner struct {
	*Stage
	TargetNode config.Node `json:"target_node"`
	ReqSize    int64       `json:"req_size"`
}

func (s *PvcCreateStageRunner) Run(ctx common.TraceContext) *StageExecResult {
	msg, err := message.NewMessage(message.SmsMessageHead_CMD_PVC_CREATE_REQ, s.Content, ctx)
	if err != nil {
		return StageExecFail(err.Error())
	}
	ret := sendAndWait(msg, s.TargetNode.Name, s.timeout())
	s.Result = ret
	return ret
}

func (s *PvcCreateStageRunner) Rollback(ctx common.TraceContext) *StageExecResult {
	return StageExecFail(fmt.Errorf("umimplement fs expand rollback").Error())
}

func (s *PvcCreateStageRunner) timeout() int64 {
	//todo fix
	reqSizeIn100GiB := s.ReqSize / (100 * 1024 * 1024 * 1024)
	return BaseTimeout + TimeoutPer100G*reqSizeIn100GiB
}

func NewPvcCreateStage(volumeId string,
	volumeType common.LvType,
	fsType common.FsType,
	format bool,
	reqSize int64,
	execNode config.Node) *PvcCreateStageRunner {
	return &PvcCreateStageRunner{
		Stage: &Stage{
			Content: &message.PvcCreateCommand{
				Format:     format,
				VolumeType: volumeType,
				VolumeId:   volumeId,
				FsType:     fsType,
			},
			SType:     PvcCreateStage,
			StartTime: 0,
			Result:    nil,
		},
		TargetNode: execNode,
		ReqSize:    reqSize,
	}
}

type PvcCreateStageConstructor struct {
}

func (c *PvcCreateStageConstructor) Construct() interface{} {
	return &PvcCreateStageRunner{
		Stage: &Stage{
			Content: &message.PvcCreateCommand{},
		},
		TargetNode: config.Node{},
	}
}

type PvcReleaseStageRunner struct {
	*Stage
}

func (s *PvcReleaseStageRunner) Run(ctx common.TraceContext) *StageExecResult {
	msg, err := message.NewMessage(message.SmsMessageHead_CMD_PVC_RELEASE_REQ, s.Content, ctx)
	if err != nil {
		return StageExecFail(err.Error())
	}
	ret := sendToAllParallel(msg, 5, 1)
	s.Result = ret
	return ret
}

func (s *PvcReleaseStageRunner) Rollback(ctx common.TraceContext) *StageExecResult {
	return StageExecFail(fmt.Errorf("umimplement fs expand rollback").Error())
}

func NewPvcReleaseStage(volumeId string,
	volumeType common.LvType,
	volumeName string) *PvcReleaseStageRunner {
	return &PvcReleaseStageRunner{
		Stage: &Stage{
			Content: &message.PvcReleaseCommand{
				Name:       volumeName,
				VolumeType: volumeType,
				VolumeId:   volumeId,
			},
			SType:     PvcReleaseStage,
			StartTime: 0,
			Result:    nil,
		},
	}
}

type PvcReleaseStageConstructor struct {
}

func (c *PvcReleaseStageConstructor) Construct() interface{} {
	return &PvcReleaseStageRunner{
		Stage: &Stage{
			Content: &message.PvcReleaseCommand{},
		},
	}
}
