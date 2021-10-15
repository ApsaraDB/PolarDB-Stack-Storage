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


package pv

import (
	"fmt"
	smslog "polardb-sms/pkg/log"
	"polardb-sms/pkg/manager/domain"
	"polardb-sms/pkg/manager/domain/repository"
	"sync"
)

type PhysicalVolumeRepositoryImpl struct {
	*repository.BaseDB
	dataConverter domain.Converter
}

func (p PhysicalVolumeRepositoryImpl) Create(pv *PhysicalVolumeEntity) (int64, error) {
	pvModel, err := p.dataConverter.ToModel(pv)
	if err != nil {
		return 0, err
	}
	affected, err := p.Engine.Insert(pvModel)
	if err != nil {
		return 0, err
	}
	return affected, nil
}

func (p PhysicalVolumeRepositoryImpl) Save(pv *PhysicalVolumeEntity) (int64, error) {
	pvModelInf, err := p.dataConverter.ToModel(pv)
	if err != nil {
		return 0, err
	}
	pvModel := pvModelInf.(*PhysicalVolume)
	result, err := p.Engine.Alias("a").
		Where("a.volume_id=? and a.node_id=?", pvModel.VolumeId, pvModel.NodeId).
		Update(pvModel)
	if err != nil {
		return 0, err
	}
	return result, nil
}

func (p PhysicalVolumeRepositoryImpl) FindByVolumeIdAndNodeId(volumeId string, nodeId string) (*PhysicalVolumeEntity, error) {
	pvModel := &PhysicalVolume{
		VolumeId: volumeId,
	}
	exist, err := p.Engine.Alias("a").
		Where("a.volume_id=? and a.node_id=?", volumeId, nodeId).
		Get(pvModel)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, nil
	}
	pvInf, err := p.dataConverter.ToEntity(pvModel)
	if err != nil {
		return nil, err
	}
	return pvInf.(*PhysicalVolumeEntity), nil
}

func (p PhysicalVolumeRepositoryImpl) FindByVolumeIds(volumeIds []string) ([]*PhysicalVolumeEntity, error) {
	var pvs []*PhysicalVolume
	err := p.Engine.In("volume_id", volumeIds).Find(&pvs)
	if err != nil {
		smslog.Infof("Error when FindByWwids %v", volumeIds)
		return nil, err
	}
	var pvEntities []*PhysicalVolumeEntity
	for _, pv := range pvs {
		pvEntityInf, err := p.dataConverter.ToEntity(pv)
		if err != nil {
			smslog.Infof("convert pvModel %v to pvEntity err %s", pv, err.Error())
			continue
		}
		pvEntities = append(pvEntities, pvEntityInf.(*PhysicalVolumeEntity))
	}
	return pvEntities, nil
}

func (p PhysicalVolumeRepositoryImpl) DeleteByVolumeId(volumeId string) (int64, error) {
	affected, err := p.Engine.Alias("a").
		Where("a.volume_id=?", volumeId).
		Delete(&PhysicalVolumeEntity{})
	if err != nil {
		return 0, err
	}
	return affected, nil
}

func (p PhysicalVolumeRepositoryImpl) QueryAll() ([]*PhysicalVolumeEntity, error) {
	return nil, fmt.Errorf("unimplemented query all")
}

func GetPhysicalVolumeRepository() PhysicalVolumeRepository {
	_pvRepoOnce.Do(func() {
		if _pvRepo == nil {
			impl := &PhysicalVolumeRepositoryImpl{
				BaseDB:        repository.GetBaseDB(),
				dataConverter: NewPhysicalVolumeConverter(),
			}
			_pvRepo = impl
		}
	})
	return _pvRepo
}

var (
	_pvRepo     PhysicalVolumeRepository
	_pvRepoOnce sync.Once
)

type PhysicalVolumeRepository interface {
	Create(pv *PhysicalVolumeEntity) (int64, error)
	Save(pv *PhysicalVolumeEntity) (int64, error)
	FindByVolumeIdAndNodeId(volumeId string, nodeId string) (*PhysicalVolumeEntity, error)
	FindByVolumeIds(volumeIds []string) ([]*PhysicalVolumeEntity, error)
	DeleteByVolumeId(volumeId string) (int64, error)
	QueryAll() ([]*PhysicalVolumeEntity, error)
}
