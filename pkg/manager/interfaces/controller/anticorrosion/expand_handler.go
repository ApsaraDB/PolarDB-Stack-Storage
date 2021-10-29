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
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/client-go/kubernetes"
	"polardb-sms/pkg/common"
	smslog "polardb-sms/pkg/log"
	"polardb-sms/pkg/manager/application/service"
	"polardb-sms/pkg/manager/application/view"
	"sync"
)

type FsExpandHandler struct {
	//value is final size of growfs result
	requestCache map[string]resource.Quantity
	growFsLock   sync.RWMutex
	clientSet    kubernetes.Interface
	lunService   *service.LvForOldLunService
	wflService   *service.WorkflowService
	pvcService   *service.PvcService
}

func (h *FsExpandHandler) handlePvcEvent(ctx common.TraceContext, pvcEvent *PvcEvent) (pvcResponse *PvcResponse, err error) {
	var (
		pvc      = pvcEvent.Pvc
		response = &PvcResponse{
			pvc:           pvc,
			reason:        "Failed",
			status:        corev1.ConditionFalse,
			message:       "",
			conditionType: PVCGrowFs,
		}
	)
	wwid, err := getWwid(pvc)
	if err != nil {
		response.message = err.Error()
		return response, err
	}

	pvName := pvc.Spec.VolumeName
	if pvName == "" {
		err = fmt.Errorf("could not find pv name by pvc.Spec.VolumeId")
		response.message = err.Error()
		return response, err
	}

	specCapacity, ok := pvc.Spec.Resources.Requests[corev1.ResourceStorage]
	if !ok {
		err = fmt.Errorf("could not get real pv size by pvc.Spec.Resources.Requests[storage]")
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
	//current only lun do this
	pvcFsExpandReq := &view.PvcExpandFsRequest{
		PvcRequest: view.PvcRequest{
			Name:      pvc.Name,
			Namespace: pvc.Namespace,
		},
		LvType:   common.MultipathVolume,
		VolumeId: wwid,
		FsType:   fsType,
		ReqSize:  specCapacity.Value(),
	}
	smslog.WithContext(ctx).Infof("expand fs %v", pvcFsExpandReq)
	wflResp, err := h.pvcService.PvcExpandFs(ctx, pvcFsExpandReq)
	if err != nil {
		response.message = err.Error()
		return response, err
	}
	err = h.wflService.WaitUntilWorkflowFinish(wflResp.WorkflowId)
	if err != nil {
		response.message = err.Error()
		return response, err
	}

	response.reason = "Success"
	response.status = corev1.ConditionFalse
	return response, nil
}
