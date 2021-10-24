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
	"fmt"
	"polardb-sms/pkg/common"
	smslog "polardb-sms/pkg/log"
	"polardb-sms/pkg/manager/domain"
	"polardb-sms/pkg/manager/domain/workflow/stage"
)

var _ domain.Converter = &WorkflowConverter{}

type WorkflowConverter struct {
	stageConstructors map[string]StageConstructor
}

type innerStage struct {
	StageType    string `json:"stage_type"`
	StageContent string `json:"stage_content"`
}

type innerStages struct {
	Stages []*innerStage `json:"stages"`
}

func (c *WorkflowConverter) ToModel(t interface{}) (interface{}, error) {
	e := t.(*WorkflowEntity)
	m := Workflow{
		WorkflowId:   e.Id,
		Type:         int(e.WflType),
		Step:         e.Step,
		Status:       int(e.Status),
		Mode:         int(e.Mode),
		LastErrMsg:   e.LastErrMsg,
		TraceContext: e.TraceContext.String(),
		VolumeId:     e.VolumeId,
		VolumeClass:  e.VolumeClass,
	}

	innerSts := &innerStages{Stages: make([]*innerStage, 0)}
	for _, st := range e.Stages {
		innerSt := &innerStage{}
		innerSt.StageType = string(st.StageType())
		stageBytes, err := json.Marshal(st)
		if err != nil {
			smslog.Infof("WorkflowEntity to model marshal with %s: %v", st, err)
			return nil, err
		}
		innerSt.StageContent = string(stageBytes)
		innerSts.Stages = append(innerSts.Stages, innerSt)
	}

	stagesBytes, err := json.Marshal(innerSts)
	if err != nil {
		smslog.Infof("WorkflowEntity to model marshal with %s: %v", innerSts, err)
		return nil, err
	}
	m.Stages = string(stagesBytes)
	return &m, nil
}

func (c *WorkflowConverter) ToEntity(t interface{}) (interface{}, error) {
	m := t.(*Workflow)
	e := WorkflowEntity{
		Id:         m.WorkflowId,
		WflType:    WflType(m.Type),
		Step:       m.Step,
		Status:     ExecStatus(m.Status),
		LastErrMsg: m.LastErrMsg,
		WflContext: WflContext{
			Mode: WorkMode(m.Mode),
		},
		CreateTime:   m.Created,
		TraceContext: common.ParseForTraceContext(m.TraceContext),
		VolumeClass:  m.VolumeClass,
		VolumeId:     m.VolumeId,
	}

	innerSts := &innerStages{Stages: make([]*innerStage, 0)}
	if err := json.Unmarshal([]byte(m.Stages), innerSts); err != nil {
		smslog.Infof("WorkflowEntity to entity unmarshal with %s: %v", m.Stages, err)
		return nil, err
	}

	var stageList []StageRunner
	for _, st := range innerSts.Stages {
		stageRunner, err := c.convertToStage(st.StageType,
			[]byte(st.StageContent))
		if err != nil {
			return nil, err
		}
		stageList = append(stageList, stageRunner)
	}
	e.Stages = stageList
	return &e, nil
}

func (c *WorkflowConverter) ToEntities(ds []interface{}) ([]interface{}, error) {
	var es []interface{}
	for _, m := range ds {
		e, err := c.ToEntity(m.(*Workflow))
		if err != nil {
			smslog.Debugf("WorkflowConverter: ToEntities err %s", err.Error())
			continue
		}
		es = append(es, e.(*WorkflowEntity))
	}
	return es, nil
}

func (c *WorkflowConverter) convertToStage(stageType string, content []byte) (StageRunner, error) {
	stageRunnerConstructor, ok := c.stageConstructors[stageType]
	if !ok {
		return nil, fmt.Errorf("can not find stage constructor for stage %s", stageType)
	}

	stageRunner := stageRunnerConstructor.Construct()
	err := common.BytesToStruct(content, stageRunner)
	if err != nil {
		return nil, err
	}
	return stageRunner.(StageRunner), nil
}

type StageConstructor interface {
	//should be StageRunner, but go maybe import cycle
	Construct() interface{}
}

func NewWorkflowConverter() domain.Converter {
	return &WorkflowConverter{
		stageConstructors: map[string]StageConstructor{
			stage.FsExpandStage:   &stage.FsExpandStageConstructor{},
			stage.FsFormatStage:   &stage.FsFormatStageConstructor{},
			stage.PvcCreateStage:  &stage.PvcCreateStageConstructor{},
			stage.PvcReleaseStage: &stage.PvcReleaseStageConstructor{},
			stage.PvCreateStage:   &stage.PvCreateStageConstructor{},
			stage.PvDeleteStage:   &stage.PvDeleteStageConstructor{},
			stage.PvRescanStage:   &stage.PvRescanStageConstructor{},
			stage.PvExpandStage:   &stage.PvExpandStageConstructor{},
			stage.LvCreateStage:   &stage.LvCreateStageConstructor{},
			stage.LvDeleteStage:   &stage.LvDeleteStageConstructor{},
			stage.LvExpandStage:   &stage.LvExpandStageConstructor{},
			stage.DmExecStage:     &stage.DmExecStageConstructor{},
			stage.PrBatchStage:    &stage.PrBatchStageConstructor{},
			stage.PrStage:         &stage.PrStageConstructor{},
			stage.DBPersistStage:  &stage.DBPersistStageConstructor{},
		},
	}
}
