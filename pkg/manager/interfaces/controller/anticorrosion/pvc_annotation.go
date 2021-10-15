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


package anticorrosion

import (
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"strings"
)

func getWwid(pvc *corev1.PersistentVolumeClaim) (string, error) {
	wwid, ok := pvc.Annotations[PreProvisionedVolumeWWID]
	if !ok {
		err := fmt.Errorf("failed get wwid form pvc annotation: %s", PreProvisionedVolumeWWID)
		return "", err
	}
	wwid = fmt.Sprintf("3%s", strings.ToLower(wwid))
	return wwid, nil
}

func getLockNodeId(pvc *corev1.PersistentVolumeClaim) (string, error) {
	nodeId, ok := pvc.Annotations[VolumeLockNodeId]
	if !ok || nodeId == "" {
		err := fmt.Errorf("failed get VolumeLockNodeId form pvc annotation: %s", PreProvisionedVolumeWWID)
		return "", err
	}
	return nodeId, nil
}