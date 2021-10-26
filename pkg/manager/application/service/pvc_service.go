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
	"polardb-sms/pkg/manager/application/assembler"
	"polardb-sms/pkg/manager/application/view"
	"polardb-sms/pkg/manager/config"
	"polardb-sms/pkg/manager/domain"
	"polardb-sms/pkg/manager/domain/k8spvc"
	"polardb-sms/pkg/manager/domain/lv"
	"polardb-sms/pkg/manager/domain/workflow"
	"polardb-sms/pkg/manager/domain/workflow/stage"
)

type PvcService struct {
	lvRepo     lv.LvRepository
	wflRepo    workflow.WorkflowRepository
	pvcRepo    k8spvc.PvcRepository
	pvcAsm     assembler.PvcAssembler
	lunService *LvForOldLunService
	lvService  *ClusterLvService
}

func (s *PvcService) PvcExpandFs(ctx common.TraceContext, request *view.PvcExpandFsRequest) (*view.WorkflowIdResponse, error) {
	pvcEntity, err := s.pvcRepo.FindByPvcName(request.Name, request.Namespace)
	if err != nil {
		return nil, err
	}
	reqSize := request.ReqSize
	fsType := request.FsType
	//todo fix
	if pvcEntity.DiskStatus.Size == reqSize {
		return &view.WorkflowIdResponse{
			WorkflowId: domain.DummyWorkflowId,
		}, nil
	}

	var wfl *workflow.WorkflowEntity
	lvEntity, err := s.lvRepo.FindByVolumeId(pvcEntity.ExpectedDiskStatus.VolumeId)
	if err != nil {
		return nil, err
	}
	if lvEntity == nil {
		err := fmt.Errorf("can not find lun for wwid %s", pvcEntity.ExpectedDiskStatus.VolumeId)
		smslog.WithContext(ctx).Error(err.Error())
		return nil, err
	}
	if lvEntity.Size < reqSize {
		return nil, fmt.Errorf("disk size %d less than request size %d", lvEntity.Size, reqSize)
	}
	if lvEntity.Size == 0 {
		return nil, fmt.Errorf("volumeId %s filesystem must not be empty, please format first", lvEntity.VolumeId)
	}

	lvEntity.SetFsType(fsType, reqSize)
	pvcEntity.SetRequestSize(reqSize)
	wfl, err = s.genWorkflow(pvcEntity, lvEntity, workflow.PvcFsExpand)
	if err != nil {
		return nil, err
	}

	wfl.SetTraceContext(ctx)
	if err := GetWorkflowEngine().Submit(wfl); err != nil {
		return nil, err
	}
	return &view.WorkflowIdResponse{WorkflowId: wfl.Id}, nil
}

func (s *PvcService) createPvc(pvcEntity *k8spvc.PersistVolumeClaimEntity, lvEntity *lv.LogicalVolumeEntity) error {
	//create pvc status is creating
	pvcEntity.PvcStatus = domain.VolumeStatus{
		StatusValue: domain.Creating,
	}
	pvcEntity.SetRequestSize(int64(lvEntity.SectorSize) * lvEntity.Sectors)
	if _, err := s.pvcRepo.Create(pvcEntity); err != nil {
		return err
	}
	return nil
}

func (s *PvcService) CreatePvcWithVolume(ctx common.TraceContext, pvcCreateView *view.PvcCreateWithVolumeRequest) (*view.WorkflowIdResponse, error) {
	pvcEntity, _ := s.pvcRepo.FindByPvcName(pvcCreateView.Name, pvcCreateView.Namespace)
	if pvcEntity != nil {
		smslog.Infof("pvcEntity already existed, skip")
		return &view.WorkflowIdResponse{WorkflowId: domain.DummyWorkflowId}, nil
	}
	pvcEntity = s.pvcAsm.ToPvcByCreateWithVolumeRequest(pvcCreateView)
	var wfl *workflow.WorkflowEntity
	lvEntity, err := s.lvRepo.FindByVolumeId(pvcEntity.ExpectedDiskStatus.VolumeId)
	if err != nil {
		return nil, err
	}
	if lvEntity == nil {
		err := fmt.Errorf("can not find lv for wwid %s", pvcEntity.ExpectedDiskStatus.VolumeId)
		smslog.Infof(err.Error())
		return nil, err
	}
	if !lvEntity.Usable() {
		return nil, fmt.Errorf("the lv device wwid: %s  status is useable %v", lvEntity.VolumeId, lvEntity.Status)
	}

	if err := s.createPvc(pvcEntity, lvEntity); err != nil {
		return nil, err
	}

	wfl, err = s.genWorkflow(pvcEntity, lvEntity, workflow.PvcCreate)
	if err != nil {
		return nil, err
	}
	wfl.SetTraceContext(ctx)
	if err := GetWorkflowEngine().Submit(wfl); err != nil {
		return nil, err
	}
	return &view.WorkflowIdResponse{WorkflowId: wfl.Id}, nil
}

