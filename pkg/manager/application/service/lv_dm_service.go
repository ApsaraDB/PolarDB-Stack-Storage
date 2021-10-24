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

//TODO merge with lv multipath
type ClusterLvService struct {
	lvRepo        lv.LvRepository
	volumeService *VolumeService
	clusterLvAsm  assembler.ClusterLvAssembler
}

func NewClusterLvService() *ClusterLvService {
	return &ClusterLvService{
		lvRepo:        lv.GetLvRepository(),
		volumeService: NewVolumeService(),
		clusterLvAsm:  assembler.NewClusterLvAssembler(),
	}
}

func (s *ClusterLvService) QueryAllDmVolumes() ([]*view.ClusterLvResponse, error) {
	dmVolumes, err := s.lvRepo.QueryAllByTypes(common.DmLinearVolume, common.DmMirrorVolume, common.DmStripVolume)
	if err != nil {
		return nil, err
	}
	responses := s.clusterLvAsm.ToClusterLvViews(dmVolumes)
	return responses, nil
}

func (s *ClusterLvService) create(lv *lv.LogicalVolumeEntity) error {
	if err := lv.Valid(); err != nil {
		return err
	}
	if _, err := s.lvRepo.Create(lv); err != nil {
		return fmt.Errorf("failed creating lv with name %s, err %v", lv.VolumeName, err)
	}
	return nil
}

//TODO generate the dm tables
func (s *ClusterLvService) Create(ctx common.TraceContext, v *view.ClusterLvCreateRequest) (*view.WorkflowIdResponse, error) {
	var (
		err      error
		lvEntity *lv.LogicalVolumeEntity
	)
	//verify
	if lvEntity, err = s.lvRepo.FindByName(v.Name); err == nil && lvEntity != nil {
		return nil, fmt.Errorf("can not create lv with name %v, lv is already existd", v.Name)
	}

	lvEntity = s.clusterLvAsm.ToClusterLvEntity(v)
	if err = s.create(lvEntity); err != nil {
		return nil, err
	}

	wfl, err := s.genWorkflow(lvEntity, workflow.ClusterLvCreate)
	if err != nil {
		return nil, fmt.Errorf("can not create workflow for entity %v, err %v", lvEntity, err)
	}

	wfl.SetTraceContext(ctx)
	if err = GetWorkflowEngine().Submit(wfl); err != nil {
		return nil, err
	}

	return &view.WorkflowIdResponse{
		WorkflowId: wfl.Id,
	}, nil
}

func (s *ClusterLvService) Format(ctx common.TraceContext, v *view.ClusterLvFormatRequest) (*view.WorkflowIdResponse, error) {
	lvEntity, err := s.lvRepo.FindByVolumeId(v.VolumeId)
	if err != nil || lvEntity == nil {
		return nil, fmt.Errorf("format: can not find lv with name %v", v.VolumeId)
	}

	if lvEntity.FsType != common.NoFs && lvEntity.FsType != v.FsType {
		smslog.Infof("Warn lv with nam %s exist FsType %d, but request to format to FsType %d", v.VolumeId, lvEntity.FsType, v.FsType)
	}
	if lvEntity.FsType == v.FsType && lvEntity.FsSize == v.FsSize {
		smslog.Infof("already formated for lv %s", v.VolumeId)
	}

	lvEntity.SetFsType(v.FsType, v.FsSize)
	wfl, err := s.genWorkflow(lvEntity, workflow.ClusterLvFormat)
	if err != nil {
		return nil, fmt.Errorf("can not create workflow for entity %v, err %v", lvEntity, err)
	}

	wfl.SetTraceContext(ctx)
	if err = GetWorkflowEngine().Submit(wfl); err != nil {
		return nil, err
	}
	return &view.WorkflowIdResponse{
		WorkflowId: wfl.Id,
	}, nil
}

