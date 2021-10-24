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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	smslog "polardb-sms/pkg/log"
	"polardb-sms/pkg/manager/application/service"
	"time"

	corev1 "k8s.io/api/core/v1"
	kubeinformers "k8s.io/client-go/informers"
	coreinformers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

const (
	works         = 5
	defaultReSync = 30 * time.Second
)

type StorageControllerConfig struct {
	NodeID          string
	NodeIP          string
	ClientSet       kubernetes.Interface
	InformerFactory kubeinformers.SharedInformerFactory
	WorkerCnt       int
	StopChan        <-chan struct{}
}

func NewControllerConfig(stopCh <-chan struct{}, nodeId, nodeIp string, clientSet kubernetes.Interface) StorageControllerConfig {
	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(clientSet, defaultReSync)
	c := StorageControllerConfig{
		NodeID:          nodeId,
		NodeIP:          nodeIp,
		ClientSet:       clientSet,
		InformerFactory: kubeInformerFactory,
		WorkerCnt:       works,
		StopChan:        stopCh,
	}
	return c
}

type StorageController struct {
	StorageControllerConfig
	pvcInformer   coreinformers.PersistentVolumeClaimInformer
	processor     *PvcEventProcessor
	innerStopChan chan struct{}
}

func NewStorageController(config StorageControllerConfig) *StorageController {
	pvcInformer := config.InformerFactory.Core().V1().PersistentVolumeClaims()

	controller := &StorageController{
		StorageControllerConfig: config,
		pvcInformer:             pvcInformer,
	}

	pvcHandler := func(obj interface{}) {
		pvc, _ := obj.(*corev1.PersistentVolumeClaim)
		annotations := pvc.GetAnnotations()
		action := annotations[ClusterAction]
		requestID := annotations[ClusterActionID]
		pvcEvent := &PvcEvent{
			RequestID: requestID,
			Action:    action,
			Pvc:       pvc.DeepCopy(),
		}
		if !DuplicatedPvcEvent(pvc, pvcEvent) {
			controller.processor.Process(pvcEvent)
		} else {
			smslog.Infof("duplicated request pvc %s action %s actionId %s", pvc.Name, action, requestID)
		}
	}

	pvcInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: func(obj interface{}) bool {
			pvc, ok := obj.(*corev1.PersistentVolumeClaim)
			if !ok {
				return false
			}
			return ValidPvc(pvc)
		},
		Handler: cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				if _, ok := obj.(*corev1.PersistentVolumeClaim); !ok {
					return
				}
				pvcHandler(obj)
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				newPvc := newObj.(*corev1.PersistentVolumeClaim)
				oldPvc := oldObj.(*corev1.PersistentVolumeClaim)
				if newPvc.ResourceVersion == oldPvc.ResourceVersion {
					return
				}
				if pvcAnnotationsEqual(oldPvc.GetAnnotations(), newPvc.GetAnnotations()) {
					smslog.Infof("Annotations not changed, drop it")
					return
				}
				smslog.Infof("Update happened new: %s, version %s, old: %s, version %s", newPvc.Name, newPvc.GetResourceVersion(), oldPvc.Name, oldPvc.GetResourceVersion())
				pvcHandler(newObj)
			},
		},
	})

	return controller
}

func ValidPvc(pvc *corev1.PersistentVolumeClaim) bool {
	storageClass := pvc.Spec.StorageClassName
	if storageClass != nil && *storageClass != StorageClassName {
		return false
	}
	requiredKeys := []string{
		ClusterName,
		ClusterAction,
		ClusterActionID,
	}
	annotations := pvc.GetAnnotations()
	for _, key := range requiredKeys {
		if _, ok := annotations[key]; !ok {
			return false
		}
	}
	if pvc.DeletionTimestamp != nil {
		return false
	}
	return true
}

func pvcAnnotationsEqual(src, tgt map[string]string) bool {
	requiredKeys := []string{
		ClusterName,
		ClusterAction,
		ClusterActionID,
	}
	for _, key := range requiredKeys {
		if src[key] != tgt[key] {
			return false
		}
	}
	return true
}

func GetPvcProcessResult(pvc *corev1.PersistentVolumeClaim) (*PvcProcessResult, error) {
	if ValidPvc(pvc) {
		annotations := pvc.GetAnnotations()
		status, exist := annotations[PvcProcessStatus]
		if exist {
			pvcStatus := &PvcProcessResult{}
			err := json.Unmarshal([]byte(status), pvcStatus)
			if err != nil {
				smslog.Infof("json.Unmarshal PvcProcessResult %s err %s", status, err.Error())
				return nil, fmt.Errorf("GetPvcProcessResult err %s", err.Error())
			}
			return pvcStatus, nil
		}
		return NonProcess, nil
	}
	return nil, fmt.Errorf("GetPvcProcessResult err: is not valid pvc %s", pvc.Name)
}

