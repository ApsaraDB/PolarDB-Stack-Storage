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
	"polardb-sms/pkg/manager/domain/lv"
	"polardb-sms/pkg/manager/domain/workflow"
	"polardb-sms/pkg/manager/domain/workflow/stage"
)

type LvForOldLunService struct {
	lvRepo        lv.LvRepository
	volumeService *VolumeService
	clusterLunAsm assembler.ClusterLunAssembler
}

func NewLvForOldLunService() *LvForOldLunService {
	return &LvForOldLunService{
		lvRepo:        lv.GetLvRepository(),
		volumeService: NewVolumeService(),
		clusterLunAsm: assembler.NewClusterLunAssembler(),
	}
}

//todo finish this
func (s *LvForOldLunService) FormatAndLock(ctx common.TraceContext, v *view.ClusterLunFormatAndLockRequest) (*view.WorkflowIdResponse, error) {
	lunEntity, err := s.lvRepo.FindByVolumeId(v.Wwid)
	if err != nil {
		return nil, fmt.Errorf("can not find lun %s with wwid %v: %v", v.Name, v.Wwid, err)
	}
	if lunEntity == nil {
		return nil, fmt.Errorf("can not find the lun wwid %s", v.Wwid)
	}

	if lunEntity.FsType != common.NoFs && lunEntity.FsType != v.FsType {
		smslog.WithContext(ctx).Infof("PvcFormatAndLock: Warn lun with wwid %s exist fsType %s, but request to format to fsType %s", v.Name, lunEntity.FsType, v.FsType)
	}

	if lunEntity.FsType == v.FsType && lunEntity.FsSize == v.FsSize {
		smslog.WithContext(ctx).Infof("already formated for lun %s", v.Name)
	}

	options := map[string]interface{}{
		"fs_size": v.FsSize,
		"fs_type": v.FsType,
	}

	rwNodeConf := getNode(v.RwNodeIp)
	if rwNodeConf == nil {
		return nil, fmt.Errorf("can not find nodeIp %s from nodes %v", v.RwNodeIp, config.GetAvailableNodes())
	}

	if lunEntity.PrKey == rwNodeConf.Name {
		smslog.WithContext(ctx).Infof("wwid % disk already has write permission, but continue", v.Wwid)
	}
	wfl, err := s.genFormatAndLockWorkflow(lunEntity, rwNodeConf, options)
	if err != nil {
		return nil, err
	}
	wfl.SetTraceContext(ctx)
	if err = GetWorkflowEngine().Submit(wfl); err != nil {
		return nil, err
	}
	return &view.WorkflowIdResponse{
		WorkflowId: wfl.Id,
	}, nil
}

//todo finish this
func (s *LvForOldLunService) SetAccessPermission(ctx common.TraceContext, v *view.LunAccessPermissionRequest) (*view.WorkflowIdResponse, error) {
	lunEntity, err := s.lvRepo.FindByVolumeId(v.Wwid)
	if err != nil {
		return nil, err
	}
	if lunEntity == nil {
		return nil, fmt.Errorf("can not find the lun wwid %s", v.Wwid)
	}
	rwNodeConf := getNode(v.RwNodeIp)
	if rwNodeConf == nil {
		return nil, fmt.Errorf("can not find nodeIp %s from nodes %v", v.RwNodeIp, config.GetAvailableNodes())
	}

	if lunEntity.PrKey == rwNodeConf.Name {
		smslog.WithContext(ctx).Infof("wwid % disk already has write permission", v.Wwid)
		return nil, nil
	}
	wfl, err := s.genPrLockWorkflow(rwNodeConf, v.Wwid, lunEntity.PrKey)
	if err != nil {
		return nil, err
	}
	wfl.SetTraceContext(ctx)
	if err := GetWorkflowEngine().Submit(wfl); err != nil {

		return nil, err
	}
	return &view.WorkflowIdResponse{
		WorkflowId: wfl.Id,
	}, nil
}

