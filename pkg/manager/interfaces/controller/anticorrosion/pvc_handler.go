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
	"hash"
	"hash/fnv"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	smslog "polardb-sms/pkg/log"
	"polardb-sms/pkg/manager/application/service"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/workqueue"
	"polardb-sms/pkg/common"
)

//pvc action value
const (
	//fuck 3 string styles
	PreProvisionVolume    = "pre_provision_volume"
	SwitchOver            = "switchover"
	UpdateRO              = "updatero"
	GrowFs                = "growfs"
	PreProvisionDelVolume = "pre_provision_del_volume"
	FormatFs              = "FormatFs"
	LockVolume            = "LockVolume"
	FormatAndLockVolume   = "FormatAndLockVolume"
)

const (
	PVCPreProvisionedVolume corev1.PersistentVolumeClaimConditionType = "PreProvisionedVolume"
	PVCSwitchover           corev1.PersistentVolumeClaimConditionType = "SwitchRWO"
	PVCGrowFs               corev1.PersistentVolumeClaimConditionType = "GrowFs"
	PVCFormatFs             corev1.PersistentVolumeClaimConditionType = "FormatFs"
	PVCLockVolume           corev1.PersistentVolumeClaimConditionType = "LockVolume"
	PVCFormatAndLockVolume  corev1.PersistentVolumeClaimConditionType = "FormatAndLockVolume"
)

var NonProcess = &PvcProcessResult{}

type PvcProcessResult struct {
	RequestId string                `json:"request_id"`
	Action    string                `json:"action"`
	Result    PvcEventProcessResult `json:"result"`
	Timestamp string                `json:"timestamp"`
	Msg       string                `json:"msg"`
}

func (r *PvcProcessResult) String() string {
	result, err := json.Marshal(r)
	if err != nil {
		smslog.Errorf("json.Marshal PvcProcessResult %v err %s", r, err.Error())
		return ""
	}
	return string(result)
}

type PvcEvent struct {
	RequestID string
	Action    string
	Pvc       *corev1.PersistentVolumeClaim
}

func (e *PvcEvent) String() string {
	return fmt.Sprintf("[requestID: %s, action: %s, pvc: %s]", e.RequestID, e.Action, e.Pvc.GetName())
}
func (e *PvcEvent) Map() map[string]string {
	return map[string]string{
		"action":   e.Action,
		"actionId": e.RequestID,
		"pvc":      e.Pvc.Name,
	}
}

//compact with old protocol
type PvcSwitch struct {
	ActionId string      `json:"actionId"`
	Details  interface{} `json:"details"`
	ErrMsg   string      `json:"errMsg"`
}

type PvcSwitchItem struct {
	PodName      string `json:"podName"`
	Host         string `json:"host"`
	OrgStatus    string `json:"orgStatus"`
	TargetStatus string `json:"targetStatus"`
	SwitchStatus string `json:"switchStatus"`
	ErrMsg       string `json:"errMsg"`
}

func (p *PvcSwitch) String(pretty bool) string {
	var (
		msg []byte
		err error
	)
	if pretty {
		msg, err = json.MarshalIndent(p, "", "  ")
	} else {
		msg, err = json.Marshal(p)
	}

	if err != nil {
		return ""
	}
	return string(msg)
}

type PvcEventProcessResult string

const (
	PvcEventProcessFail      PvcEventProcessResult = "Failed"
	PvcEventProcessSuccess   PvcEventProcessResult = "Success"
	PvcEventProcessSwitching PvcEventProcessResult = "switching"
	PvcEventProcessing       PvcEventProcessResult = "processing"
)

type PvcResponse struct {
	reason          PvcEventProcessResult
	message         string
	pvc             *corev1.PersistentVolumeClaim
	status          corev1.ConditionStatus
	conditionType   corev1.PersistentVolumeClaimConditionType
	pvcSwitchResult *PvcSwitch
}

func (p *PvcResponse) String() string {
	return fmt.Sprintf("reason: %s, message:%s, pvc:%s, conditionType:%s, status:%s", p.reason, p.message, p.pvc.GetName(), p.conditionType, p.status)
}

