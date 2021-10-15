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


package agent

import (
	"fmt"
	"polardb-sms/pkg/manager/domain"
	"polardb-sms/pkg/manager/domain/repository"
	"sync"
)

type ClusterAgentRepositoryImpl struct {
	dataConverter domain.Converter
	*repository.BaseDB
}

func (c *ClusterAgentRepositoryImpl) Create(agentEntity *ClusterAgentEntity) error {
	agentModel, err := c.dataConverter.ToModel(agentEntity)
	if err != nil {
		return err
	}
	_, err = c.Engine.Insert(agentModel)
	if err != nil {
		return err
	}
	return nil
}

func (c *ClusterAgentRepositoryImpl) FindByAgentId(agentId string) (*ClusterAgentEntity, error) {
	agentModel := &ClusterAgent{
		AgentId: agentId,
	}
	exist, err := c.Engine.Alias("a").Where("a.agent_id=?", agentId).Get(agentModel)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, fmt.Errorf("agent %s not exist", agentId)
	}

	eInf, err := c.dataConverter.ToEntity(agentModel)
	if err != nil {
		return nil, err
	}
	return eInf.(*ClusterAgentEntity), nil
}

func (c *ClusterAgentRepositoryImpl) Save(agentEntity *ClusterAgentEntity) error {
	agentModelInf, err := c.dataConverter.ToModel(agentEntity)
	if err != nil {
		return err
	}
	agentModel := agentModelInf.(*ClusterAgent)
	_, err = c.Engine.
		Alias("a").Where("a.agent_id=?", agentEntity.AgentId).Update(agentModel)
	if err != nil {
		return err
	}
	return nil
}

func (c *ClusterAgentRepositoryImpl) QueryAll() ([]*ClusterAgentEntity, error) {
	var agents []*ClusterAgent
	err := c.Engine.Find(&agents)
	if err != nil {
		return nil, err
	}

	var as = make([]interface{}, len(agents))
	for i, a := range agents {
		as[i] = a
	}

	var es []*ClusterAgentEntity
	esi, _ := c.dataConverter.ToEntities(as)
	for _, e := range esi {
		es = append(es, e.(*ClusterAgentEntity))
	}

	return es, nil
}

func (c *ClusterAgentRepositoryImpl) Delete(agentId string) error {
	_, err := c.Engine.Alias("a").
		Where("a.agent_id=?", agentId).
		Delete(&ClusterAgent{})
	if err != nil {
		return err
	}
	return nil
}

func GetClusterAgentRepository() ClusterAgentRepository {
	_agentRepoOnce.Do(func() {
		if _agentRepo == nil {
			agRepo := &ClusterAgentRepositoryImpl{
				dataConverter: NewClusterAgentConverter(),
				BaseDB:        repository.GetBaseDB(),
			}
			agRepo.CacheTable(&ClusterAgent{})
			_agentRepo = agRepo
		}
	})
	return _agentRepo
}

var (
	_agentRepo     ClusterAgentRepository
	_agentRepoOnce sync.Once
)

type ClusterAgentRepository interface {
	Create(agentEntity *ClusterAgentEntity) error
	FindByAgentId(agentId string) (*ClusterAgentEntity, error)
	Save(agentEntity *ClusterAgentEntity) error
	QueryAll() ([]*ClusterAgentEntity, error)
	Delete(agentId string) error
}