func (s *LvForOldLunService) genPrLockWorkflow(prNode *config.Node, wwid string, existPrKey string) (*workflow.WorkflowEntity, error) {
	wb := workflow.NewWflBuilder().WithType(workflow.PrLock)
	prStage, err := stage.NewPrLockStage(*prNode, prNode.Ip, wwid, existPrKey, common.MultipathVolume)
	if err != nil {
		return nil, err
	}
	wb.WithStageRunner(prStage)
	return wb.Build(), nil
}

func (s *LvForOldLunService) genFormatAndLockWorkflow(lun *lv.LogicalVolumeEntity, prNode *config.Node, options map[string]interface{}) (*workflow.WorkflowEntity, error) {
	wb := workflow.NewWflBuilder().WithType(workflow.ClusterLunFormatAndLock)
	fsSize := options["fs_size"].(int64)
	fsType := options["fs_type"].(common.FsType)

	formatStageRunner := stage.NewFsFormatStage(lun.VolumeId, common.MultipathVolume, fsType, fsSize, prNode)
	wb.WithStageRunner(formatStageRunner)

	lockStageRunner, err := stage.NewPrLockStage(*prNode, prNode.Ip, lun.VolumeId, lun.PrKey, common.MultipathVolume)
	if err != nil {
		return nil, err
	}
	wb.WithStageRunner(lockStageRunner)

	lun.FsSize = fsSize
	lun.FsType = common.FsType(fsType)
	lvUpdateStageRunner, err := stage.NewDBPersistLvUpdateStage(lun)
	if err != nil {
		return nil, err
	}
	wb.WithStageRunner(lvUpdateStageRunner)

	return wb.Build(), nil
}

func getNode(nodeIp string) *config.Node {
	for _, nodeConf := range config.GetAvailableNodes() {
		if nodeConf.Ip == nodeIp {
			return &nodeConf
		}
	}
	return nil
}

func (s *LvForOldLunService) FindLunPrKey(wwid string) (string, error) {
	lunEntity, err := s.lvRepo.FindByVolumeId(wwid)
	if err != nil || lunEntity == nil {
		if err == nil {
			err = fmt.Errorf("can not find lun for %s", wwid)
		}
		smslog.Infof("can not find lun for %s", wwid)
		return "", err
	}
	if lunEntity.PrKey == "" {
		return "", nil
	}
	//todo fix this
	return lunEntity.PrKey, nil
}

func (s *LvForOldLunService) Create(ctx common.TraceContext, v *view.ClusterLunCreateRequest) (*view.WorkflowIdResponse, error) {
	lunEntity, err := s.lvRepo.FindByName(v.Wwid)
	if err == nil {
		return nil, fmt.Errorf("can not create lun with name %v, lun is already existd", v.Name)
	}

	lunEntity = s.clusterLunAsm.ToClusterLunEntity(v)
	wfl, err := s.genWorkflow(lunEntity, workflow.ClusterLunCreate)
	if err != nil {
		return nil, fmt.Errorf("can not create workflow for entity %v, err %v", lunEntity, err)
	}

	wfl.SetTraceContext(ctx)
	if err = GetWorkflowEngine().Submit(wfl); err != nil {
		return nil, err
	}
	return &view.WorkflowIdResponse{WorkflowId: wfl.Id}, nil
}

func (s *LvForOldLunService) Format(ctx common.TraceContext, v *view.ClusterLunFormatRequest) (*view.WorkflowIdResponse, error) {
	lunEntity, err := s.lvRepo.FindByVolumeId(v.Wwid)
	if err != nil {
		return nil, fmt.Errorf("can not find lun %s with wwid %v: %v", v.Name, v.Wwid, err)
	}

	if v.FsType == common.NoFs {
		v.FsType = common.Pfs
	}

	if lunEntity.FsType != common.NoFs && lunEntity.FsType != v.FsType {
		smslog.WithContext(ctx).Infof("Format: Warn lun with wwid %s exist fsType %s, but request to format to fsType %s", v.Name, lunEntity.FsType, v.FsType)
	}
	if lunEntity.FsType == v.FsType && lunEntity.FsSize == v.FsSize {
		smslog.WithContext(ctx).Infof("already formated for lun %s", v.Name)
	}

	lunEntity.SetFsType(v.FsType, v.FsSize)
	wfl, err := s.genWorkflow(lunEntity, workflow.ClusterLunFormat)
	if err != nil {
		return nil, fmt.Errorf("can not create workflow for entity %v, err %v", lunEntity, err)
	}

	wfl.SetTraceContext(ctx)
	if err = GetWorkflowEngine().Submit(wfl); err != nil {
		return nil, err
	}
	return &view.WorkflowIdResponse{WorkflowId: wfl.Id}, nil
}

