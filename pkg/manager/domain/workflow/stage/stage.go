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
	"polardb-sms/pkg/network/message"
)

type StageType string

const (
	UnStageType     StageType = "Non"
	FsExpandStage             = "fs-expand"
	FsFormatStage             = "fs-format"
	PrStage                   = "pr"
	PrBatchStage              = "pr-batch"
	DmExecStage               = "dm-exec"
	LvCreateStage             = "lv-create"
	LvDeleteStage             = "lv-delete"
	LvExpandStage             = "lv-expand"
	PvCreateStage             = "pv-create"
	PvDeleteStage             = "pv-delete"
	PvExpandStage             = "pv-expand"
	PvRescanStage             = "pv-rescan"
	PvcCreateStage            = "pvc-create"
	PvcReleaseStage           = "pvc-release"
	DBPersistStage            = "db-persist"
)

type StageExecStatus int

const (
	StageSuccess StageExecStatus = iota
	StageFail
)

type TargetAgent struct {
	ClusterId string `json:"cluster_id"`
	Id        string `json:"id"`
	Ip        string `json:"ip"`
}

type StageContext struct {
	TargetAgent
	Content interface{} `json:"content"`
}

type StageExecResult struct {
	ExecStatus StageExecStatus
	ErrMsg     string
	Content    []byte
}

func (r *StageExecResult) IsSuccess() bool {
	if r != nil && r.ExecStatus == StageSuccess {
		return true
	}
	return false
}

func StageExecFail(errMsg string) *StageExecResult {
	return &StageExecResult{
		ExecStatus: StageFail,
		ErrMsg:     errMsg,
	}
}

func StageExecSuccess(content []byte) *StageExecResult {
	return &StageExecResult{
		ExecStatus: StageSuccess,
		Content:    content,
	}
}

func FromMessageBody(execResult *message.MessageBody) *StageExecResult {
	result := &StageExecResult{
		ExecStatus: StageExecStatus(execResult.ExecStatus),
		ErrMsg:     execResult.ErrMsg,
		Content:    execResult.Content,
	}
	return result
}

type Stage struct {
	SType     StageType        `json:"s_type"`
	StartTime int64            `json:"start_time"`
	Result    *StageExecResult `json:"result"`
	Content   interface{}      `json:"content"`
}

func (s Stage) StageType() StageType {
	return s.SType
}

func (s Stage) GetExecResult() *StageExecResult {
	return s.Result
}