func DuplicatedPvcEvent(pvc *corev1.PersistentVolumeClaim, pvcEvent *PvcEvent) bool {
	pvcProcessResult, err := GetPvcProcessResult(pvc)
	if err != nil || pvcProcessResult == NonProcess {
		return false
	}
	if pvcProcessResult.Action != pvcEvent.Action {
		return false
	}
	if pvcProcessResult.Action == pvcEvent.Action &&
		pvcProcessResult.RequestId == pvcEvent.RequestID &&
		pvcProcessResult.Result == PvcEventProcessing {
		return false
	}

	if pvcProcessResult.RequestId >= pvcEvent.RequestID &&
		pvcProcessResult.Action == pvcEvent.Action {
		//pvcProcessResult.Result != handlers.PvcEventProcessFail {
		return true
	}

	return false
}

func GetAllNamespace(clientSet kubernetes.Interface) ([]string, error) {
	namespaceList, err := clientSet.CoreV1().Namespaces().List(metav1.ListOptions{})
	if err != nil {
		smslog.Errorf("cant not get namespaceList err %s", err.Error())
		return nil, err
	}
	var namespaces []string
	for _, namespace := range namespaceList.Items {
		namespaces = append(namespaces, namespace.Name)
	}
	return namespaces, nil
}

func GetAllPvcs(clientSet kubernetes.Interface) ([]corev1.PersistentVolumeClaim, error) {
	namespaces, err := GetAllNamespace(clientSet)
	if err != nil {
		return nil, err
	}
	var pvcs []corev1.PersistentVolumeClaim
	for _, namespace := range namespaces {
		pvcList, err := clientSet.CoreV1().PersistentVolumeClaims(namespace).List(metav1.ListOptions{})
		if err != nil {
			smslog.Infof("get pvclist err from namespace %s", namespace)
			continue
		}
		pvcs = append(pvcs, pvcList.Items...)
	}
	return pvcs, nil
}

func (c *StorageController) processUnfinishedAction() {
	pvcs, err := GetAllPvcs(c.ClientSet)
	if err != nil {
		smslog.Infof("can not processUnfinishedAction, for get pvcs err %s", err.Error())
		return
	}
	for _, pvc := range pvcs {
		if ValidPvc(&pvc) {
			annotations := pvc.GetAnnotations()
			action := annotations[ClusterAction]
			requestID := annotations[ClusterActionID]
			pvcEvent := &PvcEvent{
				RequestID: requestID,
				Action:    action,
				Pvc:       pvc.DeepCopy(),
			}
			if !DuplicatedPvcEvent(&pvc, pvcEvent) {
				smslog.Infof("processUnfinishedAction pvc %s event action %s", pvc.Name, pvcEvent.Action)
				c.processor.Process(pvcEvent)
			}
		}
	}
}

func (c *StorageController) Run() {
	defer smslog.LogPanic()
	c.innerStopChan = make(chan struct{})
	c.processor = NewPvcEventProcessor(c.NodeID, c.NodeIP, c.WorkerCnt, c.ClientSet)
	//Run pvc informer
	go c.InformerFactory.Start(c.innerStopChan)
	// Wait for the caches to be synced before starting workers
	if ok := cache.WaitForCacheSync(c.innerStopChan, c.pvcInformer.Informer().HasSynced); !ok {
		smslog.Fatal("Could not wait for caches to sync")
	}

	//c.processUnfinishedAction()
	c.processor.Run(c.innerStopChan)

	var clean = &PreProvisionedDeleteHandler{
		ClientSet:  c.ClientSet,
		pvcService: service.NewPvcService(),
		wflService: service.NewWorkflowService(),
	}
	go clean.DeletePreProvisionedVolumeLoop(c.innerStopChan)
}

func (c *StorageController) Stop() {
	smslog.Infof("Stop StorageController")
	if c.innerStopChan != nil {
		close(c.innerStopChan)
		c.innerStopChan = nil
	}
	if c.processor != nil && c.processor.pvcEventQueues != nil {
		for _, q := range c.processor.pvcEventQueues {
			q.ShutDown()
		}
	}
	if c.processor != nil && c.processor.eventCh != nil {
		close(c.processor.eventCh)
	}
	c.processor = nil
}

func (c *StorageController) Identify() string {
	return "storageController"
}
