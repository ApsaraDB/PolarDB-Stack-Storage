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


package lv

import (
	"polardb-sms/pkg/common"
	smslog "polardb-sms/pkg/log"
	"polardb-sms/pkg/manager/domain"
	"polardb-sms/pkg/manager/domain/repository"
	"strings"
	"sync"
)

type LvRepositoryImpl struct {
	dataConverter domain.Converter
	*repository.BaseDB
}

func (c *LvRepositoryImpl) Create(lvEntity *LogicalVolumeEntity) (int64, error) {
	lvModel, err := c.dataConverter.ToModel(lvEntity)
	if err != nil {
		return 0, err
	}
	if _, err := c.Engine.Insert(lvModel); err != nil {
		return 0, err
	}
	return 0, nil
}

func (c *LvRepositoryImpl) UpdateUsed(lvEntity *LogicalVolumeEntity) (int64, error) {
	lvModelInf, err := c.dataConverter.ToModel(lvEntity)
	if err != nil {
		return 0, err
	}
	lvModel := lvModelInf.(*LogicalVolume)
	if _, err := c.Engine.Alias("a").
		Where("a.volume_id=?", lvModel.VolumeId).Cols("used_by_type", "used_by_name", "related_pvc").
		Update(lvModel); err != nil {
		return 0, err
	}
	return 0, nil
}

func (c *LvRepositoryImpl) UpdatePr(lvEntity *LogicalVolumeEntity) (int64, error) {
	lvModelInf, err := c.dataConverter.ToModel(lvEntity)
	if err != nil {
		return 0, err
	}
	lvModel := lvModelInf.(*LogicalVolume)
	if _, err := c.Engine.Alias("a").
		Where("a.volume_id=?", lvModel.VolumeId).Cols("pr_node_id").
		Update(lvModel); err != nil {
		return 0, err
	}
	return 0, nil
}

func (c *LvRepositoryImpl) Save(lvEntity *LogicalVolumeEntity) (int64, error) {
	lvModelInf, err := c.dataConverter.ToModel(lvEntity)
	if err != nil {
		return 0, err
	}
	lvModel := lvModelInf.(*LogicalVolume)
	if _, err := c.Engine.Alias("a").
		Where("a.volume_id=?", lvModel.VolumeId).
		Update(lvModel); err != nil {
		return 0, err
	}
	return 0, nil
}

func (c *LvRepositoryImpl) FindByName(name string) (*LogicalVolumeEntity, error) {
	lvModel := &LogicalVolume{
		VolumeName: name,
	}
	exist, err := c.Engine.Alias("a").Where("a.volume_name=?", name).Get(lvModel)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, nil
	}

	lvEntityInf, err := c.dataConverter.ToEntity(lvModel)
	if err != nil {
		return nil, err
	}
	return lvEntityInf.(*LogicalVolumeEntity), nil
}

func (c *LvRepositoryImpl) FindByVolumeId(volumeId string) (*LogicalVolumeEntity, error) {
	lvModel := &LogicalVolume{}
	exist, err := c.Engine.Alias("a").Where("a.volume_id=?", volumeId).Get(lvModel)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, nil
	}

	lvEntityInf, err := c.dataConverter.ToEntity(lvModel)
	if err != nil {
		return nil, err
	}
	return lvEntityInf.(*LogicalVolumeEntity), nil
}

func (c *LvRepositoryImpl) QueryAll() ([]*LogicalVolumeEntity, error) {
	var lvs = make([]*LogicalVolume, 0)
	if err := c.Engine.Find(&lvs); err != nil {
		return nil, err
	}
	var lvInfs []interface{}
	for _, lv := range lvs {
		lvInfs = append(lvInfs, lv)
	}
	results, err := c.dataConverter.ToEntities(lvInfs)
	if err != nil {
		return nil, err
	}
	var resultInfs []*LogicalVolumeEntity
	for _, ret := range results {
		resultInfs = append(resultInfs, ret.(*LogicalVolumeEntity))
	}
	return resultInfs, nil
}