func (s *LvForOldLunService) Expand(ctx common.TraceContext, v *view.ClusterLunCreateRequest) (*view.WorkflowIdResponse, error) {
	lunEntity, err := s.lvRepo.FindByVolumeId(v.Wwid)
	if err != nil {
		return nil, fmt.Errorf("can not find lun with name %v, lun is already existd", v.Name)
	}

	wfl, err := s.genWorkflow(lunEntity, workflow.ClusterLunExpand)
	if err != nil {
		return nil, fmt.Errorf("can not create workflow for entity %v, err %v", lunEntity, err)
	}

	wfl.SetTraceContext(ctx)
	if err = GetWorkflowEngine().Submit(wfl); err != nil {
		return nil, err
	}
	return &view.WorkflowIdResponse{WorkflowId: wfl.Id}, nil
}

func (s *LvForOldLunService) FsExpand(ctx common.TraceContext, v *view.ClusterLunFsExpandRequest) (*view.WorkflowIdResponse, error) {
	lunEntity, err := s.lvRepo.FindByVolumeId(v.Wwid)
	if err == nil {
		return nil, fmt.Errorf("can not create lun with name %v, lun is already existd", v.Wwid)
	}

	if lunEntity.FsType == common.NoFs {
		return nil, fmt.Errorf("fsexpand: can not find lv FsType with name %v", v.Name)
	}

	if lunEntity.FsType != common.NoFs && lunEntity.FsType != v.FsType {
		err = fmt.Errorf("warning lv with nam %s exist FsType %v, but request to format to FsType %v", lunEntity.VolumeId, lunEntity.FsType, v.FsType)
		smslog.Infof(err.Error())
		return nil, err
	}

	if lunEntity.FsSize == v.ExpandFsSize {
		return nil, fmt.Errorf("already expanded for lv %s", v.Wwid)
	}

	wfl, err := s.genWorkflow(lunEntity, workflow.ClusterLunFsExpand)
	if err != nil {
		return nil, fmt.Errorf("can not create workflow for entity %v, err %v", lunEntity, err)
	}

	wfl.SetTraceContext(ctx)
	if err = GetWorkflowEngine().Submit(wfl); err != nil {
		return nil, err
	}
	return &view.WorkflowIdResponse{WorkflowId: wfl.Id}, nil
}

func (s *LvForOldLunService) QueryAllMultipathVolumes() ([]*view.ClusterLunResponse, error) {
	multipathVolumes, err := s.lvRepo.QueryAllByType(common.MultipathVolume)
	if err != nil {
		return nil, err
	}
	responses := s.clusterLunAsm.ToClusterLunViews(multipathVolumes)
	return responses, nil
}

func (s *LvForOldLunService) LunsFromSameSan(req *view.ClusterLunSameSanRequest) (bool, error) {
	serialNumberMap := make(map[string]bool, 0)
	for _, volumeId := range req.VolumeIds {
		lvEntity, err := s.lvRepo.FindByVolumeId(volumeId)
		if err != nil {
			return false, err
		}
		serialNumberMap[lvEntity.SerialNumber] = true
	}
	cnt := 0
	for key, _ := range serialNumberMap {
		if key == "" {
			return false, nil
		}
		cnt++
	}
	return cnt == 1, nil
}

