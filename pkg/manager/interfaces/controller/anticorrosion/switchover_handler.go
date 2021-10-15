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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"polardb-sms/pkg/common"
	smslog "polardb-sms/pkg/log"
	"polardb-sms/pkg/manager/application/service"
	"polardb-sms/pkg/manager/application/view"
	"polardb-sms/pkg/manager/config"
	"strings"
)

type SwitchOverHandler struct {
	clientSet  kubernetes.Interface
	wflService *service.WorkflowService
	pvcService *service.PvcService
	lunService *service.LvForOldLunService
}
type HostInfo struct {
	ID          string `json:"ID"`
	IP          string `json:"IP"`
	Name        string `json:"NAME"`
	Type        string `json:"TYPE"`
	Username    string `json:"USERNAME"`
	Password    string `json:"PASSWORD"`
	SanHostname string `json:"SAN_HOST_NAME"`
}

type StoragePod struct {
	HostIp   string
	PodName  string
	NodeName string
}

//目前只用到rwIp
type ClusterTopology struct {
	RwPod        *StoragePod
	RoPods       []*StoragePod
	RoPodHostIps map[string]bool
}

func GetClusterRwByPvc(clientSet kubernetes.Interface, pvc *corev1.PersistentVolumeClaim) (*ClusterTopology, error) {
	annotations := pvc.GetAnnotations()

	clusterName, ok := annotations[ClusterName]
	if !ok {
		return nil, fmt.Errorf("could not get pvc annotation %s", ClusterName)
	}

	rwPodName, ok := annotations[ClusterRwPod]
	if !ok {
		return nil, fmt.Errorf("could not get pvc annotation %s", ClusterRwPod)
	}

	roNodes, ok := annotations[ClusterRoPods]
	if !ok {
		return nil, fmt.Errorf("could not get pvc annotation %s", ClusterRoPods)
	}

	var roPodNames []string
	if err := json.Unmarshal([]byte(roNodes), &roPodNames); err != nil {
		return nil, fmt.Errorf("failed parse annotation: %v", err)
	}

	podList, err := clientSet.CoreV1().Pods(pvc.GetNamespace()).
		List(metav1.ListOptions{LabelSelector: fmt.Sprintf("apsara.metric.clusterName=%s", clusterName)})
	if err != nil {
		return nil, fmt.Errorf("failed list pods with ppascluster %q: %v", clusterName, err)
	}

	topology := ClusterTopology{
		RoPods:       make([]*StoragePod, 0),
		RoPodHostIps: make(map[string]bool),
	}

	for _, pod := range podList.Items {
		dumpPod := &StoragePod{
			PodName:  pod.GetName(),
			HostIp:   pod.Status.HostIP,
			NodeName: pod.Spec.NodeName,
		}
		if dumpPod.PodName == rwPodName {
			topology.RwPod = dumpPod
		} else {
			topology.RoPodHostIps[dumpPod.HostIp] = true
			topology.RoPods = append(topology.RoPods, dumpPod)
		}
	}

	if topology.RwPod == nil {
		return nil, fmt.Errorf("could not find rw pod with pvc(%s) and rw pod(%s)", pvc.GetName(), rwPodName)
	}
	return &topology, nil
}

func (h *SwitchOverHandler) constructSwitchItems(clusterTopo *ClusterTopology, switchError error, prKey string) []*PvcSwitchItem {
	var (
		details []*PvcSwitchItem
	)

	switchOverTask := &PvcSwitchItem{
		PodName:      clusterTopo.RwPod.PodName,
		Host:         clusterTopo.RwPod.HostIp,
		TargetStatus: SwitchModeRw,
		SwitchStatus: SwitchStatusSuccess,
		OrgStatus:    SwitchModeRo,
		ErrMsg:       "",
	}

	if prKey == common.IpV4ToPrKey(clusterTopo.RwPod.HostIp) {
		switchOverTask.OrgStatus = SwitchModeRw
	}

	if switchError != nil {
		switchOverTask.ErrMsg = switchError.Error()
		switchOverTask.SwitchStatus = SwitchStatusFailed
	}

	details = append(details, switchOverTask)
	for _, roPod := range clusterTopo.RoPods {
		details = append(details, &PvcSwitchItem{
			PodName:      roPod.PodName,
			Host:         roPod.HostIp,
			OrgStatus:    SwitchModeRo,
			TargetStatus: SwitchModeRo,
			SwitchStatus: SwitchStatusSuccess,
			ErrMsg:       "",
		})
	}
	return details
}