func (s *PvcService) updatePvcStatus(pvcEntity *k8spvc.PersistVolumeClaimEntity, status domain.VolumeStatusValue) error {
	oldStatus := pvcEntity.PvcStatus
	newStatus := domain.VolumeStatus{
		StatusValue: status,
	}
	if oldStatus == newStatus {
		return nil
	}

	pvcEntity.PvcStatus = newStatus
	if _, err := s.pvcRepo.UpdateStatus(pvcEntity); err != nil {
		return fmt.Errorf("failed update VolumeStatus[%v -> %v] pvc with volumeId %s, err %v", oldStatus, newStatus, pvcEntity.DiskStatus.VolumeId, err)
	}
	return nil
}

func (s *PvcService) DeletePvc(ctx common.TraceContext, pvcRequest *view.PvcRequest) (*view.WorkflowIdResponse, error) {
	pvcEntity, err := s.pvcRepo.FindByPvcName(pvcRequest.Name, pvcRequest.Namespace)
	if err != nil {
		return nil, err
	}
	var wfl *workflow.WorkflowEntity

	lvEntity, err := s.lvRepo.FindByVolumeId(pvcEntity.DiskStatus.VolumeId)
	if err != nil || lvEntity == nil {
		if err == nil {
			err = fmt.Errorf("can not find logic volume %s", pvcEntity.DiskStatus.VolumeId)
		}
		smslog.Infof("DeletePvc err :%s ", err.Error())
	}
	if lvEntity == nil {
		lvEntity = &lv.LogicalVolumeEntity{
			VolumeInfo: domain.VolumeInfo{
				VolumeId: pvcEntity.DiskStatus.VolumeId,
			},
			LvType: pvcEntity.DiskStatus.VolumeType,
		}
	}

	wfl, err = s.genWorkflow(pvcEntity, lvEntity, workflow.PvcDelete)
	if err != nil {
		return nil, err
	}

	wfl.SetTraceContext(ctx)
	if err := GetWorkflowEngine().Submit(wfl); err != nil {
		return nil, err
	}
	return &view.WorkflowIdResponse{WorkflowId: wfl.Id}, nil
}

func (s *PvcService) ReleasePvc(ctx common.TraceContext, pvcRequest *view.PvcRequest) (*view.WorkflowIdResponse, error) {
	pvcEntity, err := s.pvcRepo.FindByPvcName(pvcRequest.Name, pvcRequest.Namespace)
	if err != nil {
		return nil, err
	}
	var wfl *workflow.WorkflowEntity

	lvEntity, err := s.lvRepo.FindByVolumeId(pvcEntity.DiskStatus.VolumeId)
	if err != nil || lvEntity == nil {
		if err == nil {
			err = fmt.Errorf("can not find logic volume %s", pvcEntity.DiskStatus.VolumeId)
		}
		smslog.Infof("find lv %s err %s, skip check and continue delete", pvcEntity.DiskStatus.VolumeId, err.Error())
	}

	wfl, err = s.genWorkflow(pvcEntity, lvEntity, workflow.PvcRelease)
	if err != nil {
		return nil, err
	}

	wfl.SetTraceContext(ctx)
	if err := GetWorkflowEngine().Submit(wfl); err != nil {
		return nil, err
	}
	return &view.WorkflowIdResponse{WorkflowId: wfl.Id}, nil
}