type EventHandlerInterface interface {
	handlePvcEvent(ctx common.TraceContext, pvcEvent *PvcEvent) (pvcResponse *PvcResponse, err error)
}

type PvcEventProcessor struct {
	handlers       map[string]EventHandlerInterface
	pvcEventQueues []workqueue.RateLimitingInterface
	workerCnt      int
	clientSet      kubernetes.Interface
	hasher         hash.Hash32
	eventCh        chan *PvcEvent
}

func NewPvcEventProcessor(nodeId string, nodeIp string, workerCnt int, clientSet kubernetes.Interface) *PvcEventProcessor {
	var (
		lunService = service.NewLvForOldLunService()
		wflService = service.NewWorkflowService()
		pvcService = service.NewPvcService()
	)
	switchOverHandler := &SwitchOverHandler{
		wflService: wflService,
		clientSet:  clientSet,
		pvcService: pvcService,
		lunService: lunService,
	}

	pvcHandlers := map[string]EventHandlerInterface{
		PreProvisionVolume: &CreateHandler{
			nodeId:     nodeId,
			nodeIp:     nodeIp,
			clientSet:  clientSet,
			pvcService: pvcService,
			wflService: wflService,
		},
		SwitchOver: switchOverHandler,
		UpdateRO:   switchOverHandler,
		GrowFs: &FsExpandHandler{
			clientSet:    clientSet,
			lunService:   lunService,
			wflService:   wflService,
			pvcService:   pvcService,
			growFsLock:   sync.RWMutex{},
			requestCache: make(map[string]resource.Quantity),
		},
		FormatFs: &FsFormatHandler{
			lunService: lunService,
			wflService: wflService,
			clientSet:  clientSet,
			nodeIp:     nodeIp,
			nodeId:     nodeId,
		},
		LockVolume: &VolumeLockHandler{
			clientSet:  clientSet,
			nodeId:     nodeId,
			nodeIp:     nodeIp,
			lunService: lunService,
			wflService: wflService,
		},
		FormatAndLockVolume: &FormatAndLockHandler{
			clientSet:  clientSet,
			nodeId:     nodeId,
			nodeIp:     nodeIp,
			lunService: lunService,
			wflService: wflService,
		},
	}

	eventQueues := make([]workqueue.RateLimitingInterface, 0)
	for i := 0; i < workerCnt; i++ {
		eventQueues = append(eventQueues, workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "pvcEvent-"+fmt.Sprint(i)))
	}
	return &PvcEventProcessor{
		handlers:       pvcHandlers,
		pvcEventQueues: eventQueues,
		workerCnt:      workerCnt,
		clientSet:      clientSet,
		hasher:         fnv.New32(),
		eventCh:        make(chan *PvcEvent),
	}
}

func (h *PvcEventProcessor) Run(stopCh chan struct{}) {
	for i := 0; i < h.workerCnt; i++ {
		go h.receive(i)
		go h.processPvcEvent(i, stopCh)
	}
}

func (h *PvcEventProcessor) Process(e *PvcEvent) {
	pvc := e.Pvc
	h.hasher.Reset()
	_, err := h.hasher.Write([]byte(pvc.Name))
	if err != nil {
		smslog.Infof("Error happened when computing hash: %v", err)
		h.pvcEventQueues[0].Add(e)
	}
	idx := h.hasher.Sum32() % uint32(h.workerCnt)
	h.pvcEventQueues[idx].Add(e)
}

func (h *PvcEventProcessor) receive(idx int) {
	for {
		eventQueue := h.pvcEventQueues[idx]
		obj, shutdown := eventQueue.Get()
		if shutdown {
			smslog.Info("pvc event queue shutdown")
			return
		}
		pvcEvent, ok := obj.(*PvcEvent)
		if !ok {
			smslog.Errorf("Could not handle obj %s, ignore it", obj)
			eventQueue.Done(obj)
			continue
		}
		h.eventCh <- pvcEvent
	}
}

