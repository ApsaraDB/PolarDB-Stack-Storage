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
	"flag"
	"github.com/stretchr/testify/assert"
	smslog "polardb-sms/pkg/log"
	_ "polardb-sms/pkg/manager/domain/repository/mysql"
	"testing"
)

func TestWorkflowRepo_Create(t *testing.T) {
	flag.Parse()
	testCase := assert.New(t)
	var r = NewWorkflowRepo()
	wfl := &Workflow{
		LastErrMsg: "",
		Mode:       0,
		Stages:     "{}",
		Status:     0,
		Step:       0,
		Type:       0,
		WorkflowId: "test_02",
	}
	affected, err := r.Create(wfl)
	testCase.NoError(err)
	testCase.Equal(int64(1), affected)
}

func TestWorkflowRepo_FindByPage(t *testing.T) {
	flag.Parse()
	testCase := assert.New(t)
	var r = NewWorkflowRepo()

	//for i := 0; i < 100; i++ {
	//	wfl := &models.Workflow{
	//		LastErrMsg: "",
	//		Mode:       0,
	//		Stages:     "{}",
	//		Status:     0,
	//		Step:       0,
	//		Type:       0,
	//		WorkflowId: "test_02",
	//	}
	//
	//	wfl.WorkflowId = fmt.Sprintf("test_000%d", i)
	//	_, err := r.Create(wfl)
	//	testCase.NoError(err)
	//}

	ret, size, err := r.FindByPage(0, 20)
	testCase.NoError(err)
	smslog.Infof("result size  %d, totoal size %d", len(ret), size)
	for _, w := range ret {
		smslog.Infof("workflow %v", w)
	}

	ret, size, err = r.FindByPage(1, 20)
	testCase.NoError(err)
	smslog.Infof("result size  %d, totoal size %d", len(ret), size)
	for _, w := range ret {
		smslog.Infof("workflow %v", w)
	}

	ret, size, err = r.FindByPage(2, 20)
	testCase.NoError(err)
	smslog.Infof("result size  %d, totoal size %d", len(ret), size)
	for _, w := range ret {
		smslog.Infof("workflow %v", w)
	}
}

func TestWorkflowRepo_FindByConditionsAndLimit(t *testing.T) {
	testCase := assert.New(t)
	var r = NewWorkflowRepo()

	//for i := 0; i < 1000; i++ {
	//	wfl := &models.Workflow{
	//		LastErrMsg: "",
	//		Mode:       0,
	//		Stages:     "{}",
	//		Status:     i % 4,
	//		Step:       0,
	//		Type:       0,
	//		WorkflowId: "test_02",
	//	}
	//
	//	wfl.WorkflowId = fmt.Sprintf("test_1000%d", i)
	//	_, err := r.Create(wfl)
	//	testCase.NoError(err)
	//}

	ret, err := r.FindByConditionsAndLimit("status=2", 5)
	testCase.NoError(err)
	for _, w := range ret {
		smslog.Infof("workflow %v", w)
	}
}