func (s *PvcService) ForceFormatAndLock(ctx common.TraceContext, request *view.PvcWriteLockRequest) (*view.WorkflowIdResponse, error) {
	pvcEntity, err := s.pvcRepo.FindByPvcName(request.Name, request.Namespace)
	if err != nil {
		return nil, err
	}

	lvEntity, err := s.lvRepo.FindByVolumeId(pvcEntity.DiskStatus.VolumeId)
	if err != nil {
		return nil, err
	}
	if lvEntity == nil {
		err := fmt.Errorf("can not find lv for wwid %s", pvcEntity.DiskStatus.VolumeId)
		smslog.Infof(err.Error())
		return nil, err
	}
	//if pvc is format and locking
	if pvcEntity.IsLocked() {
		lockWflId := pvcEntity.GetLockedWorkflow()
		lockWfl, err := s.wflRepo.FindByWorkflowId(lockWflId)
		if err != nil {
			smslog.WithContext(ctx).Errorf("ForceFormatAndLock err find by workflowId [%s]  err: %s", lockWflId, err.Error())
			return nil, err
		}
		if lockWfl.WflType == workflow.PvcFormatAndLock && !lockWfl.IsFinished() {
			return &view.WorkflowIdResponse{WorkflowId: lockWfl.Id}, nil
		}
	}
	volumeClass := string(pvcEntity.GetVolumeType().ToVolumeClass())
	wfl, err := s.wflRepo.FindByVolumeIdAndClass(pvcEntity.GetVolumeId(), volumeClass, int(workflow.PvcFormatAndLock))
	if err != nil {
		smslog.WithContext(ctx).Errorf("ForceFormatAndLock FindByVolumeIdAndClass class [%s] err %s", volumeClass, err.Error())
	}
	if wfl != nil && !wfl.IsFinished() {
		return &view.WorkflowIdResponse{WorkflowId: wfl.Id}, nil
	}

	var prNode *config.Node
	if request.WriteLockNodeIp != "" {
		prNode = config.GetNodeByIp(request.WriteLockNodeIp)
	}
	if prNode == nil {
		prNode = config.GetNodeById(request.WriteLockNodeId)
		if prNode == nil {
			return nil, fmt.Errorf("can not find NodeId %s in clusterConf", request.WriteLockNodeId)
		}
	}

	prKey := common.IpV4ToPrKey(prNode.Ip)
	pvcEntity.SetRequestPrKey(prKey)
	wfl, err = s.genWorkflow(pvcEntity, lvEntity, workflow.PvcFormatAndLock)
	if err != nil {
		return nil, err
	}

	wfl.SetVolumeClass(volumeClass)
	wfl.SetVolumeId(pvcEntity.GetVolumeId())

	if err := GetWorkflowEngine().Submit(wfl); err != nil {
		return nil, err
	}

	err = pvcEntity.Lock(wfl.Id)
	if err != nil {
		smslog.WithContext(ctx).Errorf("can not lock the pvc for %s", err.Error())
	}

	return &view.WorkflowIdResponse{
		WorkflowId: wfl.Id,
	}, nil
}

func (s *PvcService) PvcFsFormat(ctx common.TraceContext, formatRequest *view.PvcFormatRequest) (*view.WorkflowIdResponse, error) {
	pvcEntity, err := s.pvcRepo.FindByPvcName(formatRequest.Name, formatRequest.Namespace)
	if err != nil {
		smslog.WithContext(ctx).Errorf("can not find pvc %v err %s", formatRequest, err.Error())
		return nil, err
	}
	lvEntity, err := s.lvRepo.FindByVolumeId(pvcEntity.GetVolumeId())
	if err != nil {
		smslog.WithContext(ctx).Errorf("can not find the lv err %s", err.Error())
		return nil, err
	}

	if pvcEntity.IsLocked() {
		lockWflId := pvcEntity.GetLockedWorkflow()
		lockWfl, err := s.wflRepo.FindByWorkflowId(lockWflId)
		if err != nil {
			smslog.WithContext(ctx).Errorf("PvcFsFormat err find by workflowId [%s]  err: %s", lockWflId, err.Error())
			return nil, err
		}
		if lockWfl.WflType == workflow.PvcFormat && !lockWfl.IsFinished() {
			return &view.WorkflowIdResponse{WorkflowId: lockWfl.Id}, nil
		}
	}

	volumeClass := string(pvcEntity.GetVolumeType().ToVolumeClass())
	wfl, err := s.wflRepo.FindByVolumeIdAndClass(pvcEntity.GetVolumeId(), volumeClass, int(workflow.PvcFormat))
	if err != nil {
		smslog.WithContext(ctx).Errorf("PvcFsFormat FindByVolumeIdAndClass class [%s] err %s", volumeClass, err.Error())
	}
	if wfl != nil && !wfl.IsFinished() {
		return &view.WorkflowIdResponse{WorkflowId: wfl.Id}, nil
	}

	lvEntity.SetFsType(common.Pfs, lvEntity.Size)
	wfl, err = s.genWorkflow(pvcEntity, lvEntity, workflow.PvcFormat)
	if err != nil {
		return nil, err
	}

	err = pvcEntity.Lock(wfl.Id)
	if err != nil {
		return nil, err
	}

	wfl.SetTraceContext(ctx)
	if err := GetWorkflowEngine().Submit(wfl); err != nil {
		return nil, err
	}

	return &view.WorkflowIdResponse{WorkflowId: wfl.Id}, nil
}

