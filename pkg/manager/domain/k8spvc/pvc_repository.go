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


package k8spvc

import (
	"fmt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"polardb-sms/pkg/common"
	smslog "polardb-sms/pkg/log"
	"polardb-sms/pkg/manager/config"
	"polardb-sms/pkg/manager/domain"
	"polardb-sms/pkg/manager/domain/repository"
	"strings"
	"sync"
)

var _pvcRepo PvcRepository
var pvcOnce sync.Once

var _pvcStatusRepo PvcStatusRepository
var pvcStatusOnce sync.Once

var NotFoundErrStr = "not found"

//this is special for mix the dao and entity repository
type PvcRepositoryImpl struct {
	clientSet     kubernetes.Interface
	dataConverter domain.Converter
	*repository.BaseDB
}

func (r *PvcRepositoryImpl) FindByPvcName(pvcName, namespace string) (*PersistVolumeClaimEntity, error) {
	pvc, err := FindPvcModel(r.clientSet, pvcName, namespace)
	if err != nil {
		if strings.Contains(err.Error(), NotFoundErrStr) {
			smslog.Errorf("can not find pvc %s in k8s", pvcName)
		}
		smslog.Errorf("find pvc %s in k8s err %s", pvcName, err.Error())
	}
	if pvc != nil {
		e, err := r.dataConverter.ToEntity(pvc)
		if err != nil {
			return nil, err
		}
		ret, ok := e.(*PersistVolumeClaimEntity)
		if !ok {
			return nil, e.(error)
		}
		ret.PvcRepo = r
		return ret, nil
	}
	return nil, fmt.Errorf("can not find pvc %s from db or k8s", pvcName)
}

func (r *PvcRepositoryImpl) FindByVolumeId(volumeId, pvcName string) (*PersistVolumeClaimEntity, error) {
	pvc, err := FindPvcModelByVolumeId(r.clientSet, volumeId, pvcName)
	if err != nil {
		if strings.Contains(err.Error(), NotFoundErrStr) {
			smslog.Errorf("can not find pvc %s in k8s", pvcName)
		}
		smslog.Errorf("find volume %s bound pvc %s in k8s err %s", volumeId, pvcName, err.Error())
	}
	if pvc != nil {
		e, err := r.dataConverter.ToEntity(pvc)
		if err != nil {
			return nil, err
		}
		ret, ok := e.(*PersistVolumeClaimEntity)
		if !ok {
			return nil, e.(error)
		}
		ret.PvcRepo = r
		return ret, nil
	}
	return nil, fmt.Errorf("can not find pvc %s from db or k8s", pvcName)
}

func (r *PvcRepositoryImpl) FindByPvcNameFromDb(pvcName, namespace string) (*PersistVolumeClaimEntity, error) {
	var pvc = &Pvc{}
	exist, err := r.Engine.Alias("a").Where("a.pvc_name=? and a.pvc_namespace=?", pvcName, namespace).Get(pvc)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, nil
	}

	return &PersistVolumeClaimEntity{
		Name:      pvc.PvcName,
		Namespace: pvc.PvcNamespace,
		PvcStatus: domain.ParseVolumeStatus(pvc.PvcStatus),
	}, nil
}

func (r *PvcRepositoryImpl) Delete(pvcEntity *PersistVolumeClaimEntity) (int64, error) {
	pvc, err := FindPvcModel(r.clientSet, pvcEntity.Name, pvcEntity.Namespace)
	if err != nil {
		smslog.Errorf("FindPvcModel %s err %s ", pvcEntity.Name, err.Error())
		return 0, err
	}
	if err := RemovePvcFinalizer(r.clientSet, pvc); err != nil {
		smslog.Errorf("remove pvc %s finalizer err %s", pvc.Name, err.Error())
		return 0, err
	}
	_ = DeletePvcModel(r.clientSet, pvc.Name, pvc.Namespace)
	return 0, nil
}