func (s *ClusterLvService) Expand(ctx common.TraceContext, v *view.ClusterLvExpandRequest) (*view.WorkflowIdResponse, error) {
	curLvEntity, err := s.lvRepo.FindByName(v.Name)
	if err != nil || curLvEntity == nil {
		return nil, fmt.Errorf("expand: can not find lv with name %v", v.Name)
	}

	lvEntity := s.clusterLvAsm.ToClusterLvEntity(v.ClusterLvCreateRequest)
	//todo more like lock
	if !needExpand(lvEntity, curLvEntity) {
		return nil, fmt.Errorf("already expanded, no need to do expand for %v", v)
	}

	wfl, err := s.genWorkflow(lvEntity, workflow.ClusterLvExpand)
	if err != nil {
		return nil, fmt.Errorf("can not create workflow for entity %v, err %v", lvEntity, err)
	}

	wfl.SetTraceContext(ctx)
	if err = GetWorkflowEngine().Submit(wfl); err != nil {
		return nil, err
	}
	return &view.WorkflowIdResponse{WorkflowId: wfl.Id}, nil
}

func (s *ClusterLvService) FsExpand(ctx common.TraceContext, v *view.ClusterLvFsExpandRequest) (*view.WorkflowIdResponse, error) {
	lvEntity, err := s.lvRepo.FindByVolumeId(v.VolumeId)
	if err != nil || lvEntity == nil {
		return nil, fmt.Errorf("fsexpand: can not find lv with name %v", v.VolumeId)
	}

	if lvEntity.FsType == common.NoFs {
		return nil, fmt.Errorf("fsexpand: can not find lv FsType with name %v", v.VolumeId)
	}

	if lvEntity.FsType != common.NoFs && lvEntity.FsType != v.FsType {
		err = fmt.Errorf("warning lv with nam %s exist FsType %v, but request to format to FsType %v", v.VolumeId, lvEntity.FsType, v.FsType)
		smslog.Infof(err.Error())
		return nil, err
	}
	if lvEntity.FsSize == v.ReqSize {
		return nil, fmt.Errorf("already expanded for lv %s", v.VolumeId)
	}

	lvEntity.FsSize = v.ReqSize
	lvEntity.FsType = v.FsType
	wfl, err := s.genWorkflow(lvEntity, workflow.ClusterLvFsExpand)
	if err != nil {
		return nil, fmt.Errorf("can not create workflow for entity %v, err %v", lvEntity, err)
	}

	wfl.SetTraceContext(ctx)
	if err = GetWorkflowEngine().Submit(wfl); err != nil {
		return nil, err
	}
	return &view.WorkflowIdResponse{WorkflowId: wfl.Id}, nil
}

func (s *ClusterLvService) Delete(ctx common.TraceContext, volumeId string) (*view.WorkflowIdResponse, error) {
	var (
		err error
		wfl *workflow.WorkflowEntity
	)

	lvEntity, err := s.lvRepo.FindByVolumeId(volumeId)
	if err != nil {
		return nil, err
	}
	if lvEntity == nil {
		smslog.Errorf("delete err: can not find the lv by volumeId %s", volumeId)
		return nil, fmt.Errorf("delete err: can not find the lv by volumeId %s", volumeId)
	}

	if wfl, err = s.genWorkflow(lvEntity, workflow.ClusterLvDelete); err != nil {
		return nil, fmt.Errorf("can not remove workflow for entity %v: %v", lvEntity, err)
	}

	wfl.SetTraceContext(ctx)
	if err = GetWorkflowEngine().Submit(wfl); err != nil {
		return nil, err
	}
	return &view.WorkflowIdResponse{WorkflowId: wfl.Id}, nil
}

func needExpand(tgt, src *lv.LogicalVolumeEntity) bool {
	if tgt == nil {
		return false
	}
	if src != nil &&
		(src.VolumeName == tgt.VolumeName &&
			src.Size == tgt.Size) {
		return false
	}
	return true
}

