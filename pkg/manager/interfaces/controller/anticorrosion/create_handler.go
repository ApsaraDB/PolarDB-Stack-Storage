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
	"k8s.io/client-go/kubernetes"
	"polardb-sms/pkg/common"
	smslog "polardb-sms/pkg/log"
	"polardb-sms/pkg/manager/application/service"
	"polardb-sms/pkg/manager/application/view"
)

type CreateHandler struct {
	nodeId     string
	nodeIp     string
	clientSet  kubernetes.Interface
	pvcService *service.PvcService
	wflService *service.WorkflowService
}

func (h *CreateHandler) handlePvcEvent(ctx common.TraceContext, pvcEvent *PvcEvent) (pvcResponse *PvcResponse, err error) {
	var (
		pvc      = pvcEvent.Pvc
		pvcName  = pvc.Name
		response = &PvcResponse{
			pvc:           pvc,
			reason:        "Failed",
			status:        corev1.ConditionFalse,
			message:       "",
			conditionType: PVCPreProvisionedVolume,
		}
	)

	if err = h.addFinalizer(ctx, pvc, PreProvisionedVolumeFinalizer); err != nil {
		err = fmt.Errorf("failed added finalizers for pvc %s: %s", pvcName, err)
		response.message = err.Error()
		return response, err
	}

	wwid, err := getWwid(pvc)
	if err != nil {
		response.message = err.Error()
		return response, err
	}
	smslog.WithContext(ctx).Debugf("start to create pvc for %s", wwid)
	workflow, err := h.pvcService.BindPvcAndVolumeLegacy(ctx, &view.PvcBindVolumeRequest{
		PvcRequest: view.PvcRequest{
			Name:      pvcName,
			Namespace: pvc.Namespace,
		},
		LvType:       common.MultipathVolume,
		VolumeId:     wwid,
		StorageClass: *pvc.Spec.StorageClassName,
	})
	if err != nil {
		response.message = err.Error()
		return response, err
	}
	if err = h.wflService.WaitUntilWorkflowFinish(workflow.WorkflowId); err != nil {
		response.message = err.Error()
		return response, err
	}
	smslog.WithContext(ctx).Infof("workflow exec successful %s", workflow.WorkflowId)
	response.reason = "Success"
	response.status = corev1.ConditionFalse
	return response, nil
}

func (h *CreateHandler) addFinalizer(ctx common.TraceContext, pvc *corev1.PersistentVolumeClaim, finalizers ...string) error {
	inUseFinalizers := make(map[string]bool)
	for _, item := range pvc.ObjectMeta.Finalizers {
		inUseFinalizers[item] = true
	}

	for _, finalizer := range finalizers {
		// finalizers register key, if key not exist
		inUseFinalizers[finalizer] = true
	}

	var result []string
	for finalizer := range inUseFinalizers {
		result = append(result, finalizer)
	}

	if err := common.UpdatePvcFinalizers(pvc.GetName(), pvc.GetNamespace(), result, h.clientSet); err != nil {
		smslog.WithContext(ctx).Errorf("Could not added finalizer [%+v] for %s/%s: %v", finalizers, pvc.Namespace, pvc.Name, err)
		return err
	}
	smslog.WithContext(ctx).Infof("Successfully added finalizer [%+v] for %s/%s", finalizers, pvc.Namespace, pvc.Name)

	return nil
}
