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

package service

import (
	"fmt"
	"polardb-sms/pkg/common"
	smslog "polardb-sms/pkg/log"
	"polardb-sms/pkg/manager/domain"
	"polardb-sms/pkg/manager/domain/k8spvc"
	"polardb-sms/pkg/manager/domain/lv"
	"polardb-sms/pkg/manager/domain/workflow"
	"polardb-sms/pkg/manager/domain/workflow/stage"
	"polardb-sms/pkg/network/message"
	"sync"
	"time"
)

//todo refactor this for too much biz logic
var _engine *WflEngine
var _engineOnce sync.Once

var wflProcessingMap = make(map[string]bool, 0)
var processingLock sync.Mutex

type WflProcessor struct {
	wflCh      chan *workflow.WorkflowEntity
	callbackCh chan interface{}
	wflRepo    workflow.WorkflowRepository
	stopChan   chan struct{}
}

func (p *WflProcessor) rollbackWorkflow(wfl *workflow.WorkflowEntity) {
	for {
		ret := p.Rollback(wfl)
		if !ret.IsSuccess() {
			smslog.Errorf("Error for rollback the stages: %v", ret)
			wfl.Status = workflow.FailRollback
			break
		}
		wfl.Step--
		if wfl.Step < 0 {
			smslog.Infof("workflow %s rollback finished", wfl.Id)
			wfl.Status = workflow.SuccessRollback
			break
		}
	}
}

func (p *WflProcessor) runWorkflow(wfl *workflow.WorkflowEntity) {
	for {
		ret := p.Run(wfl)
		if !ret.IsSuccess() {
			smslog.WithContext(wfl.TraceContext).Errorf("Error for run the stages: %v", ret)
			wfl.Status = workflow.Fail
			wfl.LastErrMsg = ret.ErrMsg
			break
		}
		wfl.Step++
		if wfl.Step >= len(wfl.Stages) {
			smslog.WithContext(wfl.TraceContext).Infof("workflow %s run finished", wfl.Id)
			wfl.Status = workflow.Success
			break
		}
	}
}

func (p *WflProcessor) run() {
	defer smslog.LogPanic()
	for {
		select {
		case <-p.stopChan:
			smslog.Infof("stop wflProcessor")
			return
		case wfl := <-p.wflCh:
			smslog.WithContext(wfl.TraceContext).Infof("start to process workflow %s", wfl.Id)
			if err := p.preProcess(wfl); err != nil {
				smslog.WithContext(wfl.TraceContext).Infof("preprocess workflow %s err %s", wfl.Id, err.Error())
				continue
			}
			processingLock.Lock()
			delete(wflProcessingMap, wfl.Id)
			processingLock.Unlock()
			switch wfl.Mode {
			case workflow.Run:
				p.runWorkflow(wfl)
			case workflow.Rollback:
				p.Rollback(wfl)
			default:
				continue
			}
			if err := p.postProcess(wfl); err != nil {
				smslog.WithContext(wfl.TraceContext).Infof("postprocess workflow %s err %s", wfl.Id, err.Error())
			}
		}
	}
}

func (p *WflProcessor) Save(w *workflow.WorkflowEntity) error {
	return common.RunWithRetry(3, 100*time.Millisecond, func(retryTimes int) error {
		_, err := p.wflRepo.Save(w)
		return err
	})
}

func (p *WflProcessor) Run(w *workflow.WorkflowEntity) *stage.StageExecResult {
	s := w.Stages[w.Step]
	ret := s.Run(w.TraceContext)
	return ret
}

func (p *WflProcessor) Rollback(w *workflow.WorkflowEntity) *stage.StageExecResult {
	s := w.Stages[w.Step]
	ret := s.Rollback(w.TraceContext)
	return ret
}

func (p *WflProcessor) preProcess(wfl *workflow.WorkflowEntity) error {
	if wfl.Status == workflow.NotStarted {
		wfl.Status = workflow.Started
		if err := p.Save(wfl); err != nil {
			return fmt.Errorf("could not update workflow %v on start status: %v", wfl, err)
		}
	}
	return nil
}

func (p *WflProcessor) postProcess(wfl *workflow.WorkflowEntity) error {
	if err := p.Save(wfl); err != nil {
		smslog.WithContext(wfl.TraceContext).Errorf("could not save workflow %v to database: %s", wfl.Id, err)
	}
	smslog.WithContext(wfl.TraceContext).Debugf("successfully save workflow %s", wfl.Id)

	//todo failed callback update entity status
	if wfl.Failed() {

	}

	switch wfl.WflType {
	case workflow.PrLock, workflow.PvcFormatAndLock, workflow.PvcFormat:
		return releasePvc(wfl)
	}
	return nil
}

