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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"polardb-sms/pkg/common"
	smslog "polardb-sms/pkg/log"
	"polardb-sms/pkg/manager/application/service"
	"polardb-sms/pkg/manager/application/view"
	"strings"
	"time"
)

type PreProvisionedDeleteHandler struct {
	ClientSet  kubernetes.Interface
	pvcService *service.PvcService
	wflService *service.WorkflowService
}

func (h *PreProvisionedDeleteHandler) DeletePreProvisionedVolumeLoop(stopCh <-chan struct{}) {
	var event = &PvcEvent{
		Action: "delete",
	}
	defer smslog.LogPanic()

	for {
		pvcList, err := h.ClientSet.CoreV1().PersistentVolumeClaims("").List(metav1.ListOptions{})
		if err != nil {
			smslog.Errorf("Failed list pvc: %v", err)
			return
		}

		for _, pvc := range pvcList.Items {
			if _, ok := pvc.Annotations[PreProvisionedVolumeWWID]; !ok {
				continue
			}

			if pvc.ObjectMeta.DeletionTimestamp == nil {
				continue
			}

			if *pvc.Spec.StorageClassName != StorageClassName {
				continue
			}

			if len(pvc.Spec.VolumeName) == 0 {
				continue
			}

			event.Pvc = pvc.DeepCopy()
			event.RequestID = pvc.Annotations[ClusterActionID]
			ctx := common.NewTraceContext(event.Map())
			if _, err = h.handlePvcEvent(ctx, event); err != nil {
				smslog.Errorf("Could not delete pre-provision volume %q: %v", pvc.Name, err)
				continue
			}
			smslog.Infof("Successfully delete pre-provision volume %q", pvc.Name)
		}

		select {
		case <-stopCh:
			return
		case <-time.After(10 * time.Minute):
			break
		}
	}
}

func (h *PreProvisionedDeleteHandler) handlePvcEvent(ctx common.TraceContext, pvcEvent *PvcEvent) (*PvcResponse, error) {
	var (
		err      error
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

	var found = false
	for _, f := range pvc.Finalizers {
		if f == PreProvisionedVolumeFinalizer {
			found = true
			break
		}
	}

	if !found {
		smslog.Infof("Not found finalizer [%s] in pvc, may be not deleted, drop it", PreProvisionedVolumeFinalizer)
		return nil, nil
	}

	if pvc.ObjectMeta.DeletionTimestamp == nil {
		smslog.Infof("Not found deletion timestamp in pvc, may be not deleted, drop it")
		return nil, nil
	}

	wwid, ok := pvc.Annotations[PreProvisionedVolumeWWID]
	if !ok {
		err = fmt.Errorf("could not find wwid in pvc annotations: %s", PreProvisionedVolumeWWID)
		response.message = err.Error()
		return response, err
	}
	wwid = fmt.Sprintf("3%s", strings.ToLower(wwid))

	wflResp, err := h.pvcService.DeletePvc(ctx, &view.PvcRequest{
		Name:      pvc.Name,
		Namespace: pvc.Namespace,
	})

	if err != nil {
		smslog.Infof("call pvcService.DeletePvc err %s", err.Error())
		response.message = err.Error()
		response.reason = "Failed"
		return response, nil
	}

	if err := h.wflService.WaitUntilWorkflowFinish(wflResp.WorkflowId); err != nil {
		smslog.Infof("wait workflow %s finished err %s", wflResp.WorkflowId, err.Error())
		response.message = err.Error()
		response.reason = "Failed"
		return response, err
	}

	smslog.Infof("Successfully removed pvc %s ", pvcName)
	response.reason = "Success"
	response.status = corev1.ConditionFalse
	return response, nil
}
