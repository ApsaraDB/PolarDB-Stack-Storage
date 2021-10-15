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
	"polardb-sms/pkg/manager/domain"
	"polardb-sms/pkg/manager/domain/repository"
	"time"
)

type WorkflowRepositoryImpl struct {
	dataConverter   domain.Converter
	dataWorkflowDAO *WorkflowRepo
	*repository.BaseDB
}

func (c *WorkflowRepositoryImpl) Create(workflowEntity *WorkflowEntity) (int64, error) {
	mInf, err := c.dataConverter.ToModel(workflowEntity)
	if err != nil {
		return 0, err
	}
	m := mInf.(*Workflow)
	m.Created = time.Now()

	if _, err := c.dataWorkflowDAO.Create(m); err != nil {
		return 0, err
	}
	return 0, nil
}

func (c *WorkflowRepositoryImpl) Save(workflowEntity *WorkflowEntity) (int64, error) {
	mInf, err := c.dataConverter.ToModel(workflowEntity)
	if err != nil {
		return 0, err
	}
	m := mInf.(*Workflow)
	m.Updated = time.Now()
	if _, err := c.dataWorkflowDAO.Save(m); err != nil {
		return 0, err
	}
	return 0, nil
}

func (c *WorkflowRepositoryImpl) Delete(entity *WorkflowEntity) (int64, error) {
	mInf, err := c.dataConverter.ToModel(entity)
	if err != nil {
		return 0, err
	}
	m := mInf.(*Workflow)
	if _, err := c.dataWorkflowDAO.Delete(m); err != nil {
		return 0, err
	}
	return 0, nil
}

func (c *WorkflowRepositoryImpl) FindByWorkflowId(workflowId string) (*WorkflowEntity, error) {
	m, err := c.dataWorkflowDAO.FindByWorkflowId(workflowId)
	if err != nil {
		return nil, err
	}

	eInf, err := c.dataConverter.ToEntity(m)
	if err != nil {
		return nil, err
	}
	e := eInf.(*WorkflowEntity)
	return e, nil
}

func (c *WorkflowRepositoryImpl) FindAll() ([]*WorkflowEntity, error) {
	workflows, err := c.dataWorkflowDAO.FindAll()
	if err != nil {
		return nil, err
	}

	var wfs = make([]interface{}, len(workflows))
	for i, d := range workflows {
		wfs[i] = d
	}

	var es []*WorkflowEntity
	esi, _ := c.dataConverter.ToEntities(wfs)
	for _, e := range esi {
		es = append(es, e.(*WorkflowEntity))
	}

	return es, nil
}

func (c *WorkflowRepositoryImpl) FindByPage(idx, pgSize int) ([]*WorkflowEntity, int64, error) {
	workflows, cnt, err := c.dataWorkflowDAO.FindByPage(idx, pgSize)
	if err != nil {
		return nil, 0, err
	}

	var wfs = make([]interface{}, len(workflows))
	for i, d := range workflows {
		wfs[i] = d
	}

	var es []*WorkflowEntity
	esi, _ := c.dataConverter.ToEntities(wfs)
	for _, e := range esi {
		es = append(es, e.(*WorkflowEntity))
	}

	return es, cnt, nil
}

func (c *WorkflowRepositoryImpl) FindByStatusAndLimit(status, limit int) ([]*WorkflowEntity, error) {
	condition := fmt.Sprintf("status=%d", status)
	workflows, err := c.dataWorkflowDAO.FindByConditionsAndLimit(condition, limit)
	if err != nil {
		return nil, err
	}
	var wfs = make([]interface{}, len(workflows))
	for i, d := range workflows {
		wfs[i] = d
	}
	var es []*WorkflowEntity
	esi, _ := c.dataConverter.ToEntities(wfs)
	for _, e := range esi {
		es = append(es, e.(*WorkflowEntity))
	}
	return es, nil
}

func (c *WorkflowRepositoryImpl) FindByVolumeIdAndClass(volumeId string, volumeClass string, wflType int) (*WorkflowEntity, error) {
	ret := &Workflow{}
	ok, err := c.Engine.Alias("a").Unscoped().
		Where("a.volume_id=? and a.volume_class=? and a.type=?", volumeId, volumeClass, wflType).
		Desc("created").
		Get(ret)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("can not find the workflow for id [%s] class [%s] type [%d]", volumeId, volumeClass, wflType)
	}
	wflEntity, err := c.dataConverter.ToEntity(ret)
	if err != nil {
		return nil, err
	}
	return wflEntity.(*WorkflowEntity), nil
}

func NewWorkflowRepository() WorkflowRepository {
	workflowRepo := &WorkflowRepositoryImpl{
		dataConverter:   NewWorkflowConverter(),
		dataWorkflowDAO: NewWorkflowRepo(),
		BaseDB:          repository.GetBaseDB(),
	}
	return workflowRepo
}

type WorkflowRepository interface {
	Create(workflowEntity *WorkflowEntity) (int64, error)
	Save(workflowEntity *WorkflowEntity) (int64, error)
	Delete(entity *WorkflowEntity) (int64, error)
	FindByWorkflowId(workflowId string) (*WorkflowEntity, error)
	FindAll() ([]*WorkflowEntity, error)
	FindByPage(idx, pgSize int) ([]*WorkflowEntity, int64, error)
	FindByStatusAndLimit(status, limit int) ([]*WorkflowEntity, error)
	FindByVolumeIdAndClass(volumeId, volumeClass string, wflType int) (*WorkflowEntity, error)
}