func (s *ClusterLvService) getLvUsedStageRunners(lvEntity *lv.LogicalVolumeEntity, byName string, byType domain.UsedByType) ([]workflow.StageRunner, error) {
	if lvEntity.LvType.ToVolumeClass() == common.LunClass {
		item, err := s.lvRepo.FindByVolumeId(lvEntity.GetVolumeId())
		if err != nil {
			return nil, fmt.Errorf("can not fine lun for %v: %v", lvEntity.GetVolumeId(), err)
		}
		if !item.Usable() {
			if item.UsedByType == byType && item.UsedByName == byName {
				return []workflow.StageRunner{}, nil
			}
			return nil, fmt.Errorf("lun %s disk already used type %v by used %s", lvEntity.GetVolumeId(), item.UsedByType, item.UsedByName)
		}
		lvEntity.SetUsedBy(byName, byType)
		stageRunner, err := stage.NewDBPersistLvUsedStage(lvEntity)
		if err != nil {
			return nil, err
		}
		return []workflow.StageRunner{stageRunner}, nil
	}
	runners := make([]workflow.StageRunner, 0)
	for _, child := range lvEntity.Children.Items {
		item, err := s.lvRepo.FindByVolumeId(child.GetVolumeId())
		if err != nil {
			return nil, fmt.Errorf("can not fine lun for %v: %v", child.GetVolumeId(), err)
		}
		if !item.Usable() {
			if item.UsedByType == byType && item.UsedByName == byName {
				continue
			}
			return nil, fmt.Errorf("lun %s disk already used type %v by used %s", child.GetVolumeId(), item.UsedByType, item.UsedByName)
		}
		item.SetUsedBy(byName, byType)
		stageRunner, err := stage.NewDBPersistLvUsedStage(item)
		if err != nil {
			return nil, err
		}
		runners = append(runners, stageRunner)
	}
	return runners, nil
}

func (s *ClusterLvService) getLvLockStageRunners(lvEntity *lv.LogicalVolumeEntity, prNode config.Node, currentIp string) ([]workflow.StageRunner, error) {
	runners := make([]workflow.StageRunner, 0)
	if lvEntity.LvType.ToVolumeClass() == common.LunClass {
		item, err := s.lvRepo.FindByVolumeId(lvEntity.GetVolumeId())
		if err != nil {
			return nil, fmt.Errorf("can not fine lun for %v: %v", lvEntity.GetVolumeId(), err)
		}
		currentPrIp := common.If(item.PrKey != "", common.PrKeyToIpV4(item.PrKey), "")
		if currentPrIp == prNode.Ip {
			smslog.Warnf("lun %s current pr %s already exist on node %s", item.PrKey, prNode.Ip)
			return []workflow.StageRunner{}, nil
		}
		stageRunner, err := stage.NewPrLockStage(prNode, prNode.Ip, lvEntity.VolumeId, currentIp, lvEntity.LvType)
		if err != nil {
			return nil, err
		}
		runners = append(runners, stageRunner)

		item.PrKey = common.IpV4ToPrKey(prNode.Ip)
		updatePrStageRunner, err := stage.NewDBPersistLvPrStage(item)
		if err != nil {
			return nil, err
		}
		runners = append(runners, updatePrStageRunner)
		return runners, nil
	}
	for _, child := range lvEntity.Children.Items {
		item, err := s.lvRepo.FindByVolumeId(child.GetVolumeId())
		if err != nil {
			return nil, fmt.Errorf("can not fine lun for %v: %v", child.GetVolumeId(), err)
		}
		currentPrIp := common.If(item.PrKey != "", common.PrKeyToIpV4(item.PrKey), "")
		if currentPrIp == prNode.Ip {
			smslog.Warnf("lun %s current pr %s already exist on node %s", item.PrKey, prNode.Ip)
			continue
		}
		stageRunner, err := stage.NewPrLockStage(prNode, prNode.Ip, child.GetVolumeId(), currentPrIp.(string), lvEntity.LvType)
		if err != nil {
			return nil, err
		}
		runners = append(runners, stageRunner)

		item.PrKey = common.IpV4ToPrKey(prNode.Ip)
		updatePrStageRunner, err := stage.NewDBPersistLvPrStage(item)
		if err != nil {
			return nil, err
		}
		runners = append(runners, updatePrStageRunner)
	}
	return runners, nil
}