func (s *PvcService) SetVolumeWriteLock(ctx common.TraceContext, lockRequest *view.PvcWriteLockRequest) (*view.WorkflowIdResponse, error) {
	pvcEntity, err := s.pvcRepo.FindByPvcName(lockRequest.Name, lockRequest.Namespace)
	if err != nil {
		return nil, err
	}
	node := config.GetNodeById(lockRequest.WriteLockNodeId)
	if node == nil {
		return nil, fmt.Errorf("can not find NodeId %s in clusterConf", lockRequest.WriteLockNodeId)
	}
	prKey := common.IpV4ToPrKey(node.Ip)
	pvcEntity.SetRequestPrKey(prKey)
	return s.setVolumeWriteLock(ctx, pvcEntity)
}

func (s *PvcService) setVolumeWriteLock(ctx common.TraceContext, pvcEntity *k8spvc.PersistVolumeClaimEntity) (*view.WorkflowIdResponse, error) {
	//已经锁住的pvc, 不允许新建workflow
	//TODO 优化，允许新加workflow, 需要增加一个queue对于同一个pvc
	if pvcEntity.IsLocked() {
		return nil, fmt.Errorf("pvc %s is locked by wfl %s", pvcEntity.Name, pvcEntity.GetLockedWorkflow())
	}
	lvEntity, err := s.lvRepo.FindByVolumeId(pvcEntity.DiskStatus.VolumeId)
	if err != nil {
		return nil, err
	}
	if lvEntity == nil {
		return nil, fmt.Errorf("can not find the lun for wwid: %s", pvcEntity.DiskStatus.VolumeId)
	}

	if lvEntity.PrKey != "" && lvEntity.PrKey == pvcEntity.ExpectedDiskStatus.PrKey {
		return &view.WorkflowIdResponse{WorkflowId: domain.DummyWorkflowId}, nil
	}

	wfl, err := s.genWorkflow(pvcEntity, lvEntity, workflow.PrLock)
	if err != nil {
		return nil, err
	}

	err = pvcEntity.Lock(wfl.Id)
	if err != nil {
		return nil, err
	}
	wfl.SetTraceContext(ctx)
	if err := GetWorkflowEngine().Submit(wfl); err != nil {
		return nil, err
	}
	return &view.WorkflowIdResponse{WorkflowId: wfl.Id}, nil
}

func (s *PvcService) PvcIsReady(pvcName, namespace, workflowId string) (*view.PvcIsReadyResponse, error) {
	if workflowId != domain.DummyWorkflowId {
		wfl, err := s.wflRepo.FindByWorkflowId(workflowId)
		if err != nil || wfl == nil {
			if err == nil {
				err = fmt.Errorf("can not find workflow %s", workflowId)
			}
			smslog.Errorf("can not find workflow %s", workflowId)
			return nil, err
		}
		if !wfl.IsFinished() {
			return nil, nil
		}
		if !wfl.SuccessfullyRun() {
			return nil, fmt.Errorf("workflow exec err %s ", wfl.GetExecResult())
		}
	}
	pvcEntity, err := s.pvcRepo.FindByPvcName(pvcName, namespace)
	if err != nil {
		return nil, err
	}
	isReady := pvcEntity.ReadyToUse()
	return &view.PvcIsReadyResponse{
		IsReady: isReady,
	}, nil
}

func (s *PvcService) QueryPvc(name, namespace string) (*view.PvcResponse, error) {
	pvc, err := s.pvcRepo.FindByPvcName(name, namespace)
	if err != nil || pvc == nil {
		if err == nil {
			err = fmt.Errorf("can not find pvc name [%s] namespace [%s]", name, namespace)
		}
		smslog.Error(err.Error())
		return nil, err
	}
	pvcResp := s.pvcAsm.ToPvcResponse(pvc)
	return pvcResp, nil
}

func (s *PvcService) QueryPvcs() ([]*view.PvcResponse, error) {
	pvcList, err := s.pvcRepo.QueryAll()
	if err != nil {
		return nil, err
	}
	pvcResps := s.pvcAsm.ToPvcResponses(pvcList)
	return pvcResps, nil
}

func (s *PvcService) QueryPvcsByType(volumeClass string) ([]*view.PvcResponse, error) {
	if volumeClass == "" {
		return s.QueryPvcs()
	}
	pvcList, err := s.pvcRepo.QueryByVolumeClass(volumeClass)
	if err != nil {
		return nil, err
	}
	pvcResps := s.pvcAsm.ToPvcResponses(pvcList)
	return pvcResps, nil
}