func (r *PvcRepositoryImpl) Create(pvcEntity *PersistVolumeClaimEntity) (int64, error) {
	pvcModel, err := r.dataConverter.ToModel(pvcEntity)
	if err != nil {
		return 0, err
	}
	pvc, ok := pvcModel.(*corev1.PersistentVolumeClaim)
	if !ok {
		return 0, fmt.Errorf("can not conver to pvcEntity %v to pvc", pvcEntity)
	}
	_, err = r.clientSet.CoreV1().PersistentVolumeClaims(pvcEntity.Namespace).Create(pvc)
	return 0, err
}

//todo update
func (r *PvcRepositoryImpl) UpdateWorkflow(pvcEntity *PersistVolumeClaimEntity) (int64, error) {
	pvcModel, err := FindPvcModel(r.clientSet, pvcEntity.Name, pvcEntity.Namespace)
	if err != nil {
		return 0, err
	}
	var wfl = pvcEntity.RelatedWorkflow
	if pvcEntity.RelatedWorkflow == "" {
		wfl = domain.DummyWorkflowId
	}
	err = UpdatePvcWorkflowLabel(r.clientSet, pvcModel, wfl)
	if err != nil {
		smslog.Errorf("Update err: %s", err.Error())
		return 0, err
	}
	return 0, nil
}

func (r *PvcRepositoryImpl) UpdateUsedDbCluster(pvcEntity *PersistVolumeClaimEntity) (int64, error) {
	pvcModel, err := FindPvcModel(r.clientSet, pvcEntity.Name, pvcEntity.Namespace)
	if err != nil {
		return 0, err
	}
	if err := UpdateUsedDbCLusterLabel(r.clientSet, pvcModel, pvcEntity.DbClusterName); err != nil {
		smslog.Errorf("Update err: %s", err.Error())
		return 0, err
	}
	return 0, nil
}

func (r *PvcRepositoryImpl) UpdatePrKey(pvcEntity *PersistVolumeClaimEntity) (int64, error) {
	pvcModel, err := FindPvcModel(r.clientSet, pvcEntity.Name, pvcEntity.Namespace)
	if err != nil {
		return 0, err
	}
	if err := UpdatePrKeyLabel(r.clientSet, pvcModel, pvcEntity.GetPrKey()); err != nil {
		smslog.Errorf("Update err: %s", err.Error())
		return 0, err
	}
	return 0, nil
}

func (r *PvcRepositoryImpl) FindByLockedWorkflow(wflId string) (*PersistVolumeClaimEntity, error) {
	pvcList, err := r.clientSet.CoreV1().PersistentVolumeClaims("").List(metav1.ListOptions{
		LabelSelector: GetPvcLockLabelSelector(wflId),
	})
	if err != nil {
		smslog.Errorf("list Pvc err: %s selector:%s", err.Error(), GetPvcLockLabelSelector(wflId))
		return nil, err
	}
	if len(pvcList.Items) != 1 {
		err := fmt.Errorf("list pvc get wrong items: %v", pvcList.Items)
		smslog.Error(err.Error())
		return nil, err
	}
	e, err := r.dataConverter.ToEntity(&(pvcList.Items[0]))
	if err != nil {
		return nil, err
	}
	ret, ok := e.(*PersistVolumeClaimEntity)
	if !ok {
		return nil, e.(error)
	}
	ret.PvcRepo = r
	return ret, nil
}

func (r *PvcRepositoryImpl) QueryAll() ([]*PersistVolumeClaimEntity, error) {
	pvcList, err := GetAllPvcs(r.clientSet)
	if err != nil {
		smslog.Errorf("PvcRepository QueryAll error %s", err.Error())
		return nil, err
	}
	var pvcEntityList []*PersistVolumeClaimEntity
	for _, pvcItem := range pvcList {
		pvcEntityInf, err := r.dataConverter.ToEntity(pvcItem)
		if err != nil {
			smslog.Warnf("convert to pvcEntity err %s", err.Error())
			continue
		}
		pvcEntityInf.(*PersistVolumeClaimEntity).PvcRepo = r
		pvcEntityList = append(pvcEntityList, pvcEntityInf.(*PersistVolumeClaimEntity))
	}
	return pvcEntityList, nil
}