func (s *ClusterLvService) getLvUnLockStageRunners(lvEntity *lv.LogicalVolumeEntity, pvName string) []workflow.StageRunner {
	runners := make([]workflow.StageRunner, 0)
	if lvEntity.LvType.ToVolumeClass() == common.LunClass {
		stageRunner := stage.NewPvcReleaseStage(lvEntity.VolumeId,
			lvEntity.LvType,
			pvName)
		runners = append(runners, stageRunner)

		lvEntity.ClearPrKey()
		dbPersistLvPRStageRunner, _ := stage.NewDBPersistLvPrStage(lvEntity)
		runners = append(runners, dbPersistLvPRStageRunner)

		return runners
	}

	for _, child := range lvEntity.Children.Items {
		stageRunner := stage.NewPvcReleaseStage(child.GetVolumeId(),
			common.MultipathVolume,
			pvName)
		runners = append(runners, stageRunner)

		if item, err := s.lvRepo.FindByVolumeId(child.GetVolumeId()); err == nil {
			item.ClearPrKey()
			dbPersistLvPRStageRunner, _ := stage.NewDBPersistLvPrStage(item)
			runners = append(runners, dbPersistLvPRStageRunner)
		}

		lvEntity.ClearPrKey()
		dbPersistLvPRStageRunner, _ := stage.NewDBPersistLvPrStage(lvEntity)
		runners = append(runners, dbPersistLvPRStageRunner)
	}
	return runners
}

func (s *ClusterLvService) getLvChildrenReleaseStageRunners(lvEntity *lv.LogicalVolumeEntity) []workflow.StageRunner {
	runners := make([]workflow.StageRunner, 0)
	for _, lun := range lvEntity.Children.Items {
		item, err := s.lvRepo.FindByVolumeId(lun.GetVolumeId())
		if err != nil {
			smslog.Errorf("can not fine lun for %v: %v", lun.GetVolumeId(), err)
			continue
		}
		item.ReleaseUsed()
		stageRunner, err := stage.NewDBPersistLvUsedStage(item)
		runners = append(runners, stageRunner)
	}
	return runners
}

func (s *ClusterLvService) genCreateWorkflow(lvEntity *lv.LogicalVolumeEntity, wb *workflow.WflBuilder) error {
	lvUsedStageRunners, err := s.getLvUsedStageRunners(lvEntity, lvEntity.GetVolumeName(), domain.LvUsed)
	if err != nil {
		return err
	}
	wb.WithStageRunners(lvUsedStageRunners)

	dmDeviceCore, err := lvEntity.GetDmDeviceCore()
	if err != nil {
		return err
	}
	lvCreateStageRunner := stage.NewLvCreateStage(dmDeviceCore)
	wb.WithStageRunner(lvCreateStageRunner)

	lvEntity.Status.StatusValue = domain.Success
	lvDBUpdateStageRunner, err := stage.NewDBPersistLvUpdateStage(lvEntity)
	if err != nil {
		return err
	}
	wb.WithStageRunner(lvDBUpdateStageRunner)

	return nil
}

func (s *ClusterLvService) genFormatWorkflow(lvEntity *lv.LogicalVolumeEntity, wb *workflow.WflBuilder) error {
	wrNode := lvEntity.GetCanWriteNode()
	if wrNode.Name == "" && wrNode.Ip == "" {
		return fmt.Errorf("get empty node from lvEntity %v", lvEntity)
	}
	if err := s.volumeService.updateVolumeStatus(lvEntity, domain.Formatting); err != nil {
		return err
	}
	stageRunner := stage.NewFsFormatStage(lvEntity.VolumeId,
		lvEntity.LvType,
		lvEntity.FsType,
		lvEntity.Size,
		&wrNode)
	wb.WithStageRunner(stageRunner)

	lvEntity.Status.StatusValue = domain.Success
	lvDBUpdateStageRunner, err := stage.NewDBPersistLvUpdateStage(lvEntity)
	if err != nil {
		return err
	}
	wb.WithStageRunner(lvDBUpdateStageRunner)

	return nil
}