//todo refact this
func (s *PvcService) BindPvcAndVolumeLegacy(ctx common.TraceContext, bindRequest *view.PvcBindVolumeRequest) (*view.WorkflowIdResponse, error) {
	pvcEntity, err := s.pvcRepo.FindByPvcName(bindRequest.Name, bindRequest.Namespace)
	if err != nil {
		smslog.Errorf("can not bind pvc, for pvc %s is not exists, bind req %v, err is %s", bindRequest.Name, bindRequest, err.Error())
		return nil, fmt.Errorf("can not find pvc %s, err %s", bindRequest.Name, err.Error())
	}
	switch bindRequest.LvType {
	case common.MultipathVolume:
		return s.BindPvcAndLunLegacy(ctx, pvcEntity)
	default:
		return nil, fmt.Errorf("unsupport the idType %v", bindRequest.LvType)
	}
}

func (s *PvcService) UseAndFormatPvc(ctx common.TraceContext, bindRequest *view.PvcBindVolumeRequest) (*view.WorkflowIdResponse, error) {
	pvcEntity, err := s.pvcRepo.FindByPvcName(bindRequest.Name, bindRequest.Namespace)
	if err != nil {
		smslog.Warnf("can not bind pvc, for pvc %s is not exists, bind req %v, err is %s", bindRequest.Name, bindRequest, err.Error())
	}
	if pvcEntity == nil {
		return nil, fmt.Errorf("can not find pvc %s", bindRequest.Name)
	}

	if pvcEntity.DbClusterName != "" && pvcEntity.DbClusterName != bindRequest.ResourceId {
		return nil, fmt.Errorf("can not bind pvc, pvc %s exists used for other resourceId %s", bindRequest.Name, pvcEntity.DbClusterName)
	}

	err = pvcEntity.SetUsed(bindRequest.ResourceId)
	if err != nil {
		return nil, err
	}

	format := bindRequest.NeedFormat
	if !format {
		return &view.WorkflowIdResponse{
			WorkflowId: domain.DummyWorkflowId,
		}, nil
	}

	return s.PvcFsFormat(ctx, &view.PvcFormatRequest{
		PvcRequest: bindRequest.PvcRequest,
	})

}

func (s *PvcService) BindPvcAndLunLegacy(ctx common.TraceContext, pvcEntity *k8spvc.PersistVolumeClaimEntity) (*view.WorkflowIdResponse, error) {
	lunEntity, err := s.lvRepo.FindByVolumeId(pvcEntity.DiskStatus.VolumeId)
	if err != nil {
		return nil, err
	}
	if lunEntity == nil {
		err := fmt.Errorf("can not find lun for wwid %s", pvcEntity.DiskStatus.VolumeId)
		smslog.Error(err.Error())
		return nil, err
	}

	wfl, err := s.genWorkflow(pvcEntity, lunEntity, workflow.PvcBind)
	if err != nil {
		return nil, err
	}
	wfl.SetTraceContext(ctx)
	if err = GetWorkflowEngine().Submit(wfl); err != nil {
		return nil, err
	}
	return &view.WorkflowIdResponse{WorkflowId: wfl.Id}, nil
}

func (s *PvcService) QueryVolumePermissionTopo(name, namespace, workflowId string) (*view.PvcVolumePermissionTopoResponse, error) {
	if workflowId != domain.DummyWorkflowId {
		wfl, err := s.wflRepo.FindByWorkflowId(workflowId)
		if err != nil || wfl == nil {
			if err == nil {
				err = fmt.Errorf("can not find workflow %s", workflowId)
			}
			smslog.Errorf("can not find workflow %s", workflowId)
			return nil, err
		}
		if !wfl.IsFinished() {
			return nil, nil
		}
		if !wfl.SuccessfullyRun() {
			return nil, fmt.Errorf("workflow exec err %s ", wfl.GetExecResult())
		}
	}
	pvcEntity, err := s.pvcRepo.FindByPvcName(name, namespace)
	if err != nil {
		return nil, err
	}

	prKey := pvcEntity.GetPrKey()
	if prKey == "" {
		lun, err := s.lvRepo.FindByVolumeId(pvcEntity.DiskStatus.VolumeId)
		if err != nil {
			smslog.Errorf("QueryVolumePermissionTopo err %s", err.Error())
			return nil, err
		}
		prKey = lun.PrKey
	}
	return s.pvcAsm.ToPvcVolumePermTopo(prKey), nil
}