//TODO fix this
func (h *SwitchOverHandler) handlePvcEvent(ctx common.TraceContext, pvcEvent *PvcEvent) (pvcResponse *PvcResponse, err error) {
	var (
		pvc      = pvcEvent.Pvc
		response = &PvcResponse{
			message:       "",
			reason:        PvcEventProcessFail,
			conditionType: PVCSwitchover,
			status:        corev1.ConditionFalse,
			pvc:           pvc,
		}
		pvName = pvc.Spec.VolumeName
		wwid   string
	)

	err = UpdateConditionsByResponse(ctx, pvc,
		&PvcResponse{
			message:       "",
			reason:        PvcEventProcessSwitching,
			conditionType: PVCSwitchover,
			status:        corev1.ConditionTrue,
			pvc:           pvc,
		},
		h.clientSet)
	if err != nil {
		smslog.WithContext(ctx).Infof("UpdateConditionsByResponse pvc %s, err %s", pvc.Name, err.Error())
		response.message = err.Error()
		return response, err
	}
	rwNodeIp, err := h.findRwPodNode(pvc)
	if err != nil {
		smslog.WithContext(ctx).Infof("can not find rwNodeIp for pvc %s, err %s", pvc.Name, err.Error())
		response.message = err.Error()
		return response, err
	}
	rwNode := config.GetNodeByIp(rwNodeIp)
	if rwNode == nil {
		err = fmt.Errorf("can not find rwNode by ip %s", rwNodeIp)
		smslog.WithContext(ctx).Info(err.Error())
		response.message = err.Error()
		return response, err
	}

	clusterTopo, err := GetClusterRwByPvc(h.clientSet, pvc)
	if err != nil {
		smslog.WithContext(ctx).Info(err.Error())
		response.message = err.Error()
		return response, err
	}

	wwid = strings.TrimPrefix(pvName, "pv-")
	prKey, err := h.lunService.FindLunPrKey(wwid)
	if err != nil {
		smslog.WithContext(ctx).Infof("Finde prKey err %s", err.Error())
		response.message = err.Error()
		return response, err
	}
	wflResp, err := h.pvcService.SetVolumeWriteLock(ctx, &view.PvcWriteLockRequest{
		PvcRequest: view.PvcRequest{
			Name:      pvc.Name,
			Namespace: pvc.Namespace,
		},
		WriteLockNodeId: rwNode.Name,
		WriteLockNodeIp: rwNodeIp,
	})
	if err != nil {
		smslog.WithContext(ctx).Infof("call SetVolumeWriteLock err %s", err.Error())
		response.message = err.Error()
		return response, err
	}

	if err := h.wflService.WaitUntilWorkflowFinish(wflResp.WorkflowId); err != nil {
		result := &PvcSwitch{
			ActionId: pvcEvent.RequestID,
			Details:  h.constructSwitchItems(clusterTopo, err, prKey),
			ErrMsg:   "",
		}
		response.pvcSwitchResult = result
		response.message = err.Error()
		return response, err
	}

	result := &PvcSwitch{
		ActionId: pvcEvent.RequestID,
		Details:  h.constructSwitchItems(clusterTopo, nil, prKey),
		ErrMsg:   "",
	}
	response.pvcSwitchResult = result
	smslog.WithContext(ctx).Debugf("construct switchOverResult: %s", result.String(false))

	response.reason = PvcEventProcessSuccess
	response.status = corev1.ConditionFalse
	return response, nil
}

func (h *SwitchOverHandler) findRwPodNode(pvc *corev1.PersistentVolumeClaim) (string, error) {
	rwPodName := pvc.Annotations[PolarBoxInstanceRW]
	rwPod, err := h.clientSet.CoreV1().Pods(pvc.Namespace).Get(rwPodName, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	return rwPod.Status.HostIP, nil
}
