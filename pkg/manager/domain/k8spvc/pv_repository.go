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
	smslog "polardb-sms/pkg/log"
	"polardb-sms/pkg/manager/config"
	"polardb-sms/pkg/manager/domain"
	"strings"
	"sync"
)

var _pvRepo PvRepository
var pvOnce sync.Once

type PvRepositoryImpl struct {
	clientSet     kubernetes.Interface
	dataConverter domain.Converter
}

func (r *PvRepositoryImpl) CreateOrUpdate(pvEntity *PersistVolumeEntity) (int64, error) {
	pvInf, err := r.dataConverter.ToModel(pvEntity)
	if err != nil {
		return 0, err
	}
	//todo persist in proper loc
	pv, ok := pvInf.(*corev1.PersistentVolume)
	if !ok {
		smslog.Infof("convert pvInf to pv err, pv %s", pvEntity.Name)
		return 0, fmt.Errorf("convert pvInf to pv err, pv %s", pvEntity.Name)
	}
	existPv, err := r.clientSet.CoreV1().PersistentVolumes().Get(pv.Name, metav1.GetOptions{})
	if err != nil {
		if strings.Contains(err.Error(), NotFoundErrStr) {
			_, err = r.clientSet.CoreV1().PersistentVolumes().Create(pv)
			return 0, err
		}
		return 0, err
	}
	if existPv != nil {
		if existPv.Status.Phase == corev1.VolumeBound {
			if existPv.Spec.ClaimRef.Name == pvEntity.PvcName &&
				existPv.Spec.ClaimRef.Namespace == pvEntity.PvcNamespace {
				return 0, nil
			}
			return 0, fmt.Errorf("pv is already bound to another pvc")
		}
		existPv.Spec.Capacity = pv.Spec.Capacity
		existPv.Labels = pv.Labels
		existPv.Spec.ClaimRef = &corev1.ObjectReference{
			APIVersion: pvcDefaultAPIVersion,
			Kind:       pvcDefaultKind,
			Name:       pvEntity.PvcName,
			Namespace:  pvEntity.PvcNamespace,
		}
		_, err = r.clientSet.CoreV1().PersistentVolumes().Update(existPv)
		return 0, err
	}
	return 0, err
}

func (r *PvRepositoryImpl) Delete(pvEntity *PersistVolumeEntity) (int64, error) {
	if err := DeletePv(r.clientSet, pvEntity.Name); err != nil {
		smslog.Infof("delete pv %s err %s", pvEntity.Name, err.Error())
		return 0, err
	}
	return 0, nil
}

func GetPvRepository() PvRepository {
	pvOnce.Do(func() {
		_pvRepo = &PvRepositoryImpl{
			clientSet:     config.ClientSet,
			dataConverter: NewPvConverter(),
		}
	})
	return _pvRepo
}
