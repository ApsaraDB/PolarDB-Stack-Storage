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
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/client-go/kubernetes"
	"polardb-sms/pkg/common"
	smslog "polardb-sms/pkg/log"
	"polardb-sms/pkg/manager/domain"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	pvcDefaultAPIVersion = "v1"
	pvcDefaultKind       = "PersistentVolumeClaim"
	DriverName           = "csi-polardb-fc-plugin"
)

var _ domain.Converter = &PvConverter{}

type PvConverter struct {
}

func (c *PvConverter) ToModel(t interface{}) (interface{}, error) {
	pvEntity := t.(*PersistVolumeEntity)
	pvcLabel := pvEntity.PvcName
	quantity, err := pvEntity.GetQuantity()
	if err != nil {
		return nil, err
	}
	pv := &corev1.PersistentVolume{
		TypeMeta: metav1.TypeMeta{
			APIVersion: pvcDefaultAPIVersion,
			Kind:       pvcDefaultKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: pvEntity.Name,
			Labels: map[string]string{
				"pvc": pvcLabel,
			},
		},
		Spec: corev1.PersistentVolumeSpec{
			StorageClassName:              pvEntity.StorageClassName,
			PersistentVolumeReclaimPolicy: corev1.PersistentVolumeReclaimRetain,
			AccessModes:                   []corev1.PersistentVolumeAccessMode{corev1.ReadWriteMany},
			Capacity: corev1.ResourceList{
				corev1.ResourceStorage: quantity,
			},
			ClaimRef: &corev1.ObjectReference{
				APIVersion: pvcDefaultAPIVersion,
				Kind:       pvcDefaultKind,
				Name:       pvEntity.PvcName,
				Namespace:  pvEntity.PvcNamespace,
			},
		},
	}
	smslog.Infof("pv storageclass name %s", pvEntity.StorageClassName)
	if pvEntity.StorageClassName != "local-storage" {
		pv.Spec.PersistentVolumeSource = corev1.PersistentVolumeSource{
			CSI: &corev1.CSIPersistentVolumeSource{
				ReadOnly: false,
				Driver:   DriverName,
				VolumeAttributes: map[string]string{
					"wwid": pvEntity.Request.VolumeId,
				},
				VolumeHandle: pvEntity.Name,
			},
		}
	} else {
		pv.Spec.Local = &corev1.LocalVolumeSource{
			Path: "/dev/mapper/" + pvEntity.Request.VolumeId,
		}
		pv.Spec.NodeAffinity = &corev1.VolumeNodeAffinity{
			Required: &corev1.NodeSelector{
				NodeSelectorTerms: []corev1.NodeSelectorTerm{
					{
						MatchExpressions: []corev1.NodeSelectorRequirement{
							{
								Key:      "nodetype",
								Operator: "In",
								Values: []string{
									"agent",
								},
							},
						},
					},
				},
			},
		}
	}

	var (
		volumeBlockMode      = corev1.PersistentVolumeBlock
		volumeFilesystemMode = corev1.PersistentVolumeFilesystem
	)
	switch pvEntity.Request.VolumeMode {
	case Block:
		pv.Spec.VolumeMode = &volumeBlockMode
	case FsExt4:
		pv.Spec.VolumeMode = &volumeFilesystemMode
		pv.Spec.PersistentVolumeSource.CSI.FSType = "ext4"
	}
	return pv, nil
}

func (c *PvConverter) ToEntity(t interface{}) (interface{}, error) {
	m := t.(*corev1.PersistentVolume)

	var mode RequestVolumeMode
	switch *m.Spec.VolumeMode {
	case corev1.PersistentVolumeBlock:
		mode = Block
	case corev1.PersistentVolumeFilesystem:
		mode = FsExt4
	default:
		mode = Block
	}

	specCapacity := m.Spec.Capacity[corev1.ResourceStorage]

	e := &PersistVolumeEntity{
		Name:             m.Name,
		PvcName:          m.Spec.ClaimRef.Name,
		PvcNamespace:     m.Spec.ClaimRef.Namespace,
		StorageClassName: m.Spec.StorageClassName,
		Request: VolumeMeta{
			Size:       specCapacity.Value(),
			VolumeId:   m.Spec.CSI.VolumeAttributes["wwid"],
			VolumeMode: mode,
		},
	}
	return e, nil
}

func (c *PvConverter) ToEntities(ds []interface{}) ([]interface{}, error) {
	var es []interface{}
	for _, m := range ds {
		e, _ := c.ToEntity(m.(*corev1.PersistentVolume))
		es = append(es, e.(*PersistVolumeClaimEntity))
	}
	return es, nil
}

func DeletePv(clientSet kubernetes.Interface, pvName string) error {
	return common.RunWithRetry(3, 1*time.Second, func(retryTimes int) error {
		err := clientSet.CoreV1().PersistentVolumes().Delete(pvName, &metav1.DeleteOptions{})
		if err != nil && !strings.Contains(err.Error(), "not found") {
			return nil
		}
		return err
	})
}

func FindPvModel(clientSet kubernetes.Interface, pvName string) (*corev1.PersistentVolume, error) {
	var (
		ret *corev1.PersistentVolume
		err error
	)
	err = common.RunWithRetry(3, 1*time.Second, func(retryTimes int) error {
		ret, err = clientSet.CoreV1().PersistentVolumes().Get(pvName, metav1.GetOptions{})
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func UpdatePvCapacity(clientSet kubernetes.Interface, pvName string, size string) error {
	return common.RunWithRetry(3, 1*time.Second, func(retryTimes int) error {
		pv, err := FindPvModel(clientSet, pvName)
		if err != nil {
			return err
		}
		newQuantity, err := resource.ParseQuantity(size)
		if err != nil {
			return err
		}
		if statusStorage, exist := pv.Spec.Capacity[corev1.ResourceStorage]; exist && statusStorage.Equal(newQuantity) {
			return nil
		}
		pv.Spec.Capacity[corev1.ResourceStorage] = newQuantity
		_, err = clientSet.CoreV1().PersistentVolumes().Update(pv)
		return err
	})
}

func NewPvConverter() domain.Converter {
	return &PvConverter{}
}