func FindExistedPvc(src []*PersistVolumeClaimEntity, item *PersistVolumeClaimEntity) *PersistVolumeClaimEntity {
	for _, existedItem := range src {
		if existedItem.Name == item.Name &&
			existedItem.Namespace == item.Namespace {
			return existedItem
		}
	}
	return nil
}

func mergePvcs(src []*PersistVolumeClaimEntity, items []*PersistVolumeClaimEntity) []*PersistVolumeClaimEntity {
	var itemChanged *PersistVolumeClaimEntity
	var addedPvcs = make([]*PersistVolumeClaimEntity, 0)
	for _, item := range items {
		itemChanged = FindExistedPvc(src, item)
		if itemChanged != nil {
			itemChanged.PvcStatus = item.PvcStatus
		} else {
			addedPvcs = append(addedPvcs, item)
		}
	}
	return append(src, addedPvcs...)
}

func (r *PvcRepositoryImpl) QueryAllFromDB() ([]*PersistVolumeClaimEntity, error) {
	var pvcs = make([]*Pvc, 0)
	if err := r.Engine.Find(&pvcs); err != nil {
		return nil, err
	}
	var resultInfs []*PersistVolumeClaimEntity
	for _, pvc := range pvcs {
		resultInfs = append(resultInfs, &PersistVolumeClaimEntity{
			Name:      pvc.PvcName,
			Namespace: pvc.PvcNamespace,
			PvcStatus: domain.ParseVolumeStatus(pvc.PvcStatus),
		})
	}
	return resultInfs, nil
}

func (r *PvcRepositoryImpl) QueryByVolumeClass(volumeType string) ([]*PersistVolumeClaimEntity, error) {
	pvcList, err := GetAllPvcsByVolumeType(r.clientSet, volumeType)
	if err != nil {
		smslog.Errorf("list Pvc err: %s", err.Error())
		return nil, err
	}
	var pvcEntityList []*PersistVolumeClaimEntity
	for _, pvcItem := range pvcList {
		pvcEntityInf, err := r.dataConverter.ToEntity(pvcItem)
		if err != nil {
			smslog.Warnf("convert to pvcEntity err %s", err.Error())
			continue
		}
		pvcEntityInf.(*PersistVolumeClaimEntity).PvcRepo = r
		pvcEntityList = append(pvcEntityList, pvcEntityInf.(*PersistVolumeClaimEntity))
	}
	pvcListFromDb, err := r.QueryFromDbByVolumeClass(volumeType)
	if err != nil {
		smslog.Errorf("QueryFromDbByVolumeClass err %s", err.Error())
	} else {
		pvcEntityList = mergePvcs(pvcEntityList, pvcListFromDb)
	}
	return pvcEntityList, nil
}

func (r *PvcRepositoryImpl) QueryFromDbByVolumeClass(volumeClass string) ([]*PersistVolumeClaimEntity, error) {
	var pvcs = make([]*Pvc, 0)
	if err := r.Engine.Alias("a").Where("a.volume_class=?", volumeClass).Find(&pvcs); err != nil {
		return nil, err
	}
	var resultInfs []*PersistVolumeClaimEntity
	for _, pvc := range pvcs {
		resultInfs = append(resultInfs, &PersistVolumeClaimEntity{
			Name:      pvc.PvcName,
			Namespace: pvc.PvcNamespace,
			PvcStatus: domain.ParseVolumeStatus(pvc.PvcStatus),
		})
	}
	return resultInfs, nil
}

