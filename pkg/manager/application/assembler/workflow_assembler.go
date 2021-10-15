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


package assembler

import (
	"polardb-sms/pkg/manager/application/view"
	"polardb-sms/pkg/manager/domain/workflow"
)

type WorkflowAssembler interface {
	ToWorkflowEntity(v *view.WorkflowResponse) *workflow.WorkflowEntity
	ToWorkflowEntities(vs []*view.WorkflowResponse) []*workflow.WorkflowEntity
	ToWorkflowView(e *workflow.WorkflowEntity) *view.WorkflowResponse
	ToWorkflowViews(es []*workflow.WorkflowEntity) []*view.WorkflowResponse
}

type WorkflowAssemblerImpl struct {
}

func (as *WorkflowAssemblerImpl) ToWorkflowEntity(v *view.WorkflowResponse) *workflow.WorkflowEntity {
	e := workflow.WorkflowEntity{
		Id:      v.WorkflowId,
		WflType: workflow.WflType(v.Type),
		Step:    v.Step,
		Status:  workflow.ExecStatus(v.Status),
		WflContext: workflow.WflContext{
			Mode: workflow.WorkMode(v.Mode),
		},
	}
	// TODO stages
	return &e
}

func (as *WorkflowAssemblerImpl) ToWorkflowEntities(vs []*view.WorkflowResponse) []*workflow.WorkflowEntity {
	var es []*workflow.WorkflowEntity
	for _, v := range vs {
		e := as.ToWorkflowEntity(v)
		es = append(es, e)
	}
	return es
}

func (as *WorkflowAssemblerImpl) ToWorkflowView(e *workflow.WorkflowEntity) *view.WorkflowResponse {
	v := view.WorkflowResponse{
		WorkflowId: e.Id,
		Type:       int(e.WflType),
		Step:       e.Step,
		Status:     int(e.Status),
		Mode:       int(e.Mode),
	}
	// TODO stages
	return &v
}

func (as *WorkflowAssemblerImpl) ToWorkflowViews(es []*workflow.WorkflowEntity) []*view.WorkflowResponse {
	var vs []*view.WorkflowResponse
	for _, e := range es {
		v := as.ToWorkflowView(e)
		vs = append(vs, v)
	}
	return vs
}

func NewWorkflowAssembler() WorkflowAssembler {
	as := &WorkflowAssemblerImpl{}
	return as
}
