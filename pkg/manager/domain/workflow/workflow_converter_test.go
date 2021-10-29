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
	"fmt"
	"polardb-sms/pkg/common"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWorkflowConverter_ToModel(t *testing.T) {
	e := &WorkflowEntity{
		Id:      "123",
		WflType: ClusterLvFsExpand,
		Step:    1,
		Stages:  []StageRunner{
			//&SimpleStage{
			//	Stage: &stage.Stage{
			//		SType: stage.ClusterLunExpandStage,
			//		StageContext: stage.StageContext{
			//			Content: []byte(`{"name": "test-01", "wwid":"1024", "path":"/dev/mapper/1024", "reqSize": 1024}`),
			//			TargetAgent: stage.TargetAgent{
			//				Id: "1111",
			//				Ip: "2222",
			//			},
			//		},
			//		StartTime: time.Now().Unix(),
			//	},
			//},
			//&SimpleStage{
			//	Stage: &stage.Stage{
			//		SType: stage.ClusterLunExpandStage,
			//		StageContext: stage.StageContext{
			//			Content: []byte(`{"name": "test-02", "wwid":"2048", "path":"/dev/mapper/2048", "reqSize": 2048}`),
			//		},
			//		StartTime: time.Now().Unix(),
			//	},
			//},
			//stage.NewFsExpandStage("ade", common.Lv, common.Pfs, 10240000000),
		},
		WflContext: WflContext{
			Mode: Run,
		},
		Status: Started,
	}
	wc := &WorkflowConverter{}
	mi, _ := wc.ToModel(e)
	m := mi.(*Workflow)
	mbytes, _ := common.StructToBytes(m)
	fmt.Printf("%s", string(mbytes))
	assert.NotEmpty(t, m)
	ei, _ := wc.ToEntity(m)
	ew := ei.(*WorkflowEntity)
	_ = fmt.Sprintf("%v", ew)
}

func TestWorkflowConverter_ToEntity(t *testing.T) {
	m := Workflow{
		Id:         123,
		LastErrMsg: "no err msg",
		Mode:       0,
		Stages: `[{
	"SType": 0,
	"Ctx": "{\"name\": \"test-01\", \"wwid\":\"1024\", \"path\":\"/dev/mapper/1024\", \"reqSize\": 1024}",
	"StartTime": 1606906176,
	"Result": {
		"Code": 1,
		"Message": "ext4 expand starting"
	}
}, {
	"SType": 1,
	"Ctx": "{\"name\": \"test-02\", \"wwid\":\"2048\", \"path\":\"/dev/mapper/2048\", \"reqSize\": 2048}",
	"StartTime": 1606906176,
	"Result": {
		"Code": 1,
		"Message": "pfs expand starting"
	}
}]`,
		Step:       1,
		Type:       8,
		WorkflowId: "123",
		Created:    time.Now(),
	}

	wc := &WorkflowConverter{}
	ei, _ := wc.ToEntity(m)
	e := ei.(*WorkflowEntity)
	assert.NotEmpty(t, e)
}