func (h *PvcEventProcessor) processPvcEvent(idx int, stopCh chan struct{}) {
	defer smslog.LogPanic()
	eventQueue := h.pvcEventQueues[idx]
	for {
		select {
		case <-stopCh:
			smslog.Infof("stop PvcEventProcessor")
			eventQueue.ShutDown()
			close(h.eventCh)
			return
		case pvcEvent := <-h.eventCh:
			ctx := common.NewTraceContext(pvcEvent.Map())
			pvcHandler, ok := h.handlers[pvcEvent.Action]
			if !ok {
				smslog.WithContext(ctx).Errorf("Could not support action %s for pvc: %s, ignore it", pvcEvent.Action, pvcEvent.Pvc.GetName())
				eventQueue.Done(pvcEvent)
				continue
			}
			err := h.updatePvcProcessResult(&PvcProcessResult{
				RequestId: pvcEvent.RequestID,
				Action:    pvcEvent.Action,
				Result:    PvcEventProcessing,
				Timestamp: time.Now().Format("2006-01-02 15:04:05"),
				Msg:       "",
			}, pvcEvent.Pvc.Name, pvcEvent.Pvc.Namespace)
			if err != nil {
				smslog.WithContext(ctx).Infof("failed to update pvc process status to %s, ignore", PvcEventProcessing)
			}
			response, err := pvcHandler.handlePvcEvent(ctx, pvcEvent)
			if err != nil {
				smslog.WithContext(ctx).Errorf("pvc %s do action: %s  error: %s", pvcEvent.Pvc.GetName(), pvcEvent.Action, err.Error())
				if count := eventQueue.NumRequeues(pvcEvent); count < 2 {
					eventQueue.AddRateLimited(pvcEvent)
					smslog.WithContext(ctx).Errorf("error processing '%s': %s, less than maxRetry, requeuing %d times", pvcEvent.Pvc.GetName(), err.Error(), count)
				} else {
					smslog.WithContext(ctx).Errorf("error processing '%s': %s, greater equal maxRetry, do not try again, return", pvcEvent.Pvc.GetName(), err.Error())
					eventQueue.Forget(pvcEvent)
					utilruntime.HandleError(err)
				}
			} else {
				eventQueue.Forget(pvcEvent)
			}

			_ = UpdateConditionsByResponse(ctx, pvcEvent.Pvc, response, h.clientSet)

			err = h.updatePvcProcessResult(&PvcProcessResult{
				RequestId: pvcEvent.RequestID,
				Action:    pvcEvent.Action,
				Result:    response.reason,
				Timestamp: time.Now().Format("2006-01-02 15:04:05"),
				Msg:       response.message,
			},
				pvcEvent.Pvc.Name, pvcEvent.Pvc.Namespace)
			if err != nil {
				smslog.WithContext(ctx).Infof("failed to update pvc process status to %s, ignore", response.reason)
			}
			eventQueue.Done(pvcEvent)
		}
	}
}

func (h *PvcEventProcessor) updatePvcProcessResult(result *PvcProcessResult, name, namespace string) error {
	return UpdatePvcByAddAnnotation(name, namespace, PvcProcessStatus, result.String(), h.clientSet)
}