func (r *PvcRepositoryImpl) UpdateCapacity(pvcEntity *PersistVolumeClaimEntity) (int64, error) {
	err := UpdatePvcRequestCapacity(r.clientSet,
		pvcEntity.Name,
		pvcEntity.Namespace,
		common.BytesToGiBString(pvcEntity.ExpectedDiskStatus.Size))
	if err != nil {
		err = fmt.Errorf("update pvc capacity err %s", err.Error())
		smslog.Error(err.Error())
		return 0, err
	}
	err = UpdatePvCapacity(r.clientSet,
		pvcEntity.PvName,
		common.BytesToGiBString(pvcEntity.ExpectedDiskStatus.Size))
	if err != nil {
		err = fmt.Errorf("update pv capacity err %s", err.Error())
		smslog.Error(err.Error())
		return 0, err
	}
	err = UpdatePvcCapacity(r.clientSet,
		pvcEntity.Name,
		pvcEntity.Namespace,
		common.BytesToGiBString(pvcEntity.ExpectedDiskStatus.Size))
	if err != nil {
		err = fmt.Errorf("update pvc capacity err %s", err.Error())
		smslog.Error(err.Error())
		return 0, err
	}
	return 0, nil
}

func (r *PvcRepositoryImpl) UpdateStatus(pvcEntity *PersistVolumeClaimEntity) (int64, error) {
	pvcModel, err := FindPvcModel(r.clientSet, pvcEntity.Name, pvcEntity.Namespace)
	if err != nil {
		return 0, err
	}
	err = UpdatePvcStatusLabel(r.clientSet, pvcModel, int(pvcEntity.PvcStatus.StatusValue))
	if err != nil {
		smslog.Errorf("Update err: %s", err.Error())
		return 0, err
	}
	return 0, nil
}

func GetPvcRepository() PvcRepository {
	pvcOnce.Do(func() {
		_pvcRepo = &PvcRepositoryImpl{
			clientSet:     config.ClientSet,
			dataConverter: NewPvcConverter(),
			BaseDB:        repository.GetBaseDB(),
		}
	})
	return _pvcRepo
}

type PvRepository interface {
	CreateOrUpdate(pvEntity *PersistVolumeEntity) (int64, error)
	Delete(pvEntity *PersistVolumeEntity) (int64, error)
}

type PvcRepository interface {
	FindByPvcName(pvcName, namespace string) (*PersistVolumeClaimEntity, error)
	FindByLockedWorkflow(wflId string) (*PersistVolumeClaimEntity, error)
	FindByVolumeId(volumeId, pvcName string) (*PersistVolumeClaimEntity, error)
	Delete(pvcEntity *PersistVolumeClaimEntity) (int64, error)
	Create(pvcEntity *PersistVolumeClaimEntity) (int64, error)
	UpdateWorkflow(pvcEntity *PersistVolumeClaimEntity) (int64, error)
	UpdateUsedDbCluster(pvcEntity *PersistVolumeClaimEntity) (int64, error)
	UpdatePrKey(pvcEntity *PersistVolumeClaimEntity) (int64, error)
	QueryAll() ([]*PersistVolumeClaimEntity, error)
	QueryByVolumeClass(volumeType string) ([]*PersistVolumeClaimEntity, error)
	UpdateCapacity(pvcEntity *PersistVolumeClaimEntity) (int64, error)
	UpdateStatus(pvcEntity *PersistVolumeClaimEntity) (int64, error)
}

type PvcStatusRepository interface {
	Create(pvcEntity *PersistVolumeClaimEntity) (int64, error)
}

type PvcStatusRepositoryImpl struct {
	*repository.BaseDB
}

func (p *PvcStatusRepositoryImpl) Create(pvcEntity *PersistVolumeClaimEntity) (int64, error) {
	status := domain.VolumeStatus{
		StatusValue: domain.Creating,
	}
	pvcModel := &Pvc{
		PvcName:      pvcEntity.Name,
		PvcNamespace: pvcEntity.Namespace,
		PvcStatus:    status.String(),
	}
	return p.Engine.Insert(pvcModel)
}

func GetPvcStatusRepository() PvcStatusRepository {
	pvcStatusOnce.Do(func() {
		_pvcStatusRepo = &PvcStatusRepositoryImpl{
			repository.GetBaseDB(),
		}
	})
	return _pvcStatusRepo
}
