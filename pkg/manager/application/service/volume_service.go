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
	"polardb-sms/pkg/manager/application/view"
	"polardb-sms/pkg/manager/domain"
	"polardb-sms/pkg/manager/domain/lv"
	"polardb-sms/pkg/manager/domain/workflow"
	"polardb-sms/pkg/manager/domain/workflow/stage"
)

type VolumeService struct {
	lvRepo lv.LvRepository
}

func (s *VolumeService) Rescan() (*view.WorkflowIdResponse, error) {
	var wb = workflow.NewWflBuilder().WithType(workflow.Rescan)
	volumeRescanStageRunner := stage.NewPvRescanStage()
	wb.WithStageRunner(volumeRescanStageRunner)
	var wfl = wb.Build()
	if err := GetWorkflowEngine().Submit(wfl); err != nil {
		return nil, err
	}
	return &view.WorkflowIdResponse{WorkflowId: wfl.Id}, nil
}

func (s *VolumeService) updateVolumeStatus(lvEntity *lv.LogicalVolumeEntity, status domain.VolumeStatusValue) error {
	oldStatus := lvEntity.Status
	newStatus := domain.VolumeStatus{
		StatusValue: status,
	}
	if oldStatus == newStatus {
		return nil
	}
	newLvEntity := &lv.LogicalVolumeEntity{
		VolumeInfo: domain.VolumeInfo{
			VolumeId: lvEntity.VolumeId,
		},
		Status: newStatus,
	}
	if _, err := s.lvRepo.Save(newLvEntity); err != nil {
		return fmt.Errorf("failed update VolumeStatus[%v -> %v] lv with volumeId %s, err %v", oldStatus, newStatus, lvEntity.VolumeId, err)
	}
	return nil
}

func NewVolumeService() *VolumeService {
	return &VolumeService{
		lvRepo: lv.GetLvRepository(),
	}
}
