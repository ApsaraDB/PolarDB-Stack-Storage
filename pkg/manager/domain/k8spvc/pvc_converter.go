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
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"polardb-sms/pkg/common"
	smslog "polardb-sms/pkg/log"
	"polardb-sms/pkg/manager/domain"
	"strconv"
	"strings"
	"time"
)

const (
	ClusterName                   = "apsara.metric.ppas_name"
	ClusterAction                 = "apsara.metric.ppas_action"
	ClusterActionID               = "apsara.metric.ppas_action_id"
	ClusterRwPod                  = "apsara.metric.ppas_rw_node"
	ClusterRoPods                 = "apsara.metric.ppas_ro_nodes"
	K8sAddedCsiPlugin             = "volume.beta.kubernetes.io/storage-provisioner"
	PreProvisionedVolumeWWID      = "apsara.metric.pre_provisioned_volume_wwid"
	PreProvisionedVolumeFormat    = "apsara.metric.pre_provisioned_volume_format"
	PreProvisionedVolumeFormatted = "apsara.metric.pre_provisioned_volume_formatted"
	PreProvisionedVolumeFinalizer = "kubernetes.io/pre-provisioned-volume"
)

const (
	DefaultApiVersion  = "v1"
	DefaultPvcKind     = "PersistentVolumeClaim"
	pvcLockedLabel     = "pvc-locked-by"
	pvcUsedByLabel     = "pvc-used-by"
	pvcVolumeTypeLabel = "attach-volume-type"
	pvcPrKeyLabel      = "pr-key"
	pvcVolumeIdLabel   = "attach-volume-id"
	pvcStatusLabel     = "pvc-status"
)

var _ domain.Converter = &PvcConverter{}

type PvcConverter struct {
}

func GetPvcLockLabelSelector(wflId string) string {
	var labelWflId = wflId
	if wflId == "" {
		labelWflId = domain.DummyWorkflowId
	}
	var labelMap = map[string]string{pvcLockedLabel: labelWflId}
	return labels.SelectorFromSet(labelMap).String()
}

func GetPvcVolumeTypeLabel(volumeType string) string {
	var labelMap = map[string]string{pvcVolumeTypeLabel: volumeType}
	return labels.SelectorFromSet(labelMap).String()
}

func GetPvcVolumeIdLabelSelector(volumeId string) string {
	var labelMap = map[string]string{pvcVolumeIdLabel: volumeId}
	return labels.SelectorFromSet(labelMap).String()
}

func GetNeedProcessNamespaces(clientSet kubernetes.Interface) ([]string, error) {
	namespaceList, err := clientSet.CoreV1().Namespaces().List(metav1.ListOptions{})
	if err != nil {
		smslog.Errorf("cant not get namespaceList err %s", err.Error())
		return nil, err
	}
	var namespaces []string
	for _, namespace := range namespaceList.Items {
		if strings.HasPrefix(namespace.Name, "rp-") ||
			namespace.Name == "default" {
			namespaces = append(namespaces, namespace.Name)
		}
	}
	return namespaces, nil
}

func GetPvcListByNamespaceAndVolumeType(clientSet kubernetes.Interface, namespace string, volumeClass string) ([]corev1.PersistentVolumeClaim, error) {
	var listOption = metav1.ListOptions{}
	if volumeClass != "" {
		listOption.LabelSelector = GetPvcVolumeTypeLabel(volumeClass)
	}
	pvcList, err := clientSet.CoreV1().PersistentVolumeClaims(namespace).List(listOption)
	if err != nil {
		return nil, err
	}
	return pvcList.Items, nil
}

func GetAllPvcs(clientSet kubernetes.Interface) ([]*corev1.PersistentVolumeClaim, error) {
	namespaces, err := GetNeedProcessNamespaces(clientSet)
	if err != nil {
		smslog.Infof("Failed to Get All Namespace err %s:", err.Error())
		return nil, err
	}
	var pvcList []*corev1.PersistentVolumeClaim
	for _, namespace := range namespaces {
		tempList, err := GetPvcListByNamespaceAndVolumeType(clientSet, namespace, "")
		if err != nil {
			smslog.Warnf("skip err %s, when get pvclist from namespace %s", err.Error(), namespace)
			continue
		}
		for _, pvc := range tempList {
			pvcList = append(pvcList, pvc.DeepCopy())
		}
	}
	return pvcList, nil
}

