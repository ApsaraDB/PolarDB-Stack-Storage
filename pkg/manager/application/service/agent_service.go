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

package service

import (
	"fmt"
	smslog "polardb-sms/pkg/log"
	"polardb-sms/pkg/manager/application/assembler"
	globalConfig "polardb-sms/pkg/manager/config"
	"polardb-sms/pkg/manager/domain/agent"
	"time"
)

type AgentService struct {
	agentRepo agent.ClusterAgentRepository
}

func (s *AgentService) Heartbeat(req *assembler.HeartbeatRequest) (string, error) {
	agentEntity, err := s.agentRepo.FindByAgentId(req.AgentId)
	if err != nil {
		agentEntity = &agent.ClusterAgentEntity{
			AgentId:   req.AgentId,
			ClusterId: "default",
			Ip:        req.NodeIp,
			Online:    0,
			Port:      req.Port,
		}
		err = s.agentRepo.Create(agentEntity)
		if err != nil {
			err := fmt.Errorf("create agent %s err %s", req.AgentId, err.Error())
			smslog.Error(err.Error())
			return "", err
		}
		return "create successful", nil
	}
	agentEntity.LastUpdate = time.Now()
	globalConfig.UpdateAgentNodeHeartbeatTime(agentEntity.AgentId)
	_ = s.agentRepo.Save(agentEntity)
	return "update successful", nil
}

func NewAgentService() *AgentService {
	return &AgentService{
		agentRepo: agent.GetClusterAgentRepository(),
	}
}
