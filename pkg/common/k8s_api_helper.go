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

package common

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"time"
)

func UpdatePvcFinalizers(name, namespace string, finalizers []string, clientSet kubernetes.Interface) error {
	return RunWithRetry(3, 1*time.Second, func(retryTimes int) error {
		pvc, err := clientSet.CoreV1().PersistentVolumeClaims(namespace).Get(name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		pvc.ObjectMeta.Finalizers = finalizers
		_, err = clientSet.CoreV1().PersistentVolumeClaims(pvc.Namespace).Update(pvc)
		return err
	})
}

func UpdatePvcConditions(name, namespace string, conditions []corev1.PersistentVolumeClaimCondition, clientSet kubernetes.Interface) error {
	return RunWithRetry(3, 1*time.Second, func(retryTimes int) error {
		pvc, err := clientSet.CoreV1().PersistentVolumeClaims(namespace).Get(name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		pvc.Status.Conditions = conditions
		_, err = clientSet.CoreV1().PersistentVolumeClaims(pvc.Namespace).UpdateStatus(pvc)
		return err
	})
}

func GetNodes(clientSet kubernetes.Interface) ([]corev1.Node, error) {
	var nodes []corev1.Node
	err := RunWithRetry(3, 1*time.Second, func(retryTimes int) error {
		nodeList, err := clientSet.CoreV1().Nodes().List(metav1.ListOptions{})
		if err != nil {
			return err
		}
		nodes = nodeList.Items
		return nil
	})
	if err != nil {
		return nil, err
	}
	return nodes, nil
}
