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
	"encoding/json"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"polardb-sms/pkg/common"
	"polardb-sms/pkg/manager/application/service"
	"polardb-sms/pkg/manager/application/view"
	"polardb-sms/pkg/manager/config"
)

type FormatAndLockHandler struct {
	clientSet  kubernetes.Interface
	nodeId     string
	nodeIp     string
	lunService *service.LvForOldLunService
	wflService *service.WorkflowService
}

type FormatLockResultMessage struct {
	ActionId string `json:"actionId"`
	ErrMsg   string `json:"errMsg"`
}

func (m *FormatLockResultMessage) String() string {
	msg, err := json.Marshal(m)
	if err != nil {
		return ""
	}
	return string(msg)
}

func (h *FormatAndLockHandler) handlePvcEvent(ctx common.TraceContext, pvcEvent *PvcEvent) (pvcResponse *PvcResponse, err error) {
	var (
		pvc      = pvcEvent.Pvc
		response = &PvcResponse{
			reason:        "Failed",
			message:       "",
			pvc:           pvc,
			status:        corev1.ConditionFalse,
			conditionType: PVCFormatAndLockVolume,
		}
		lockResultMsg = FormatLockResultMessage{
			ActionId: pvcEvent.RequestID,
			ErrMsg:   "",
		}
	)
	wwid, err := getWwid(pvc)
	if err != nil {
		response.message = err.Error()
		return response, err
	}

	nodeId, err := getLockNodeId(pvc)
	if err != nil {
		response.message = err.Error()
		return response, err
	}
	node := config.GetNodeById(nodeId)

	var fsType common.FsType
	switch *pvc.Spec.VolumeMode {
	case corev1.PersistentVolumeBlock:
		fsType = common.Pfs
	case corev1.PersistentVolumeFilesystem:
		fsType = common.Ext4
	}

	specCapacity, ok := pvc.Spec.Resources.Requests[corev1.ResourceStorage]
	if !ok {
		err = fmt.Errorf("could not get real pv size by pvc.Spec.Resources.Requests[storage]")
		response.message = err.Error()
		return response, err
	}

	lunLockView := &view.ClusterLunFormatAndLockRequest{
		Name:      wwid,
		Wwid:      wwid,
		FsType:    fsType,
		FsSize:    specCapacity.Value(),
		RwNodeIp:  node.Ip,
		RoNodesIp: nil,
	}

	wflResp, err := h.lunService.FormatAndLock(ctx, lunLockView)
	if err != nil {
		lockResultMsg.ErrMsg = err.Error()
		response.message = lockResultMsg.String()
		return response, err
	}

	err = h.wflService.WaitUntilWorkflowFinish(wflResp.WorkflowId)
	if err != nil {
		response.message = err.Error()
		return response, err
	}

	response.reason = "Success"
	response.message = lockResultMsg.String()
	response.status = corev1.ConditionFalse
	return response, nil
}