type WflEngine struct {
	wflCh       chan *workflow.WorkflowEntity
	wflRepo     workflow.WorkflowRepository
	lvRepo      lv.LvRepository
	pvcRepo     k8spvc.PvcRepository
	pNum        int
	innerStopCh chan struct{}
}

func (e *WflEngine) Run() {
	defer smslog.LogPanic()
	e.innerStopCh = make(chan struct{})
	e.wflCh = make(chan *workflow.WorkflowEntity)
	//start pNum processors
	for i := 0; i < e.pNum; i++ {
		p := &WflProcessor{
			wflCh:      e.wflCh,
			callbackCh: make(chan interface{}),
			stopChan:   e.innerStopCh,
			wflRepo:    e.wflRepo,
		}
		go p.run()
	}
	//ticker to fetch job from db
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		<-ticker.C
		select {
		case <-e.innerStopCh:
			smslog.Info("stop wflEngine")
			return
		default:
			for _, w := range e.Pick() {
				if w.Valid() {
					e.wflCh <- w
				} else {
					w.Status = workflow.Fail
					w.LastErrMsg = "workflow overdue to invalid"
					_, err := e.wflRepo.Save(w)
					if err != nil {
						smslog.Debugf("to invalid wfl%s err %s", w.Id, err.Error())
					}
					processingLock.Lock()
					delete(wflProcessingMap, w.Id)
					processingLock.Unlock()
				}
			}
		}
	}
}

func (e *WflEngine) Stop() {
	smslog.Infof("Stop WflEngine")
	if e.innerStopCh != nil {
		close(e.innerStopCh)
		e.innerStopCh = nil
	}
	if e.wflCh != nil {
		close(e.wflCh)
		e.wflCh = nil
	}
}

func (e *WflEngine) Identify() string {
	return "workflowEngine"
}

func (e *WflEngine) Submit(w *workflow.WorkflowEntity) error {
	_, err := e.wflRepo.Create(w)
	if err != nil {
		return err
	}
	return nil
}

func (e *WflEngine) Pick() []*workflow.WorkflowEntity {
	for {
		processingLock.Lock()
		if len(wflProcessingMap) != 0 {
			processingLock.Unlock()
			time.Sleep(1 * time.Second)
			continue
		}
		processingLock.Unlock()
		break
	}
	wfls, err := e.wflRepo.FindByStatusAndLimit(int(workflow.NotStarted), 10)
	if err != nil {
		smslog.Infof("Pick error from DB %v", err)
		return nil
	}
	smslog.Debugf("pick workflow from db result %v", wfls)
	processingLock.Lock()
	defer processingLock.Unlock()
	for _, wfl := range wfls {
		wflProcessingMap[wfl.Id] = true
	}
	return wfls
}

func releasePvc(wfl *workflow.WorkflowEntity) error {
	pvcEntity, err := GetWorkflowEngine().pvcRepo.FindByLockedWorkflow(wfl.Id)
	if err != nil {
		return err
	}
	if pvcEntity == nil {
		return nil
	}
	if pvcEntity.GetVolumeId() != wfl.VolumeId {
		return fmt.Errorf("can not release lock the pvcModel in all namespace by wflId %s", wfl.Id)
	}
	return pvcEntity.UnLock()
}

//todo better way
func updateForPrResult(w *workflow.WorkflowEntity) error {
	lvEntity, err := GetWorkflowEngine().lvRepo.FindByVolumeId(w.VolumeId)
	if err != nil {
		smslog.Errorf("updateForPrResult FindByVolumeId err %s", err.Error())
		return err
	}
	return updateLvForPrResult(lvEntity, w.Stages)
}

func updateLvForPrResult(lvEntity *lv.LogicalVolumeEntity, stages []workflow.StageRunner) error {
	err := updatePrSupportStatus(&lvEntity.PrInfo, stages)
	if err != nil {
		smslog.Errorf("updateLvForPrResult updatePrSupportStatus err %s", err.Error())
		return err
	}
	_, err = GetWorkflowEngine().lvRepo.Save(lvEntity)
	if err != nil {
		smslog.Errorf("updateLvForPrResult save err %s", err.Error())
		return err
	}
	return nil
}