func (s *PvcService) genPvcBindWorkflow(pvcEntity *k8spvc.PersistVolumeClaimEntity, lvEntity *lv.LogicalVolumeEntity, wb *workflow.WflBuilder) error {
	format := pvcEntity.DiskStatus.NeedFormat
	var fsType = common.Pfs
	if pvcEntity.DiskStatus.VolumeMode == k8spvc.FsExt4 {
		fsType = common.Ext4
	}
	for _, nodeConf := range config.GetAvailableNodes() {
		pvcCreateStageRunner := stage.NewPvcCreateStage(pvcEntity.DiskStatus.VolumeId,
			common.MultipathVolume,
			fsType,
			format,
			pvcEntity.ExpectedDiskStatus.Size,
			nodeConf)
		wb.WithStageRunner(pvcCreateStageRunner)
		format = false
	}

	//update pvc
	err := pvcEntity.SetUsed(pvcEntity.DbClusterName)
	if err != nil {
		return err
	}

	//update pv
	pvEntity := pvcEntity.ConstructPvEntity()
	k8sPvCreateRunner, err := stage.NewDBPersistK8sPvCreateStage(pvEntity)
	if err != nil {
		return err
	}
	wb.WithStageRunner(k8sPvCreateRunner)

	//update lun/lv
	lvEntity.SetUsedBy(pvcEntity.Name, domain.DBUsed)
	lvUpdateRunner, err := stage.NewDBPersistLvUsedStage(lvEntity)
	if err != nil {
		return err
	}
	wb.WithStageRunner(lvUpdateRunner)
	return nil
}

func (s *PvcService) genPvcCreateWorkflow(pvcEntity *k8spvc.PersistVolumeClaimEntity, lvEntity *lv.LogicalVolumeEntity, wb *workflow.WflBuilder) error {
	format := pvcEntity.ExpectedDiskStatus.NeedFormat
	var fsType = common.Pfs
	if pvcEntity.ExpectedDiskStatus.VolumeMode == k8spvc.FsExt4 {
		fsType = common.Ext4
	}
	if format {
		lvEntity.SetFsType(fsType, pvcEntity.ExpectedDiskStatus.Size)
	}
	for _, nodeConf := range config.GetAvailableNodes() {
		pvcCreateStageRunner := stage.NewPvcCreateStage(lvEntity.VolumeId,
			lvEntity.LvType,
			fsType,
			format,
			pvcEntity.ExpectedDiskStatus.Size,
			nodeConf)
		format = false
		wb.WithStageRunner(pvcCreateStageRunner)
	}

	//create pv
	pvEntity := pvcEntity.ConstructPvEntity()
	pvPersistStageRunner, err := stage.NewDBPersistK8sPvCreateStage(pvEntity)
	if err != nil {
		return err
	}
	wb.WithStageRunner(pvPersistStageRunner)
	//update lun/lv
	lvEntity.SetUsedBy(pvcEntity.Name, domain.DBUsed)
	lvUpdateRunner, err := stage.NewDBPersistLvUpdateStage(lvEntity)
	if err != nil {
		return err
	}
	wb.WithStageRunner(lvUpdateRunner)
	//update pvc status
	pvcEntity.PvcStatus = domain.VolumeStatus{
		StatusValue: domain.Success,
	}
	pvcPersistStageRunner, err := stage.NewDBPersistPvcUpdateStatusStage(pvcEntity)
	if err != nil {
		return err
	}
	wb.WithStageRunner(pvcPersistStageRunner)

	return nil
}

func (s *PvcService) genPvcDeleteWorkflow(pvcEntity *k8spvc.PersistVolumeClaimEntity, lvEntity *lv.LogicalVolumeEntity, wb *workflow.WflBuilder) error {
	err := s.updatePvcStatus(pvcEntity, domain.Deleting)
	if err != nil {
		return err
	}
	clearPrLockStageRunners := s.lvService.getLvUnLockStageRunners(lvEntity, pvcEntity.PvName)
	wb.WithStageRunners(clearPrLockStageRunners)

	lvEntity.ReleaseUsed()
	updateLvStageRunner, err := stage.NewDBPersistLvUsedStage(lvEntity)
	if err != nil {
		return err
	}
	wb.WithStageRunner(updateLvStageRunner)

	//delete pvc
	deletePvcStageRunner, err := stage.NewDBPersistPvcDeleteStage(pvcEntity)
	if err != nil {
		return err
	}
	wb.WithStageRunner(deletePvcStageRunner)
	return nil
}

func (s *PvcService) genPvcReleaseWorkflow(pvcEntity *k8spvc.PersistVolumeClaimEntity, lvEntity *lv.LogicalVolumeEntity, wb *workflow.WflBuilder) error {
	err := s.updatePvcStatus(pvcEntity, domain.Releasing)
	if err != nil {
		return err
	}
	clearPrLockStageRunners := s.lvService.getLvUnLockStageRunners(lvEntity, pvcEntity.PvName)
	wb.WithStageRunners(clearPrLockStageRunners)

	pvcEntity.Release()

	pvcEntity.SetPrKey("")
	pvcLockStageRunners, err := stage.NewDBPersistPvcUpdatePrKeyStage(pvcEntity)
	if err != nil {
		return err
	}
	wb.WithStageRunner(pvcLockStageRunners)
	//update pvc status
	pvcEntity.PvcStatus = domain.VolumeStatus{
		StatusValue: domain.Success,
	}
	pvcPersistStageRunner, err := stage.NewDBPersistPvcUpdateStatusStage(pvcEntity)
	if err != nil {
		return err
	}
	wb.WithStageRunner(pvcPersistStageRunner)

	return nil
}

