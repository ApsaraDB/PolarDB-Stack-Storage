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
	"polardb-sms/pkg/manager/application/view"
	"polardb-sms/pkg/manager/config"
	"polardb-sms/pkg/manager/domain/lv"
	"polardb-sms/pkg/manager/domain/pv"
	"polardb-sms/pkg/manager/domain/workflow"
	"polardb-sms/pkg/manager/domain/workflow/stage"
	"polardb-sms/pkg/network/message"
)

type PrCheckService struct {
	lvRepo  lv.LvRepository
	pvRepo  pv.PhysicalVolumeRepository
	wflRepo workflow.WorkflowRepository
}

func NewPrCheckService() *PrCheckService {
	return &PrCheckService{
		lvRepo:  lv.GetLvRepository(),
		pvRepo:  pv.GetPhysicalVolumeRepository(),
		wflRepo: workflow.NewWorkflowRepository(),
	}
}

//同步调用，直接查数据库
func (s *PrCheckService) CheckOverall(v *view.PrCheckRequest) (*view.PrCheckOverallResponse, error) {
	switch v.VolumeClass {
	case common.LunClass:
		pvEntity, err := s.pvRepo.FindByVolumeIdAndNodeId(v.VolumeName, v.NodeId)
		if err != nil {
			return nil, err
		}
		return &view.PrCheckOverallResponse{
			PrInfo: pvEntity.PrSupport,
		}, nil
	case common.LvClass:
		lvEntity, err := s.lvRepo.FindByName(v.VolumeName)
		if err != nil {
			return nil, err
		}
		return &view.PrCheckOverallResponse{
			PrInfo: lvEntity.PrSupport,
		}, nil
	default:
		return nil, fmt.Errorf("unsupprt the volume type for %v", v)
	}
}

func (s *PrCheckService) CheckDetail(v *view.PrCheckRequest) (string, error) {
	lvEntity, err := s.lvRepo.FindByVolumeId(v.VolumeId)
	if err != nil || lvEntity == nil {
		if err == nil {
			err = fmt.Errorf("can not find lv by id %s", v.VolumeId)
			smslog.Errorf(err.Error())
			return "", err
		}
		smslog.Errorf("can not find lv by id %s err %s", v.VolumeId, err.Error())
		return "", err
	}

	var wb = workflow.NewWflBuilder().WithType(workflow.PrCheck)
	var nodeList = make([]config.Node, 0)
	//Check PathNum
	for _, node := range config.GetAvailableNodes() {
		nodeList = append(nodeList, node)
		stageRunner, err := stage.NewCheckPathNumCmdStage(node, lvEntity.VolumeId, lvEntity.LvType)
		if err != nil {
			return "", err
		}
		wb.WithStageRunner(stageRunner)
	}

	if len(nodeList)%2 != 0 {
		nodeList = append(nodeList, nodeList[0])
	}

	for idx, node := range nodeList[:len(nodeList)-1] {
		nextNode := nodeList[idx+1]
		stages, err := s.genPairCheckStages(lvEntity, node, nextNode)
		if err != nil {
			return "", err
		}
		wb.WithStageRunners(stages)
	}
	wfl := wb.Build()
	wfl.SetVolumeClass(string(v.VolumeClass))
	wfl.SetVolumeId(v.VolumeId)
	if err := GetWorkflowEngine().Submit(wfl); err != nil {
		return "", err
	}
	return wfl.Id, nil
}

func (s *PrCheckService) CheckDetailResponse(volumeId string, volumeClass string) (*view.PrCheckResponse, error) {
	wfl, err := s.wflRepo.FindByVolumeIdAndClass(volumeId, volumeClass, int(workflow.PrCheck))
	if err != nil {
		smslog.Errorf("CheckDetailResponse err %s", err.Error())
		return nil, err
	}
	if wfl.IsFinished() {
		if wfl.SuccessfullyRun() {
			//todo return pr check detail result
			lvEntity, err := s.lvRepo.FindByVolumeId(volumeId)
			if err != nil {
				smslog.Errorf("can not find lv entity %s  err %s", volumeId, err.Error())
				return nil, err
			}
			if lvEntity == nil {
				err = fmt.Errorf("can not find lv entity %s", volumeId)
				smslog.Error(err.Error())
				return nil, err
			}
			return &view.PrCheckResponse{
				PrInfo: lvEntity.PrInfo,
			}, nil
		}
		smslog.Errorf("workflow err %s", wfl.GetExecResult())
		return nil, fmt.Errorf("workflow run err %s", wfl.GetExecResult())
	}
	return nil, nil
}

//pair(left, right) pr check stage
func (s *PrCheckService) genPairCheckStages(v *lv.LogicalVolumeEntity, leftNode, rightNode config.Node) ([]workflow.StageRunner, error) {
	prReserveType := message.WEAR
	leftNodePrKey := common.IpV4ToPrKey(leftNode.Ip)
	rightNodePrKey := common.IpV4ToPrKey(rightNode.Ip)
	var stages []workflow.StageRunner

	leftRegAndReserveRunner, err := stage.NewRegisterAndReserveCmdStage(leftNode, v.VolumeId, v.LvType, leftNodePrKey, prReserveType)
	if err != nil {
		return nil, err
	}
	stages = append(stages, leftRegAndReserveRunner)

	rightRegAndPreemptRunner, err := stage.NewRegAndPreemptCmdStage(rightNode, v.VolumeId, v.LvType, rightNodePrKey, leftNodePrKey, prReserveType)
	if err != nil {
		return nil, err
	}
	stages = append(stages, rightRegAndPreemptRunner)

	leftPreemptAndClearRunner, err := stage.NewPreemptAndClearCmdStage(leftNode, v.VolumeId, v.LvType, leftNodePrKey, rightNodePrKey, prReserveType)
	if err != nil {
		return nil, err
	}
	stages = append(stages, leftPreemptAndClearRunner)

	rightRelAndClearRunner, err := stage.NewReleaseAndClearCmdStage(rightNode, v.VolumeId, v.LvType, rightNodePrKey, prReserveType)
	if err != nil {
		return nil, err
	}
	stages = append(stages, rightRelAndClearRunner)
	return stages, nil
}