func (c *LvRepositoryImpl) QueryAllByType(lvType common.LvType) ([]*LogicalVolumeEntity, error) {
	var lvs = make([]*LogicalVolume, 0)
	if err := c.Engine.Alias("a").Where("a.lv_type=?", lvType).OrderBy("id desc").Find(&lvs); err != nil {
		return nil, err
	}
	var lvInfs []interface{}
	for _, lv := range lvs {
		lvInfs = append(lvInfs, lv)
	}
	results, err := c.dataConverter.ToEntities(lvInfs)
	if err != nil {
		return nil, err
	}
	var resultInfs []*LogicalVolumeEntity
	for _, ret := range results {
		resultInfs = append(resultInfs, ret.(*LogicalVolumeEntity))
	}
	return resultInfs, nil
}

func (c *LvRepositoryImpl) QueryAllByTypes(lvTypes ...common.LvType) ([]*LogicalVolumeEntity, error) {
	var lvs = make([]*LogicalVolume, 0)
	var whereClauses []string
	var whereClauseValues []interface{}
	for _, lvType := range lvTypes {
		whereClauses = append(whereClauses, "a.lv_type=?")
		whereClauseValues = append(whereClauseValues, lvType)
	}
	if err := c.Engine.Alias("a").Where(strings.Join(whereClauses, "||"), whereClauseValues...).OrderBy("id desc").Find(&lvs); err != nil {
		return nil, err
	}
	var lvInfs []interface{}
	for _, lv := range lvs {
		lvInfs = append(lvInfs, lv)
	}
	results, err := c.dataConverter.ToEntities(lvInfs)
	if err != nil {
		return nil, err
	}
	var resultInfs []*LogicalVolumeEntity
	for _, ret := range results {
		resultInfs = append(resultInfs, ret.(*LogicalVolumeEntity))
	}
	return resultInfs, nil
}

func (c *LvRepositoryImpl) DeleteByVolumeId(volumeId string) (int64, error) {
	if _, err := c.Engine.Unscoped().Delete(&LogicalVolume{VolumeId: volumeId}); err != nil {
		return 0, err
	}
	return 0, nil
}

func (c *LvRepositoryImpl) Delete(lvEntity *LogicalVolumeEntity) (int64, error) {
	return c.DeleteByVolumeId(lvEntity.VolumeId)
}

func (c *LvRepositoryImpl) FindByVolumeIds(volumeIds []string) ([]*LogicalVolumeEntity, error) {
	var lvs []*LogicalVolume
	err := c.Engine.In("volume_id", volumeIds).Find(&lvs)
	if err != nil {
		smslog.Infof("Error when FindByWwids %v", volumeIds)
		return nil, err
	}
	var pvEntities []*LogicalVolumeEntity
	for _, pv := range lvs {
		pvEntityInf, err := c.dataConverter.ToEntity(pv)
		if err != nil {
			smslog.Infof("convert pvModel %v to pvEntity err %s", pv, err.Error())
			continue
		}
		pvEntities = append(pvEntities, pvEntityInf.(*LogicalVolumeEntity))
	}
	return pvEntities, nil
}

func GetLvRepository() LvRepository {
	_lvRepoOnce.Do(func() {
		if _lvRepo == nil {
			impl := &LvRepositoryImpl{
				dataConverter: NewLvConverter(),
				BaseDB:        repository.GetBaseDB(),
			}
			_lvRepo = impl
		}
	})

	return _lvRepo
}

var (
	_lvRepo     LvRepository
	_lvRepoOnce sync.Once
)

type LvRepository interface {
	Create(lv *LogicalVolumeEntity) (int64, error)
	Save(clusterLun *LogicalVolumeEntity) (int64, error)
	UpdateUsed(clusterLun *LogicalVolumeEntity) (int64, error)
	UpdatePr(clusterLun *LogicalVolumeEntity) (int64, error)
	FindByName(name string) (*LogicalVolumeEntity, error)
	FindByVolumeId(volumeId string) (*LogicalVolumeEntity, error)
	FindByVolumeIds(volumeIds []string) ([]*LogicalVolumeEntity, error)
	QueryAll() ([]*LogicalVolumeEntity, error)
	QueryAllByType(lvType common.LvType) ([]*LogicalVolumeEntity, error)
	QueryAllByTypes(lvType ...common.LvType) ([]*LogicalVolumeEntity, error)
	DeleteByVolumeId(volumeId string) (int64, error)
	Delete(lvEntity *LogicalVolumeEntity) (int64, error)
}
