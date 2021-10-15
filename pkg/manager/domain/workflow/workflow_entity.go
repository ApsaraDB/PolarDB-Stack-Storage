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


package workflow

import (
	"encoding/json"
	"k8s.io/apimachinery/pkg/util/uuid"
	"polardb-sms/pkg/common"
	smslog "polardb-sms/pkg/log"
	"polardb-sms/pkg/manager/domain"
	"polardb-sms/pkg/manager/domain/workflow/stage"
	"time"
)

type ExecStatus int

const (
	NotStarted ExecStatus = iota
	Started
	Success
	Fail
	SuccessRollback
	FailRollback
)

type WorkMode int

const (
	Run WorkMode = iota
	Rollback
)

type WflType int

const (
	DummyWfl WflType = iota
	Rescan
	PrCheck
	PrLock
	ClusterLvCreate
	ClusterLvExpand
	ClusterLvFormat
	ClusterLvFsExpand
	ClusterLvDelete
	ClusterLunCreate
	ClusterLunExpand
	ClusterLunFsExpand
	ClusterLunFormat
	ClusterLunFormatAndLock
	PvcCreate
	PvcRelease
	PvcFormatAndLock
	PvcFsExpand
	PvcFormat
	PvcDelete
	PvcBind
)

var DummyWorkflow = &WorkflowEntity{Id: domain.DummyWorkflowId}

type WflContext struct {
	startTime   int64
	timeout     int
	currentStep int
	Mode        WorkMode
	data        interface{}
}

type WorkflowEntity struct {
	WflContext
	Id           string
	WflType      WflType
	Step         int
	Stages       []StageRunner
	Status       ExecStatus
	LastErrMsg   string
	CreateTime   time.Time
	TraceContext common.TraceContext
	VolumeId     string
	VolumeClass  string
}

func (w *WorkflowEntity) SetTraceContext(value map[string]string) {
	if value == nil {
		w.TraceContext = map[string]string{"workflow": w.Id}
	} else {
		w.TraceContext = value
		w.TraceContext["workflow"] = w.Id
	}
}

func (w *WorkflowEntity) GetTraceContext() map[string]string {
	if w.TraceContext == nil {
		w.TraceContext = make(map[string]string)
	}
	return w.TraceContext
}

func (w *WorkflowEntity) Valid() bool {
	now := time.Now()
	overdue := w.CreateTime.Add(1 * time.Minute)
	if now.After(overdue) {
		smslog.Debugf("workflow %s overdue, ignore", w.Id)
		return false
	}
	return true
}

func (w *WorkflowEntity) IsFinished() bool {
	return w.SuccessfullyRun() || w.SuccessfullyRollback() || w.Failed()
}

func (w *WorkflowEntity) SuccessfullyRun() bool {
	return w.Status == Success && w.Step >= len(w.Stages)
}

func (w *WorkflowEntity) SuccessfullyRollback() bool {
	return w.Status == SuccessRollback && w.Step < 0
}

func (w *WorkflowEntity) Failed() bool {
	return w.Status == Fail
}

func (w *WorkflowEntity) GetExecResult() string {
	if w.SuccessfullyRun() {
		return "success run"
	}
	if w.SuccessfullyRollback() {
		return "success rollback"
	}
	if w.Failed() {
		if w.Step >= 0 {
			return w.Stages[w.Step].GetExecResult().ErrMsg
		}
		return "failed"
	}
	return "unknown status"
}

func (w *WorkflowEntity) String() string {
	msg, err := json.Marshal(w)
	if err != nil {
		return ""
	}
	return string(msg)
}

func (w *WorkflowEntity) AddStage(s StageRunner) {
	w.Stages = append(w.Stages, s)
}

func (w *WorkflowEntity) SetVolumeId(volumeId string) {
	w.VolumeId = volumeId
}

func (w *WorkflowEntity) SetVolumeClass(volumeClass string) {
	w.VolumeClass = volumeClass
}

func (w *WorkflowEntity) Run() {

}

type WflBuilder struct {
	wflType WflType
	step    int
	stages  []StageRunner
	ctx     WflContext
}

func NewWflBuilder() *WflBuilder {
	return &WflBuilder{}
}

func (b *WflBuilder) WithType(t WflType) *WflBuilder {
	b.wflType = t
	return b
}

func (b *WflBuilder) WithStageRunner(stageRunner StageRunner) *WflBuilder {
	b.stages = append(b.stages, stageRunner)
	return b
}

func (b *WflBuilder) WithStageRunners(stages []StageRunner) *WflBuilder {

	b.stages = append(b.stages, stages...)
	return b
}

func (b *WflBuilder) Build() *WorkflowEntity {
	return &WorkflowEntity{
		Id:         string(uuid.NewUUID()),
		WflType:    b.wflType,
		Step:       0,
		Stages:     b.stages,
		WflContext: b.ctx,
	}
}

type StageRunner interface {
	//TODO remove this method
	GetExecResult() *stage.StageExecResult
	Run(ctx common.TraceContext) *stage.StageExecResult
	Rollback(ctx common.TraceContext) *stage.StageExecResult
	StageType() stage.StageType
}