func (s *PvcService) genPvcLockWorkflow(pvcEntity *k8spvc.PersistVolumeClaimEntity, lvEntity *lv.LogicalVolumeEntity, wb *workflow.WflBuilder) error {
	prNodeIp := common.PrKeyToIpV4(pvcEntity.ExpectedDiskStatus.PrKey)
	prNode := config.GetNodeByIp(prNodeIp)
	if prNode == nil {
		err := fmt.Errorf("can not find the prNode for PrKey %s", pvcEntity.ExpectedDiskStatus.PrKey)
		smslog.Infof(err.Error())
		return err
	}
	var currentNodeIp string
	if lvEntity.PrKey != "" {
		currentNodeIp = common.PrKeyToIpV4(lvEntity.PrKey)
	}

	err := s.updatePvcStatus(pvcEntity, domain.PrLocking)
	if err != nil {
		return err
	}

	//lock lv
	prLockStageRunners, err := s.lvService.getLvLockStageRunners(lvEntity, *prNode, currentNodeIp)
	if err != nil {
		return err
	}
	wb.WithStageRunners(prLockStageRunners)

	//update lun
	lvEntity.PrKey = pvcEntity.ExpectedDiskStatus.PrKey
	lvUpdateStageRunner, err := stage.NewDBPersistLvUpdateStage(lvEntity)
	if err != nil {
		return err
	}
	wb.WithStageRunner(lvUpdateStageRunner)

	pvcEntity.SetPrKey(pvcEntity.ExpectedDiskStatus.PrKey)
	pvcLockStageRunners, err := stage.NewDBPersistPvcUpdatePrKeyStage(pvcEntity)
	if err != nil {
		return err
	}
	wb.WithStageRunner(pvcLockStageRunners)

	//update pvc status
	pvcEntity.PvcStatus = domain.VolumeStatus{
		StatusValue: domain.Success,
	}
	pvcPersistStageRunner, err := stage.NewDBPersistPvcUpdateStatusStage(pvcEntity)
	if err != nil {
		return err
	}
	wb.WithStageRunner(pvcPersistStageRunner)
	return nil
}

func (s *PvcService) genPvcExpandFsWorkflow(pvcEntity *k8spvc.PersistVolumeClaimEntity, lvEntity *lv.LogicalVolumeEntity, wb *workflow.WflBuilder) error {
	//expand fs
	//update lun fs size
	err := s.updatePvcStatus(pvcEntity, domain.Expanding)
	if err != nil {
		return err
	}

	prNodeIp := common.PrKeyToIpV4(pvcEntity.DiskStatus.PrKey)
	var prNode *config.Node
	if prNodeIp == "" {
		prNode = config.GetOneNode()
	} else {
		prNode = config.GetNodeByIp(prNodeIp)
	}
	if prNode == nil {
		return fmt.Errorf("can not get prNode for lun %v", lvEntity)
	}

	if pvcEntity.DiskStatus.PrKey != "" && lvEntity.LvType.ToVolumeClass() == common.LvClass {
		//append children lun pr
		prLockStageRunners, err := s.lvService.getLvLockStageRunners(lvEntity, *prNode, "")
		if err != nil {
			return err
		}
		wb.WithStageRunners(prLockStageRunners)
	}

	fsExpandStageRunner := stage.NewFsExpandStage(lvEntity.VolumeId, lvEntity.LvType, lvEntity.FsType, pvcEntity.ExpectedDiskStatus.Size, pvcEntity.DiskStatus.Size, prNode)
	wb.WithStageRunner(fsExpandStageRunner)

	lvEntity.FsSize = pvcEntity.ExpectedDiskStatus.Size
	lvUpdateStageRunner, err := stage.NewDBPersistLvUpdateStage(lvEntity)
	if err != nil {
		return err
	}
	wb.WithStageRunner(lvUpdateStageRunner)

	pvcUpdateStageRunner, err := stage.NewDBPersistPvcUpdateCapacityStage(pvcEntity)
	if err != nil {
		return err
	}
	wb.WithStageRunner(pvcUpdateStageRunner)

	//update pvc status
	pvcEntity.PvcStatus = domain.VolumeStatus{
		StatusValue: domain.Success,
	}
	pvcPersistStageRunner, err := stage.NewDBPersistPvcUpdateStatusStage(pvcEntity)
	if err != nil {
		return err
	}
	wb.WithStageRunner(pvcPersistStageRunner)
	return nil
}

