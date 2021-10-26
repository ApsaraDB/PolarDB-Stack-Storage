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
)

type FsFormatHandler struct {
	clientSet  kubernetes.Interface
	nodeId     string
	nodeIp     string
	lunService *service.LvForOldLunService
	wflService *service.WorkflowService
}

type FormatResultMessage struct {
	ActionId string `json:"actionId"`
	ErrMsg   string `json:"errMsg"`
}

func (m *FormatResultMessage) String() string {
	msg, err := json.Marshal(m)
	if err != nil {
		return ""
	}
	return string(msg)
}

func (h *FsFormatHandler) handlePvcEvent(ctx common.TraceContext, pvcEvent *PvcEvent) (pvcResponse *PvcResponse, err error) {
	var (
		pvc      = pvcEvent.Pvc
		response = &PvcResponse{
			reason:        "Failed",
			message:       "",
			pvc:           pvc,
			status:        corev1.ConditionFalse,
			conditionType: PVCFormatFs,
		}
		formatResultMsg = FormatResultMessage{
			ActionId: pvcEvent.RequestID,
			ErrMsg:   "",
		}
	)
	wwid, err := getWwid(pvc)
	if err != nil {
		response.message = err.Error()
		return response, err
	}

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

	lunFsFormatView := &view.ClusterLunFormatRequest{
		Name:   wwid,
		Wwid:   wwid,
		FsType: fsType,
		FsSize: specCapacity.Value(),
	}

	wflResp, err := h.lunService.Format(ctx, lunFsFormatView)
	if err != nil {
		formatResultMsg.ErrMsg = err.Error()
		response.message = formatResultMsg.String()
		return response, err
	}
	err = h.wflService.WaitUntilWorkflowFinish(wflResp.WorkflowId)
	if err != nil {
		response.message = err.Error()
		return response, err
	}

	response.reason = "Success"
	response.message = formatResultMsg.String()
	response.status = corev1.ConditionFalse
	return response, nil
}