func GetAllPvcsByVolumeType(clientSet kubernetes.Interface, volumeClass string) ([]*corev1.PersistentVolumeClaim, error) {
	namespaces, err := GetNeedProcessNamespaces(clientSet)
	if err != nil {
		smslog.Infof("Failed to Get All Namespace err %s:", err.Error())
		return nil, err
	}
	var pvcList []*corev1.PersistentVolumeClaim
	for _, namespace := range namespaces {
		tempList, err := GetPvcListByNamespaceAndVolumeType(clientSet, namespace, volumeClass)
		if err != nil {
			smslog.Infof("skip err %s, when get pvclist from namespace %s", err.Error(), namespace)
			continue
		}
		for _, pvc := range tempList {
			pvcList = append(pvcList, pvc.DeepCopy())
		}
	}
	return pvcList, nil
}

func GetAllPvcsByVolumeId(clientSet kubernetes.Interface, volumeId string) (*corev1.PersistentVolumeClaimList, error) {
	var (
		ret *corev1.PersistentVolumeClaimList
		err error
	)

	err = common.RunWithRetry(3, 1*time.Second, func(retryTimes int) error {
		ret, err = clientSet.CoreV1().PersistentVolumeClaims("").List(metav1.ListOptions{
			LabelSelector: GetPvcVolumeIdLabelSelector(volumeId),
		})
		return err
	})
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func RemovePvcFinalizer(clientSet kubernetes.Interface, pvc *corev1.PersistentVolumeClaim) error {
	return common.RunWithRetry(3, 1*time.Second, func(retryTimes int) error {
		latestPVC, err := clientSet.CoreV1().
			PersistentVolumeClaims(pvc.GetNamespace()).
			Get(pvc.GetName(), metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("failed get latest pvc: %v", err)
		}

		latestPVC.ObjectMeta.Finalizers = nil

		_, err = clientSet.CoreV1().PersistentVolumeClaims(latestPVC.Namespace).Update(latestPVC)
		if err != nil {
			return err
		}
		return nil
	})
}

func FindPvcModel(clientSet kubernetes.Interface, pvcName, namespace string) (*corev1.PersistentVolumeClaim, error) {
	var (
		ret *corev1.PersistentVolumeClaim
		err error
	)
	err = common.RunWithRetry(3, 1*time.Second, func(retryTimes int) error {
		ret, err = clientSet.CoreV1().PersistentVolumeClaims(namespace).Get(pvcName, metav1.GetOptions{})
		return err
	})
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func FindPvcModelByVolumeId(clientSet kubernetes.Interface, volumeId, pvcName string) (*corev1.PersistentVolumeClaim, error) {
	var (
		ret *corev1.PersistentVolumeClaim
		err error
	)
	err = common.RunWithRetry(3, 1*time.Second, func(retryTimes int) error {
		pvcList, err := GetAllPvcsByVolumeId(clientSet, volumeId)
		if err != nil {
			return err
		}
		if len(pvcList.Items) > 1 {
			return fmt.Errorf("volumeId %s bound pvc %d is not equal to 1: %v", volumeId, len(pvcList.Items), pvcList.Items)
		}
		for _, item := range pvcList.Items {
			if item.Name == pvcName {
				ret = &item
			}
		}
		return err
	})
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func DeletePvcModel(clientSet kubernetes.Interface, pvcName, namespace string) error {
	return common.RunWithRetry(3, 1*time.Second, func(retryTimes int) error {
		return clientSet.CoreV1().PersistentVolumeClaims(namespace).Delete(pvcName, &metav1.DeleteOptions{})
	})
}

func UpdatePvcLabel(clientSet kubernetes.Interface, pvcModel *corev1.PersistentVolumeClaim, key, value string) error {
	smslog.Debugf("update pvc [%s] label [%s] value [%s]", pvcModel.Name, key, value)
	return common.RunWithRetry(3, 1*time.Second, func(retryTimes int) error {
		latestPvcModel, err := FindPvcModel(clientSet, pvcModel.Name, pvcModel.Namespace)
		if err != nil {
			return err
		}
		labelMap := getPvcLabelsFromPvcModel(latestPvcModel)
		labelMap[key] = value
		latestPvcModel.Labels = labelMap
		_, err = clientSet.CoreV1().PersistentVolumeClaims(pvcModel.Namespace).Update(latestPvcModel)
		return err
	})
}

func BatchUpdatePvcLabel(clientSet kubernetes.Interface, pvcModel *corev1.PersistentVolumeClaim, batchUpdate map[string]string) error {
	return common.RunWithRetry(3, 1*time.Second, func(retryTimes int) error {
		pvcModel, err := FindPvcModel(clientSet, pvcModel.Name, pvcModel.Namespace)
		if err != nil {
			return err
		}
		labelMap := getPvcLabelsFromPvcModel(pvcModel)
		for key, value := range batchUpdate {
			labelMap[key] = value
		}
		pvcModel.Labels = labelMap
		_, err = clientSet.CoreV1().PersistentVolumeClaims(pvcModel.Namespace).Update(pvcModel)
		return err
	})
}

func UpdatePvcWorkflowLabel(clientSet kubernetes.Interface, pvcModel *corev1.PersistentVolumeClaim, workflowId string) error {
	return UpdatePvcLabel(clientSet, pvcModel, pvcLockedLabel, workflowId)
}

func UpdateUsedDbCLusterLabel(clientSet kubernetes.Interface, pvcModel *corev1.PersistentVolumeClaim, dbClusterName string) error {
	return BatchUpdatePvcLabel(clientSet, pvcModel,
		map[string]string{
			pvcUsedByLabel: dbClusterName,
		})
}

func UpdatePrKeyLabel(clientSet kubernetes.Interface, pvcModel *corev1.PersistentVolumeClaim, prKey string) error {
	return UpdatePvcLabel(clientSet, pvcModel, pvcPrKeyLabel, prKey)
}

func UpdatePvcStatusLabel(clientSet kubernetes.Interface, pvcModel *corev1.PersistentVolumeClaim, status int) error {
	return UpdatePvcLabel(clientSet, pvcModel, pvcStatusLabel, strconv.FormatInt(int64(status), 10))
}

func getPvcLabelsFromPvcModel(pvcModel *corev1.PersistentVolumeClaim) map[string]string {
	ret := pvcModel.GetLabels()
	if ret == nil {
		ret = make(map[string]string)
	}
	return ret
}

func getPvcLabelFromPvcEntity(pvcEntity *PersistVolumeClaimEntity) map[string]string {
	var pvcLabels = make(map[string]string)
	pvcLabels[pvcLockedLabel] = pvcEntity.RelatedWorkflow
	pvcLabels[pvcVolumeTypeLabel] = string(pvcEntity.GetVolumeType())
	pvcLabels[pvcUsedByLabel] = pvcEntity.DbClusterName
	pvcLabels[pvcVolumeIdLabel] = pvcEntity.GetVolumeId()
	pvcLabels[pvcPrKeyLabel] = pvcEntity.GetPrKey()
	pvcLabels[pvcStatusLabel] = strconv.FormatInt(int64(pvcEntity.PvcStatus.StatusValue), 10)
	return pvcLabels
}

func UpdatePvcRequestCapacity(clientSet kubernetes.Interface, name, namespace, size string) error {
	return common.RunWithRetry(3, 1*time.Second, func(retryTimes int) error {
		pvc, err := FindPvcModel(clientSet, name, namespace)
		if err != nil {
			return err
		}
		newQuantity, err := resource.ParseQuantity(size)
		if err != nil {
			return err
		}
		if pvc.Spec.Resources.Requests == nil {
			pvc.Spec.Resources.Requests = make(map[corev1.ResourceName]resource.Quantity)
		}
		if statusStorage, exist := pvc.Spec.Resources.Requests[corev1.ResourceStorage]; exist && statusStorage.Equal(newQuantity) {
			return nil
		}
		pvc.Spec.Resources.Requests[corev1.ResourceStorage] = newQuantity
		_, err = clientSet.CoreV1().PersistentVolumeClaims(namespace).Update(pvc)
		return err
	})
}

func UpdatePvcCapacity(clientSet kubernetes.Interface, name, namespace, size string) error {
	return common.RunWithRetry(3, 1*time.Second, func(retryTimes int) error {
		pvc, err := FindPvcModel(clientSet, name, namespace)
		if err != nil {
			return err
		}
		newQuantity, err := resource.ParseQuantity(size)
		if err != nil {
			return err
		}
		if pvc.Status.Capacity == nil {
			pvc.Status.Capacity = make(map[corev1.ResourceName]resource.Quantity)
		}
		if statusStorage, exist := pvc.Status.Capacity[corev1.ResourceStorage]; exist && statusStorage.Equal(newQuantity) {
			return nil
		}
		pvc.Status.Capacity[corev1.ResourceStorage] = newQuantity
		_, err = clientSet.CoreV1().PersistentVolumeClaims(namespace).UpdateStatus(pvc)
		return err
	})
}

func UpdatePvcStatus(clientSet kubernetes.Interface, name, namespace, status string) error {
	return common.RunWithRetry(3, 1*time.Second, func(retryTimes int) error {
		pvc, err := FindPvcModel(clientSet, name, namespace)
		if err != nil {
			return err
		}
		pvc.Status.Phase = corev1.PersistentVolumeClaimPhase(status)
		_, err = clientSet.CoreV1().PersistentVolumeClaims(namespace).UpdateStatus(pvc)
		return err
	})
}

func (c *PvcConverter) ToModel(t interface{}) (interface{}, error) {
	pvcEntity := t.(*PersistVolumeClaimEntity)
	var volumeMode corev1.PersistentVolumeMode
	switch pvcEntity.DiskStatus.VolumeMode {
	case Block:
		volumeMode = corev1.PersistentVolumeBlock
	case FsExt4:
		volumeMode = corev1.PersistentVolumeFilesystem
	}

	quantity, err := resource.ParseQuantity(common.BytesToGiBString(pvcEntity.ExpectedDiskStatus.Size))
	if err != nil {
		smslog.Errorf("ToModel: parse quantity error %s", err.Error())
		return nil, err
	}
	var requests = make(map[corev1.ResourceName]resource.Quantity)

	requests[corev1.ResourceStorage] = quantity
	pvcModel := &corev1.PersistentVolumeClaim{
		TypeMeta: metav1.TypeMeta{
			Kind:       DefaultPvcKind,
			APIVersion: DefaultApiVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      pvcEntity.Name,
			Namespace: pvcEntity.Namespace,
			Labels:    getPvcLabelFromPvcEntity(pvcEntity),
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			Resources: corev1.ResourceRequirements{
				Requests: requests,
			},
			VolumeMode:       &volumeMode,
			AccessModes:      []corev1.PersistentVolumeAccessMode{corev1.ReadWriteMany},
			StorageClassName: &pvcEntity.StorageClassName,
		},
	}
	return pvcModel, nil
}

func getVolumeTypeFromLabel(pvc *corev1.PersistentVolumeClaim) (string, error) {
	labelMap := pvc.GetLabels()
	val, exist := labelMap[pvcVolumeTypeLabel]
	if !exist || val == "" {
		return "", fmt.Errorf("cannot find volumeType from %s", pvc.Name)
	}
	return val, nil
}

func getVolumeIdFromLabel(pvc *corev1.PersistentVolumeClaim) (string, error) {
	labelMap := pvc.GetLabels()
	val, exist := labelMap[pvcVolumeIdLabel]
	if !exist || val == "" {
		return "", fmt.Errorf("cannot find wwid from %s", pvc.Name)
	}
	return val, nil
}

func getVolumeIdFromAnnotation(pvc *corev1.PersistentVolumeClaim) (string, error) {
	annotationMap := pvc.GetAnnotations()
	val, exist := annotationMap[PreProvisionedVolumeWWID]
	if !exist || val == "" {
		return "", fmt.Errorf("cannot find wwid from %s", pvc.Name)
	}
	return fmt.Sprintf("3%s", strings.ToLower(val)), nil
}

func getPvcStatus(pvc *corev1.PersistentVolumeClaim) (*domain.VolumeStatus, error) {
	labelMap := pvc.GetLabels()
	val, exist := labelMap[pvcStatusLabel]
	if !exist || val == "" {
		return nil, fmt.Errorf("cannot find pvc status from %s", pvc.Name)
	}
	statusVal, err := strconv.Atoi(val)
	if err != nil {
		return nil, err
	}
	pvcStatus := domain.VolumeStatus{
		StatusValue: domain.VolumeStatusValue(statusVal),
	}
	return &pvcStatus, nil
}
func (c *PvcConverter) ToEntity(t interface{}) (interface{}, error) {
	pvc := t.(*corev1.PersistentVolumeClaim)

	var volumeMode RequestVolumeMode
	switch *pvc.Spec.VolumeMode {
	case corev1.PersistentVolumeBlock:
		volumeMode = Block
	case corev1.PersistentVolumeFilesystem:
		volumeMode = FsExt4
	}

	volumeId, err := getVolumeIdFromLabel(pvc)
	if err != nil {
		volumeId, err = getVolumeIdFromAnnotation(pvc)
		if err != nil {
			return nil, err
		}
	}
	pvName := fmt.Sprintf("pv-%s", volumeId)

	reqQuantity := pvc.Spec.Resources.Requests[corev1.ResourceStorage]
	realQuantity := pvc.Status.Capacity[corev1.ResourceStorage]
	needFormatStr, ok := pvc.Annotations[PreProvisionedVolumeFormat]
	needFormat := false
	if !ok {
		needFormat = false
	} else {
		needFormat = needFormatStr == "true"
	}

	var volumeType common.LvType
	volumeTypeStr, err := getVolumeTypeFromLabel(pvc)
	if err != nil {
		volumeType = common.MultipathVolume
	} else {
		volumeType = common.LvType(volumeTypeStr)
	}

	diskRequest := &VolumeMeta{
		VolumeId: volumeId,
		//通过pvc过来的先定位Lun
		VolumeType: volumeType,
		Size:       (&reqQuantity).Value(),
		VolumeMode: volumeMode,
		NeedFormat: needFormat,
	}

	diskStatus := &VolumeMeta{
		VolumeId: volumeId,
		//通过pvc过来的先定位Lun
		VolumeType: volumeType,
		Size:       (&realQuantity).Value(),
		VolumeMode: volumeMode,
		NeedFormat: needFormat,
	}

	//todo refactor this
	prKeyLabel, ok := pvc.Labels[pvcPrKeyLabel]
	if ok {
		diskStatus.PrKey = prKeyLabel
	}
	clusterName, ok := pvc.Labels[pvcUsedByLabel]
	if !ok {
		clusterName = pvc.Annotations[ClusterName]
	}
	relatedWorkflow, _ := pvc.Labels[pvcLockedLabel]

	pvcStatus, err := getPvcStatus(pvc)
	if err != nil {
		smslog.Errorf("can not get PvcStatus err %s", err.Error())
		pvcStatus = &domain.VolumeStatus{}
	}

	return &PersistVolumeClaimEntity{
		Name:               pvc.Name,
		PvName:             pvName,
		Namespace:          pvc.Namespace,
		RuntimeStatus:      PvcRuntimeStatus(pvc.Status.Phase),
		DiskStatus:         diskStatus,
		ExpectedDiskStatus: diskRequest,
		StorageClassName:   *pvc.Spec.StorageClassName,
		DbClusterName:      clusterName,
		PvcStatus:          *pvcStatus,
		RelatedWorkflow:    relatedWorkflow,
		CreateTime:         strconv.FormatInt(pvc.CreationTimestamp.Unix(), 10),
	}, nil
}

func (c *PvcConverter) ToEntities(t []interface{}) ([]interface{}, error) {
	return nil, nil
}

func NewPvcConverter() domain.Converter {
	return &PvcConverter{}
}
