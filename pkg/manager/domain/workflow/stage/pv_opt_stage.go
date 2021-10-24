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

//TODO fine tune this
type PvCreateStageRunner struct {
	*Stage
}

func (s *PvCreateStageRunner) Run(ctx common.TraceContext) *StageExecResult {
	msg, err := message.NewMessage(message.SmsMessageHead_CMD_EXPAND_FS_REQ, s.Content, ctx)
	if err != nil {
		return StageExecFail(err.Error())
	}
	ret := sendToAllSequential(msg, s.timeout(), 1)
	s.Result = ret
	return ret
}

func (s *PvCreateStageRunner) Rollback(ctx common.TraceContext) *StageExecResult {
	return StageExecFail(fmt.Errorf("umimplement fs expand rollback").Error())
}

func (s *PvCreateStageRunner) timeout() int64 {
	//todo fix
	reqSizeIn100GiB := s.Content.(*message.FsExpandCommand).ReqSize / (100 * 1024 * 1024 * 1024)
	return BaseTimeout + TimeoutPer100G*reqSizeIn100GiB
}

//todo fix this
func NewPvCreateStage(volumeId string,
	volumeType common.VolumeType,
	fsType common.FsType,
	expandSize int64) *FsExpandStageRunner {
	return &FsExpandStageRunner{
		Stage: &Stage{
			//Content: &message.FsExpandCommand{
			//	VolumeId:       volumeId,
			//	FsType:     fsType,
			//	ReqSize:    expandSize,
			//	VolumeClass: volumeType,
			//},
			SType:     PvCreateStage,
			StartTime: 0,
			Result:    nil,
		},
	}
}

type PvCreateStageConstructor struct {
}

func (c *PvCreateStageConstructor) Construct() interface{} {
	return &PvCreateStageRunner{}
}

type PvDeleteStageRunner struct {
	*Stage
}

func (s *PvDeleteStageRunner) Run(ctx common.TraceContext) *StageExecResult {
	msg, err := message.NewMessage(message.SmsMessageHead_CMD_EXPAND_FS_REQ, s.Content, ctx)
	if err != nil {
		return StageExecFail(err.Error())
	}
	ret := sendToAllSequential(msg, s.timeout(), 1)
	s.Result = ret
	return ret
}

func (s *PvDeleteStageRunner) Rollback(ctx common.TraceContext) *StageExecResult {
	return StageExecFail(fmt.Errorf("umimplement fs expand rollback").Error())
}

func (s *PvDeleteStageRunner) timeout() int64 {
	//todo fix
	reqSizeIn100GiB := s.Content.(*message.FsExpandCommand).ReqSize / (100 * 1024 * 1024 * 1024)
	return BaseTimeout + TimeoutPer100G*reqSizeIn100GiB
}

//todo fix this
func NewPvDeleteStage(volumeId string,
	volumeType common.VolumeType,
	fsType common.FsType,
	expandSize int64) *FsExpandStageRunner {
	return &FsExpandStageRunner{
		Stage: &Stage{
			//Content: &message.FsExpandCommand{
			//	VolumeId:       volumeId,
			//	FsType:     fsType,
			//	ReqSize:    expandSize,
			//	VolumeClass: volumeType,
			//},
			SType:     PvDeleteStage,
			StartTime: 0,
			Result:    nil,
		},
	}
}

type PvDeleteStageConstructor struct {
}

func (c *PvDeleteStageConstructor) Construct() interface{} {
	return &PvDeleteStageRunner{}
}

type PvExpandStageRunner struct {
	*Stage
}

func (s *PvExpandStageRunner) Run(ctx common.TraceContext) *StageExecResult {
	msg, err := message.NewMessage(message.SmsMessageHead_CMD_EXPAND_FS_REQ, s.Content, ctx)
	if err != nil {
		return StageExecFail(err.Error())
	}
	ret := sendToAllSequential(msg, s.timeout(), 1)
	s.Result = ret
	return ret
}

func (s *PvExpandStageRunner) Rollback(ctx common.TraceContext) *StageExecResult {
	return StageExecFail(fmt.Errorf("umimplement fs expand rollback").Error())
}

func (s *PvExpandStageRunner) timeout() int64 {
	//todo fix
	reqSizeIn100GiB := s.Content.(*message.FsExpandCommand).ReqSize / (100 * 1024 * 1024 * 1024)
	return BaseTimeout + TimeoutPer100G*reqSizeIn100GiB
}

//todo fix this
func NewPvExpandStage(volumeId string,
	volumeType common.VolumeType,
	fsType common.FsType,
	expandSize int64) *FsExpandStageRunner {
	return &FsExpandStageRunner{
		Stage: &Stage{
			//Content: &message.FsExpandCommand{
			//	VolumeId:       volumeId,
			//	FsType:     fsType,
			//	ReqSize:    expandSize,
			//	VolumeClass: volumeType,
			//},
			SType:     PvExpandStage,
			StartTime: 0,
			Result:    nil,
		},
	}
}

type PvExpandStageConstructor struct {
}

func (c *PvExpandStageConstructor) Construct() interface{} {
	return &PvExpandStageRunner{}
}

type PvRescanStageRunner struct {
	*Stage
}

func (s *PvRescanStageRunner) Run(ctx common.TraceContext) *StageExecResult {
	msg, err := message.NewMessage(message.SmsMessageHead_CMD_RESCAN_REQ, s.Content, ctx)
	if err != nil {
		return StageExecFail(err.Error())
	}
	ret := sendToAllParallel(msg, 30, config.AvailableNodes())
	s.Result = ret
	return ret
}

func (s *PvRescanStageRunner) Rollback(ctx common.TraceContext) *StageExecResult {
	return StageExecFail(fmt.Errorf("umimplement fs expand rollback").Error())
}

//todo fix this
func NewPvRescanStage() *PvRescanStageRunner {
	return &PvRescanStageRunner{
		Stage: &Stage{
			SType:     PvRescanStage,
			StartTime: 0,
			Result:    nil,
		},
	}
}

type PvRescanStageConstructor struct {
}

func (c *PvRescanStageConstructor) Construct() interface{} {
	return &PvRescanStageRunner{}
}
