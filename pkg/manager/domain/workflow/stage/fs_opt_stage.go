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

const (
	TimeoutPer100G = 10 // time.Second
	BaseTimeout    = 20 // time.Second
)

type FsExpandStageRunner struct {
	*Stage
	TargetNode *config.Node
}

func (s *FsExpandStageRunner) Run(ctx common.TraceContext) *StageExecResult {
	msg, err := message.NewMessage(message.SmsMessageHead_CMD_EXPAND_FS_REQ, s.Content, ctx)
	if err != nil {
		return StageExecFail(err.Error())
	}
	ret := sendAndWait(msg, s.TargetNode.Name, s.timeout())
	s.Result = ret
	return ret
}

func (s *FsExpandStageRunner) Rollback(ctx common.TraceContext) *StageExecResult {
	return StageExecFail(fmt.Errorf("umimplement fs expand rollback").Error())
}

func (s *FsExpandStageRunner) timeout() int64 {
	//todo fix
	reqSizeIn100GiB := s.Content.(*message.FsExpandCommand).ReqSize / (100 * 1024 * 1024 * 1024)
	return BaseTimeout + TimeoutPer100G*reqSizeIn100GiB
}

func NewFsExpandStage(volumeId string,
	volumeType common.LvType,
	fsType common.FsType,
	expandSize int64,
	originSize int64,
	execNode *config.Node) *FsExpandStageRunner {
	return &FsExpandStageRunner{
		Stage: &Stage{
			Content: &message.FsExpandCommand{
				VolumeId:   volumeId,
				FsType:     fsType,
				ReqSize:    expandSize,
				OriginSize: originSize,
				VolumeType: volumeType,
			},
			SType:     FsExpandStage,
			StartTime: 0,
			Result:    nil,
		},
		TargetNode: execNode,
	}
}

type FsExpandStageConstructor struct {
}

func (c *FsExpandStageConstructor) Construct() interface{} {
	return &FsExpandStageRunner{
		Stage: &Stage{
			Content: &message.FsExpandCommand{},
		},
	}
}

type FsFormatStageRunner struct {
	*Stage
	TargetNode *config.Node `json:"target_node"`
}

func (s *FsFormatStageRunner) Run(ctx common.TraceContext) *StageExecResult {
	msg, err := message.NewMessage(message.SmsMessageHead_CMD_FORMAT_FS_REQ, s.Content, ctx)
	if err != nil {
		return StageExecFail(err.Error())
	}
	ret := sendAndWait(msg, s.TargetNode.Name, s.timeout())
	s.Result = ret
	return ret
}

func (s *FsFormatStageRunner) Rollback(ctx common.TraceContext) *StageExecResult {
	return StageExecFail(fmt.Errorf("umimplement fs format rollback").Error())
}

func (s *FsFormatStageRunner) timeout() int64 {
	//todo fix
	reqSizeIn100GiB := s.Content.(*message.FsFormatCommand).ReqSize / (100 * 1024 * 1024 * 1024)
	return BaseTimeout + TimeoutPer100G*reqSizeIn100GiB
}

func NewFsFormatStage(volumeId string,
	volumeType common.LvType,
	fsType common.FsType,
	reqSize int64,
	execNode *config.Node) *FsFormatStageRunner {
	return &FsFormatStageRunner{
		Stage: &Stage{
			Content: &message.FsFormatCommand{
				VolumeId:   volumeId,
				VolumeType: volumeType,
				FsType:     fsType,
				ReqSize:    reqSize,
			},
			SType:     FsFormatStage,
			StartTime: 0,
			Result:    nil,
		},
		TargetNode: execNode,
	}
}

type FsFormatStageConstructor struct {
}

func (c *FsFormatStageConstructor) Construct() interface{} {
	return &FsFormatStageRunner{
		Stage: &Stage{
			Content: &message.FsFormatCommand{},
		},
		TargetNode: &config.Node{},
	}
}
