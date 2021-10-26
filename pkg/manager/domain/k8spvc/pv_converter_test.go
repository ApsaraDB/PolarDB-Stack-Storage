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
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

var c *PvConverter

func TestPvConverter_ToModel(t *testing.T) {
	e := &PersistVolumeEntity{
		Name:             "test",
		PvcName:          "polar-test-007",
		PvcNamespace:     "default",
		StorageClassName: "csi-polardb-fc",
		Request: VolumeMeta{
			Size:       "123",
			VolumeId:   "wwid",
			VolumeMode: Block,
		},
	}
	m, err := c.ToModel(e)
	assert.NotEmpty(t, m)
	assert.NoError(t, err)
}

func TestPvConverter_ToEntity(t *testing.T) {
	quantity, _ := resource.ParseQuantity("123")
	mode := corev1.PersistentVolumeBlock
	pv := &corev1.PersistentVolume{
		TypeMeta: metav1.TypeMeta{
			APIVersion: pvcDefaultAPIVersion,
			Kind:       pvcDefaultKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
			Labels: map[string]string{
				"pvc": "pvcLabel",
			},
		},
		Spec: corev1.PersistentVolumeSpec{
			StorageClassName:              "csi-polardb-fc",
			PersistentVolumeReclaimPolicy: corev1.PersistentVolumeReclaimRetain,
			AccessModes:                   []corev1.PersistentVolumeAccessMode{corev1.ReadWriteMany},
			Capacity: corev1.ResourceList{
				corev1.ResourceStorage: quantity,
			},
			VolumeMode: &mode,
			PersistentVolumeSource: corev1.PersistentVolumeSource{
				CSI: &corev1.CSIPersistentVolumeSource{
					ReadOnly: false,
					Driver:   DriverName,
					VolumeAttributes: map[string]string{
						"wwid": "wwid",
					},
					VolumeHandle: "pv-test",
				},
			},
			ClaimRef: &corev1.ObjectReference{
				APIVersion: pvcDefaultAPIVersion,
				Kind:       pvcDefaultKind,
				Name:       "polar-test-007",
				Namespace:  "default",
			},
		},
	}
	e, err := c.ToEntity(pv)
	assert.NotEmpty(t, e)
	assert.NoError(t, err)
}
