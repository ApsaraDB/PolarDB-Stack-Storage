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
	"polardb-sms/pkg/manager/domain"
)

var _ domain.Converter = &ClusterAgentConverter{}

type ClusterAgentConverter struct {
}

func (converter *ClusterAgentConverter) ToModel(t interface{}) (interface{}, error) {
	e := t.(*ClusterAgentEntity)
	m := ClusterAgent{
		AgentId:   e.AgentId,
		ClusterId: e.ClusterId,
		Ip:        e.Ip,
		Port:      e.Port,
		Online:    e.Online,
		Updated:   e.LastUpdate,
	}
	return &m, nil
}

func (converter *ClusterAgentConverter) ToEntity(t interface{}) (interface{}, error) {
	m := t.(*ClusterAgent)
	e := ClusterAgentEntity{
		AgentId:    m.AgentId,
		ClusterId:  m.ClusterId,
		Ip:         m.Ip,
		Port:       m.Port,
		Online:     m.Online,
		LastUpdate: m.Updated,
	}
	return &e, nil
}

func (converter *ClusterAgentConverter) ToEntities(ds []interface{}) ([]interface{}, error) {
	var es []interface{}
	for _, t := range ds {
		e, _ := converter.ToEntity(t.(*ClusterAgent))
		es = append(es, e.(*ClusterAgentEntity))
	}
	return es, nil
}

func NewClusterAgentConverter() domain.Converter {
	return &ClusterAgentConverter{}
}