func (s *PvcService) genPvcFormatAndLockWorkflow(pvcEntity *k8spvc.PersistVolumeClaimEntity, lvEntity *lv.LogicalVolumeEntity, wb *workflow.WflBuilder) error {
	lvEntity, err := s.lvRepo.FindByVolumeId(pvcEntity.GetVolumeId())
	if err != nil {
		smslog.Errorf("genFormatAndLockWorkflow: can not find lv [%s], err %s", pvcEntity.GetVolumeId(), err.Error())
		return err
	}

	if err := s.genPvcFormatWorkflow(pvcEntity, lvEntity, wb); err != nil {
		return err
	}

	if err := s.genPvcLockWorkflow(pvcEntity, lvEntity, wb); err != nil {
		return err
	}

	//update pvc status
	pvcEntity.PvcStatus = domain.VolumeStatus{
		StatusValue: domain.Success,
	}
	pvcPersistStageRunner, err := stage.NewDBPersistPvcUpdateStatusStage(pvcEntity)
	if err != nil {
		return err
	}
	wb.WithStageRunner(pvcPersistStageRunner)

	return nil
}

func (s *PvcService) genPvcFormatWorkflow(pvcEntity *k8spvc.PersistVolumeClaimEntity, lvEntity *lv.LogicalVolumeEntity, wb *workflow.WflBuilder) error {
	var (
		err error
	)
	volumeClass := pvcEntity.GetVolumeType().ToVolumeClass()

	err = s.updatePvcStatus(pvcEntity, domain.Formatting)
	if err != nil {
		return err
	}

	if volumeClass == common.LunClass {
		err = s.lunService.genFormatWorkflow(lvEntity, wb)
	}
	if volumeClass == common.LvClass {
		err = s.lvService.genFormatWorkflow(lvEntity, wb)
	}

	if err != nil {
		return fmt.Errorf("can not create workflow for entity %v, err %v", lvEntity, err)
	}

	//update pvc status
	pvcEntity.PvcStatus = domain.VolumeStatus{
		StatusValue: domain.Success,
	}
	pvcPersistStageRunner, err := stage.NewDBPersistPvcUpdateStatusStage(pvcEntity)
	if err != nil {
		return err
	}
	wb.WithStageRunner(pvcPersistStageRunner)

	return nil
}

func (s *PvcService) genWorkflow(pvcEntity *k8spvc.PersistVolumeClaimEntity, lvEntity *lv.LogicalVolumeEntity, t workflow.WflType) (*workflow.WorkflowEntity, error) {
	var err error
	wb := workflow.NewWflBuilder().WithType(t)
	switch t {
	case workflow.PvcCreate:
		err = s.genPvcCreateWorkflow(pvcEntity, lvEntity, wb)
	case workflow.PvcBind:
		err = s.genPvcBindWorkflow(pvcEntity, lvEntity, wb)
	case workflow.PvcFsExpand:
		err = s.genPvcExpandFsWorkflow(pvcEntity, lvEntity, wb)
	case workflow.PvcRelease:
		err = s.genPvcReleaseWorkflow(pvcEntity, lvEntity, wb)
	case workflow.PvcDelete:
		err = s.genPvcDeleteWorkflow(pvcEntity, lvEntity, wb)
	case workflow.PvcFormatAndLock:
		err = s.genPvcFormatAndLockWorkflow(pvcEntity, lvEntity, wb)
	case workflow.PvcFormat:
		err = s.genPvcFormatWorkflow(pvcEntity, lvEntity, wb)
	case workflow.PrLock:
		err = s.genPvcLockWorkflow(pvcEntity, lvEntity, wb)
	}

	if err != nil {
		return nil, err
	}
	wfl := wb.Build()
	wfl.SetVolumeId(lvEntity.VolumeId)
	volumeClass := lvEntity.LvType.ToVolumeClass()
	wfl.SetVolumeClass(string(volumeClass))

	return wfl, nil
}

func NewPvcService() *PvcService {
	return &PvcService{
		lvRepo:     lv.GetLvRepository(),
		wflRepo:    workflow.NewWorkflowRepository(),
		pvcRepo:    k8spvc.GetPvcRepository(),
		pvcAsm:     assembler.NewPvcAssembler(),
		lunService: NewLvForOldLunService(),
		lvService:  NewClusterLvService(),
	}
}
