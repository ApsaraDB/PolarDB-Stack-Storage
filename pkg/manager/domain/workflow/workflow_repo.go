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
	"polardb-sms/pkg/manager/domain/repository"
)

type WorkflowRepo struct {
	*repository.BaseDB
}

func NewWorkflowRepo() *WorkflowRepo {
	return &WorkflowRepo{
		BaseDB: repository.GetBaseDB(),
	}
}

func (r *WorkflowRepo) Save(wfl *Workflow) (int64, error) {
	result, err := r.Engine.Where("workflow_id=?", wfl.WorkflowId).Update(wfl)
	if err != nil {
		return 0, err
	}

	return result, nil
}

func (r *WorkflowRepo) Create(wfl *Workflow) (int64, error) {
	affected, err := r.Engine.Insert(wfl)
	if err != nil {
		return 0, err
	}
	return affected, nil
}

func (r *WorkflowRepo) Delete(wfl *Workflow) (int64, error) {
	affected, err := r.Engine.Delete(wfl)
	if err != nil {
		return 0, err
	}
	return affected, nil
}

func (r *WorkflowRepo) FindByWorkflowId(wid string) (*Workflow, error) {
	wfl := Workflow{
		WorkflowId: wid,
	}

	exist, err := r.Engine.Alias("a").Where("a.workflow_id=?", wid).Get(&wfl)
	if err != nil {
		return nil, err
	}

	if !exist {
		return nil, fmt.Errorf("workflow %s not exist", wid)
	}

	return &wfl, nil
}

func (r *WorkflowRepo) FindAll() ([]*Workflow, error) {
	var wfls []*Workflow
	err := r.Engine.Find(&wfls)
	return wfls, err
}

/**
return total_size
*/
func (r *WorkflowRepo) FindByPage(idx, pgSize int) ([]*Workflow, int64, error) {
	var wfls []*Workflow
	cnt, err := r.Engine.Limit(pgSize, idx*pgSize).FindAndCount(&wfls)
	return wfls, cnt, err
}

func (r *WorkflowRepo) FindByConditionsAndLimit(c string, limit int) ([]*Workflow, error) {
	var wfls []*Workflow
	err := r.Engine.Where(c).Limit(limit).Find(&wfls)
	return wfls, err
}