func (s *ClusterLvService) genExpandWorkflow(lvEntity *lv.LogicalVolumeEntity, wb *workflow.WflBuilder) error {
	dmDeviceCore, err := lvEntity.GetDmDeviceCore()
	if err != nil {
		return err
	}
	if err := s.volumeService.updateVolumeStatus(lvEntity, domain.Expanding); err != nil {
		return err
	}

	lvUsedStageRunners, err := s.getLvUsedStageRunners(lvEntity, lvEntity.GetVolumeName(), domain.LvUsed)
	if err != nil {
		return err
	}
	wb.WithStageRunners(lvUsedStageRunners)

	stageRunner := stage.NewLvExpandStage(dmDeviceCore)
	wb.WithStageRunner(stageRunner)

	lvEntity.Status.StatusValue = domain.Success
	lvDBUpdateStageRunner, err := stage.NewDBPersistLvUpdateStage(lvEntity)
	if err != nil {
		return err
	}
	wb.WithStageRunner(lvDBUpdateStageRunner)

	return nil
}

func (s *ClusterLvService) genFsExpandWorkflow(lvEntity *lv.LogicalVolumeEntity, wb *workflow.WflBuilder) error {
	wrNode := lvEntity.GetCanWriteNode()
	if wrNode.Name == "" && wrNode.Ip == "" {
		return fmt.Errorf("get empty node from lvEntity %v", lvEntity)
	}

	if err := s.volumeService.updateVolumeStatus(lvEntity, domain.Expanding); err != nil {
		return err
	}

	stageRunner := stage.NewFsExpandStage(lvEntity.VolumeId, lvEntity.LvType, common.Pfs, lvEntity.Size, lvEntity.FsSize, &wrNode)
	wb.WithStageRunner(stageRunner)

	lvEntity.Status.StatusValue = domain.Success
	lvDBUpdateStageRunner, err := stage.NewDBPersistLvUpdateStage(lvEntity)
	if err != nil {
		return err
	}
	wb.WithStageRunner(lvDBUpdateStageRunner)

	return nil
}

func (s *ClusterLvService) genDeleteWorkflow(lvEntity *lv.LogicalVolumeEntity, wb *workflow.WflBuilder) error {
	if err := s.volumeService.updateVolumeStatus(lvEntity, domain.Deleting); err != nil {
		return err
	}

	dmDeviceCore, err := lvEntity.GetDmDeviceCore()
	if err != nil {
		return err
	}
	stageRunner := stage.NewLvDeleteStage(dmDeviceCore)
	wb.WithStageRunner(stageRunner)

	childrenReleaseStageRunners := s.getLvChildrenReleaseStageRunners(lvEntity)
	wb.WithStageRunners(childrenReleaseStageRunners)

	lvDeleteStageRunner, err := stage.NewDBPersistLvDeleteStage(lvEntity)
	if err != nil {
		return err
	}
	wb.WithStageRunner(lvDeleteStageRunner)

	return nil
}

func (s *ClusterLvService) genWorkflow(lvEntity *lv.LogicalVolumeEntity, t workflow.WflType) (*workflow.WorkflowEntity, error) {
	var err error
	wb := workflow.NewWflBuilder().WithType(t)
	switch t {
	case workflow.ClusterLvCreate:
		err = s.genCreateWorkflow(lvEntity, wb)

	case workflow.ClusterLvFormat:
		err = s.genFormatWorkflow(lvEntity, wb)

	case workflow.ClusterLvExpand:
		err = s.genExpandWorkflow(lvEntity, wb)

	case workflow.ClusterLvFsExpand:
		err = s.genFsExpandWorkflow(lvEntity, wb)

	case workflow.ClusterLvDelete:
		err = s.genDeleteWorkflow(lvEntity, wb)
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