func (s *LvForOldLunService) genFsExpandWorkflow(lv *lv.LogicalVolumeEntity, wb *workflow.WflBuilder) error {
	prNodeIp := common.PrKeyToIpV4(lv.PrKey)
	var prNode *config.Node
	if prNodeIp == "" {
		prNode = config.GetOneNode()
	} else {
		prNode = config.GetNodeByIp(prNodeIp)
	}
	if prNode == nil {
		return fmt.Errorf("can not get prNode for lun %v", lv)
	}

	if err := s.volumeService.updateVolumeStatus(lv, domain.Expanding); err != nil {
		return err
	}

	fsExpandStageRunner := stage.NewFsExpandStage(lv.VolumeId, common.MultipathVolume, common.Pfs, lv.Size, lv.FsSize, prNode)
	wb.WithStageRunner(fsExpandStageRunner)

	lv.FsType = common.Pfs
	lv.FsSize = lv.Size
	lv.Status.StatusValue = domain.Success
	lvUpdateStageRunner, err := stage.NewDBPersistLvUpdateStage(lv)
	if err != nil {
		return err
	}
	wb.WithStageRunner(lvUpdateStageRunner)

	return nil
}

func (s *LvForOldLunService) genFormatWorkflow(lvEntity *lv.LogicalVolumeEntity, wb *workflow.WflBuilder) error {
	if err := s.volumeService.updateVolumeStatus(lvEntity, domain.Formatting); err != nil {
		return err
	}

	wrNode := lvEntity.GetCanWriteNode()
	if wrNode.Name == "" && wrNode.Ip == "" {
		return fmt.Errorf("get empty node from lvEntity %v", lvEntity)
	}
	formatStageRunner := stage.NewFsFormatStage(lvEntity.VolumeId, lvEntity.LvType, lvEntity.FsType, lvEntity.FsSize, &wrNode)
	wb.WithStageRunner(formatStageRunner)

	lvEntity.Status.StatusValue = domain.Success
	lvUpdateStageRunner, err := stage.NewDBPersistLvUpdateStage(lvEntity)
	if err != nil {
		return err
	}
	wb.WithStageRunner(lvUpdateStageRunner)

	return nil
}

func (s *LvForOldLunService) genWorkflow(lvEntity *lv.LogicalVolumeEntity, t workflow.WflType) (*workflow.WorkflowEntity, error) {
	var err error
	wb := workflow.NewWflBuilder().WithType(t)

	switch t {
	//case workflow.ClusterLunCreate:
	//	for nodeName, nodeConf := range config.GetAvailableNodes() {
	//		wb.WithSimpleStage(stage.NewStageBuilder().
	//			WithType(stage.ClusterLunCreateStage).
	//			WithTargetAgent(nodeName, nodeConf.Ip).
	//			Build())
	//	}
	//	metaPersistContent, err := stage.NewMetaDataPersistContent(stage.Create, "lv", e)
	//	if err != nil {
	//		return nil, err
	//	}
	//	wb.WithSimpleStage(stage.NewStageBuilder().
	//		WithType(stage.MetaPersistStage).
	//		WithContent(metaPersistContent).
	//		Build())
	//case workflow.ClusterLunExpand:
	//	for nodeName, nodeConf := range config.GetAvailableNodes() {
	//		wb.WithSimpleStage(stage.NewStageBuilder().
	//			WithType(stage.ClusterLunExpandStage).
	//			WithTargetAgent(nodeName, nodeConf.Ip).
	//			Build())
	//	}
	//	metaPersistContent, err := stage.NewMetaDataPersistContent(stage.Update, "lv", e)
	//	if err != nil {
	//		return nil, err
	//	}
	//	wb.WithSimpleStage(stage.NewStageBuilder().
	//		WithType(stage.MetaPersistStage).
	//		WithContent(metaPersistContent).
	//		Build())
	//
	case workflow.ClusterLunFsExpand:
		err = s.genFsExpandWorkflow(lvEntity, wb)
	case workflow.ClusterLunFormat:
		err = s.genFormatWorkflow(lvEntity, wb)
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