func UpdatePvcByAddAnnotation(name, namespace, key, value string, clientSet kubernetes.Interface) error {
	return common.RunWithRetry(3, 1*time.Second, func(retryTimes int) error {
		pvc, err := clientSet.CoreV1().PersistentVolumeClaims(namespace).Get(name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		pvc.Annotations[key] = value
		_, err = clientSet.CoreV1().PersistentVolumeClaims(pvc.Namespace).Update(pvc)
		return err
	})
}

func UpdateConditionsByResponse(ctx common.TraceContext, pvc *corev1.PersistentVolumeClaim, response *PvcResponse, clientSet kubernetes.Interface) error {
	pvc, err := clientSet.CoreV1().PersistentVolumeClaims(pvc.GetNamespace()).Get(pvc.GetName(), metav1.GetOptions{})
	if err != nil {
		smslog.WithContext(ctx).Infof("Failed get latest pvc: %v", err)
		return err
	}

	condition := corev1.PersistentVolumeClaimCondition{
		Type:               response.conditionType,
		Status:             response.status,
		Reason:             string(response.reason),
		LastProbeTime:      metav1.Time{},
		LastTransitionTime: metav1.Time{Time: time.Now()},
	}

	if response.pvcSwitchResult != nil {
		condition.Message = response.pvcSwitchResult.String(false)
	} else {
		condition.Message = response.message
	}

	var (
		conditions    []corev1.PersistentVolumeClaimCondition
		lastCondition *corev1.PersistentVolumeClaimCondition
	)

	pvcConditions := pvc.Status.Conditions
	for _, c := range pvcConditions {
		if c.Type == response.conditionType {
			lastCondition = &c
			continue
		}
		conditions = append(conditions, c)
	}

	if lastCondition != nil {
		smslog.WithContext(ctx).Infof("Start to update pvc condition %v from %v/%s to %s/%s",
			response.conditionType, lastCondition.Status, lastCondition.Reason, condition.Status, condition.Reason)
	}
	conditions = append(conditions, condition)

	if err := common.UpdatePvcConditions(pvc.GetName(), pvc.GetNamespace(), conditions, clientSet); err != nil {
		smslog.WithContext(ctx).Errorf("Failed update pvc conditions after 3 times try: %v", err)
		return err
	}

	return nil
}

//
//func FormatVolumeForBlock(nodeId string, nodeIp string, pvName string, logger log.Logger) error {
//	var (
//		addr   = fmt.Sprintf("%s:%d", nodeId, 22)
//		device = utils.GetDevicePath(pvName)
//		prInfo *common.PersistentReserve
//		err    error
//	)
//	// 格式化前检测残留的 mpathpersist pr key 存在并清理
//	if prInfo, err = common.GetScsiPrInfo(addr, device); err != nil {
//		return err
//	}
//	logger.Infof("Successfully find mpathpersist pr info [%+v] on %s", prInfo, nodeId)
//
//	if prInfo.ReservationKey != "" {
//		logger.Infof("The %s exist mpathpersist pr key [%+v], removing...", pvName, prInfo.Keys)
//
//		// pr key成员个数不能超过一个
//		if len(prInfo.Keys) > 1 {
//			err = fmt.Errorf("mpathpersist pr keys members %d is not equal to 1: %v", len(prInfo.Keys), prInfo.Keys)
//			return err
//		}
//
//		// 如果pr key不存在，要先注册, 再抢占，最后兜底清理
//		// 如果pr key存在，直接清理
//		var rk = utils.NodeToScsiPrKey(nodeIp)
//		if _, exist := prInfo.Keys[rk]; !exist {
//			if err = common.RegisterScsiPrKey(prInfo, addr, device, rk); err != nil {
//				return err
//			}
//			logger.Infof("Successfully register mpathpersist pr on %s with rk", nodeId)
//
//			if err = common.PreemptScsiPrReservationWithRk(addr, device, rk); err != nil {
//				return err
//			}
//			logger.Infof("Successfully preempt mpathpersist pr on %s with only rk", nodeId)
//		}
//
//		if err = common.RemoveScsiPrKey(addr, device, false); err != nil {
//			return err
//		}
//
//		logger.Infof("Successfully remove mpathpersist pr key on host %s", nodeId)
//	} else {
//		logger.Infof("Not found %s mpathpersist pr key, skip...", pvName)
//	}
//
//	if err = common.FormatVolumeForBlock(logger, pvName, nodeId); err != nil {
//		return err
//	}
//	logger.Infof("Successfully format volume on %s type %s", nodeId, corev1.PersistentVolumeBlock)
//	return nil
//}
//
//func FormatVolumeForExt4(pvName string, nodeId string, logger log.Logger) error {
//	if err := common.FormatVolumeForFilesystem(pvName, logger); err != nil {
//		err = fmt.Errorf("failed format volume on %s: %v", nodeId, err)
//		return err
//	}
//	logger.Infof("Successfully format volume on %s type %s", nodeId, corev1.PersistentVolumeFilesystem)
//	return nil
//}