func updatePrSupportStatus(prInfo *lv.PrInfo, stages []workflow.StageRunner) error {
	for _, s := range stages {
		switch s.StageType() {
		case stage.PrStage:
			var prExecResult = &message.PrCheckCmdResult{}
			err := common.BytesToStruct(s.GetExecResult().Content, prExecResult)
			if err != nil {
				smslog.Errorf("BytesToStruct err %s for [%s] stage is [%v]", err.Error(), string(s.GetExecResult().Content), s.(*stage.PrStageRunner).Content)
				return err
			}
			//todo
			//updateCapabilityByCmdResult(prInfo.GetPrCheckListByKey(s.(*stage.PrStageRunner).TargetNode.VolumeId), prExecResult)
		case stage.PrBatchStage:
			var prBatchExecResult = &message.PrBatchCheckCmdResult{}
			err := common.BytesToStruct(s.GetExecResult().Content, prBatchExecResult)
			if err != nil {
				smslog.Errorf("updatePrSupportStatus err %s ", err.Error())
				return err
			}
			for _, prResult := range prBatchExecResult.Results {
				tgtNode := s.(*stage.PrBatchStageRunner).TargetNode
				checkList := prInfo.GetPrCheckListByKey(tgtNode.Name)
				updateCapabilityByCmdResult(checkList, prResult)
			}
		default:
			return fmt.Errorf("do not support get pr info from stage: %v", s)
		}
	}
	return nil
}

func updateCapabilityByCmdResult(prCheckList *lv.PrCheckList, result *message.PrCheckCmdResult) {
	switch result.CheckType {
	case message.PrRegister:
		prCheckList.SupportPrRegister = result.CheckResult == 0
	case message.PrReserve:
		prCheckList.SupportPrReserve = result.CheckResult == 0
	case message.PrRelease:
		prCheckList.SupportPrRelease = result.CheckResult == 0
	case message.PrClear:
		prCheckList.SupportPrClear = result.CheckResult == 0
	case message.PathCanWrite, message.PathCannotWrite:
		//special process skip
	case message.PrPreempt:
		prCheckList.SupportPrPreempt = result.CheckResult == 0
	case message.PrPathNum:
		//special process skip
	default:
		return
	}
}

//func updateCapabilityByBatchCmdResult(prCheckList *lv.PrCheckList, result *message.PrBatchCheckCmdResult, stageType stage.StageType, agentId string) {
//	for _, ret := range result.Results {
//		updateCapabilityByCmdResult(prCheckList, ret)
//	}
//	//do special process for
//	switch stageType {
//	case stage.PrBatchRegAndResStage:
//		batchResults := result.Results
//		pathCanWriteResult := batchResults[2]
//		if !(pathCanWriteResult.CheckResult == 0) {
//			prCheckList.SupportPrRegister = false
//			prCheckList.SupportPrReserve = false
//			prCheckList.SupportPr7 = false
//		} else {
//			prCheckList.SupportPr7 = true
//		}
//	case stage.PrBatchRegAndPreemptStage, stage.PrBatchPreemptAndClearStage:
//		batchResults := result.Results
//		pathCannotWriteResult := batchResults[0]
//		pathCanWriteResult := batchResults[3]
//		if !(pathCanWriteResult.CheckResult == 0 && pathCannotWriteResult.CheckResult == 0) {
//			smslog.Errorf("this test stage %d on agent %s test result %v is failed", stageType, agentId, result)
//			prCheckList.SupportPrRegister = false
//			prCheckList.SupportPrPreempt = false
//			prCheckList.SupportPr7 = false
//		} else {
//			prCheckList.SupportPr7 = true
//		}
//	default:
//		return
//	}
//}

func GetWorkflowEngine() *WflEngine {
	_engineOnce.Do(
		func() {
			_engine = &WflEngine{
				pNum:    5,
				wflRepo: workflow.NewWorkflowRepository(),
				pvcRepo: k8spvc.GetPvcRepository(),
				lvRepo:  lv.GetLvRepository(),
			}
		})
	return _engine
}

const (
	WorkflowExecTimeout                 = 1000 //10 second
	WorkflowCheckInterval time.Duration = 1    //1 second
)

type WorkflowService struct {
	wflRepo workflow.WorkflowRepository
}

func (s *WorkflowService) WaitUntilWorkflowFinish(workflowId string) error {
	if workflowId == domain.DummyWorkflowId {
		return nil
	}
	waitCnt := 0
	for {
		time.Sleep(WorkflowCheckInterval * time.Second)
		wfl, err := s.wflRepo.FindByWorkflowId(workflowId)
		if err != nil {
			continue
		}
		if wfl.IsFinished() {
			return workflowExecSuccess(wfl)
		}
		waitCnt += 1
		if waitCnt >= WorkflowExecTimeout {
			return fmt.Errorf("exec workflow %s timeout", workflowId)
		}
	}
}

func workflowExecSuccess(wfl *workflow.WorkflowEntity) error {
	if wfl.Failed() || wfl.SuccessfullyRollback() {
		return fmt.Errorf("workflow execute err %s", wfl.GetExecResult())
	}
	return nil
}

func NewWorkflowService() *WorkflowService {
	return &WorkflowService{
		wflRepo: workflow.NewWorkflowRepository(),
	}
}
