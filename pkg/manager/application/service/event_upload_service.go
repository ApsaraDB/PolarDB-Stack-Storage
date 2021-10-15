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
	"polardb-sms/pkg/manager/config"
	"polardb-sms/pkg/manager/domain"
	"polardb-sms/pkg/manager/domain/lv"
	"polardb-sms/pkg/manager/domain/pv"
	"polardb-sms/pkg/protocol"
	"strings"
)

type EventUploadService struct {
	handlers map[protocol.EventType]EventHandler
	pvRepo   pv.PhysicalVolumeRepository
	lvRepo   lv.LvRepository
	asm      assembler.PvAssemblerForEvent
}

type EventHandler func(body string) error

func NewEventUploadService() *EventUploadService {
	es := &EventUploadService{
		handlers: make(map[protocol.EventType]EventHandler),
		pvRepo:   pv.GetPhysicalVolumeRepository(),
		lvRepo:   lv.GetLvRepository(),
		asm:      assembler.NewHostLunFullInfoAssemblerForEvent(),
	}
	es.register(protocol.LunAdd, es.HandleLunAddEvent)
	es.register(protocol.LunUpdate, es.HandleLunUpdateEvent)
	es.register(protocol.LunRemove, es.HandleLunRemoveEvent)
	es.register(protocol.LvAdd, es.HandleLvAddEvent)
	es.register(protocol.LvUpdate, es.HandleLvUpdateEvent)
	es.register(protocol.LvRemove, es.HandleLvRemoveEvent)
	return es
}

func (s *EventUploadService) UploadEvents(events []*protocol.Event) error {
	for _, e := range events {
		if err := s.UploadEvent(e); err != nil {
			smslog.Errorf("err when upload event %v :%s, ", e, err.Error())
		}
	}
	return nil
}

func (s *EventUploadService) UploadEvent(e *protocol.Event) error {
	handlerFunc, ok := s.handlers[e.EventType]
	if !ok {
		return fmt.Errorf("can not find eventHandler for event: %v", e)
	}
	return handlerFunc(e.Body)
}

func (s *EventUploadService) register(t protocol.EventType, hFunc EventHandler) {
	s.handlers[t] = hFunc
}

func (s *EventUploadService) HandleLunAddEvent(e string) error {
	event := protocol.LunAddEvent{}
	if err := protocol.Decode(e, &event); err != nil {
		smslog.Errorf("LunAddEvent: could not decode event %s: %v", e, err)
		return err
	}
	if event.VolumeId == "" || event.Name == "" {
		smslog.Debugf("LunAddEvent: ignore to process for id or name is nil: id: [%s] name [%s]", event.VolumeId, event.Name)
		return nil
	} else {
		smslog.Debugf("LunAddEvent: start to process: id: %s name %s pathNum %d fs_type [%s]", event.VolumeId, event.Name, event.PathNum, event.FsType)
	}

	lunEntity, err := s.pvRepo.FindByVolumeIdAndNodeId(event.VolumeId, event.NodeId)
	if err != nil {
		smslog.Errorf("find lun by id %s err %s", event.VolumeId, err.Error())
		return err
	}
	if lunEntity == nil {
		lunEntity, err = s.createLunByEvent(&event)
		if err != nil {
			smslog.Errorf("create lun by event %s err %s", event.VolumeId, err.Error())
			return err
		}
	} else {
		if err := s.updateLunByEvent(lunEntity, &event); err != nil {
			smslog.Errorf("update Lun %s By event $v err %s", lunEntity.GetVolumeId(), event, err.Error())
			return err
		}
	}

	lvEntity, err := s.lvRepo.FindByVolumeId(event.VolumeId)
	if err != nil {
		smslog.Errorf(err.Error())
		return err
	}
	if lvEntity == nil {
		if err := s.createLvByLun(lunEntity); err != nil {
			smslog.Errorf("create lv by lun %s err %s", lunEntity.VolumeId, err.Error())
			return err
		}
	} else {
		return s.updateLvByEvent(lvEntity, &event)
	}
	return nil
}

func (s *EventUploadService) createLunByEvent(event *protocol.LunAddEvent) (*pv.PhysicalVolumeEntity, error) {
	lunEntity := s.asm.ToClusterLunEntityByEvent(&event.Lun)
	_, err := s.pvRepo.Create(lunEntity)
	if err != nil {
		smslog.Errorf("Creat lun err for id %s err %s", event.VolumeId, err.Error())
		return nil, err
	}
	return lunEntity, nil
}

func (s *EventUploadService) createLvByLun(pvEntity *pv.PhysicalVolumeEntity) error {
	lvEntity := &lv.LogicalVolumeEntity{
		VolumeInfo: pvEntity.VolumeInfo,
		LvType:     common.MultipathVolume,
		ClusterId:  0,
		PrKey:      pvEntity.PrSupport.PrKey,
		Children:   &lv.Children{},
		PrInfo:     make(map[string]*lv.PrCheckList),
		Extend:     make(map[string]interface{}, 0),
		NodeIds:    []string{pvEntity.Hostname},
		Status:     *domain.NonStatus,
	}
	lvEntity.Children.AddChild(pvEntity)
	_, err := s.lvRepo.Create(lvEntity)
	if err != nil {
		smslog.Errorf("create lv err for id %s err %s", pvEntity.GetVolumeId(), err.Error())
		return err
	}
	return nil
}

func (s *EventUploadService) updateLunByEvent(lunEntity *pv.PhysicalVolumeEntity, event *protocol.LunAddEvent) error {
	var reportDiskSize = event.Size
	if lunEntity.Size < reportDiskSize {
		smslog.Infof("VolumeInfo %s is expanded from %d to %d", event.VolumeId, lunEntity.Size, event.Size)
		lunEntity.Size = reportDiskSize
	} else if lunEntity.Size > reportDiskSize {
		smslog.Errorf("can not support shrink the disk %s size from %d to %d, but do not return error", event.VolumeId, lunEntity.Size, event.Size)
	}
	lunEntity.PrSupport = event.PrSupport
	lunEntity.FsType = event.FsType
	lunEntity.FsSize = event.FsSize
	lunEntity.UsedSize = event.UsedSize
	lunEntity.PathNum = event.PathNum
	lunEntity.Paths = event.Paths
	_, err := s.pvRepo.Save(lunEntity)
	if err != nil {
		smslog.Errorf("update lun %s err %s", lunEntity.GetVolumeId(), err.Error())
		return err
	}
	return nil
}

func (s *EventUploadService) updateLvByEvent(lvEntity *lv.LogicalVolumeEntity, event *protocol.LunAddEvent) error {
	if lvEntity.VolumeName != event.Name && strings.HasPrefix(event.Name, "pv-") {
		lvEntity.VolumeName = event.Name
	}

	lvEntity.PrSupport = event.PrSupport
	if lvEntity.PrKey != event.PrSupport.PrKey && event.PrSupport.PrKey != "" {
		lvEntity.PrKey = event.PrSupport.PrKey
	}
	var reportDiskSize = event.Size
	if lvEntity.Size < reportDiskSize {
		smslog.Infof("VolumeInfo %s is expanded from %d to %d", event.VolumeId, lvEntity.Size, event.Size)
		lvEntity.Size = reportDiskSize
	} else if lvEntity.Size > reportDiskSize {
		smslog.Errorf("can not support shrink the disk %s size from %d to %d, but do not return error",
			event.VolumeId, lvEntity.Size, event.Size)
	}
	lvEntity.FsType = event.FsType
	lvEntity.FsSize = event.FsSize
	lvEntity.UsedSize = event.UsedSize
	lvEntity.AddNodeId(event.NodeId)
	lvEntity.AddChildByTypeAndId(common.Pv, event.VolumeId, event.NodeId)
	if len(lvEntity.NodeIds) >= (len(config.GetAvailableNodes()) - 1) && lvEntity.Status.StatusValue == domain.NoAction {
		lvEntity.Status.StatusValue = domain.Success
	}

	_, err := s.lvRepo.Save(lvEntity)
	if err != nil {
		smslog.Errorf("update lv %s err %s", lvEntity.VolumeId, err.Error())
		return err
	}
	return nil
}

func (s *EventUploadService) HandleLunUpdateEvent(e string) error {
	return nil
}

func (s *EventUploadService) HandleLunRemoveEvent(e string) error {
	return nil
}

func (s *EventUploadService) createLvByEvent(event *protocol.LvAddEvent) error {
	var lvType common.LvType
	if event.VolumeType == string(common.DmStripVolume) {
		lvType = common.DmStripVolume
	} else {
		lvType = common.DmLinearVolume
	}

	lvEntity := &lv.LogicalVolumeEntity{
		VolumeInfo: domain.VolumeInfo{
			VolumeName: strings.TrimPrefix(event.VolumeId, common.DmNamePrefix),
			VolumeId:   event.VolumeId,
			Size:       event.Size,
			Sectors:    event.Sectors,
			SectorSize: event.SectorSize,
			FsType:     event.FsType,
			FsSize:     event.FsSize,
			UsedSize:   event.UsedSize,
		},
		LvType:    lvType,
		ClusterId: 0,
		//todo prkey
		//PrKey:     pvEntity.PrSupport.PrKey,
		Children: &lv.Children{},
		PrInfo:   make(map[string]*lv.PrCheckList),
		Extend:   make(map[string]interface{}, 0),
		NodeIds:  []string{event.NodeId},
	}
	for _, child := range event.Children {
		childLv, err := s.lvRepo.FindByVolumeId(child)
		if err != nil {
			smslog.Errorf("createLvByEvent lv [%s] children [%s] err [%s]", event.VolumeId, event.Children, err.Error())
			return err
		}
		lvEntity.Children.AddChild(childLv)
	}
	_, err := s.lvRepo.Create(lvEntity)
	if err != nil {
		smslog.Errorf("create lv err for id %s err %s", event.VolumeId, err.Error())
		return err
	}
	return nil
}

func (s *EventUploadService) updateLvByLvAddEvent(lvEntity *lv.LogicalVolumeEntity, event *protocol.LvAddEvent) error {
	lvEntity.PrSupport = event.PrSupport
	if lvEntity.PrKey != event.PrSupport.PrKey && event.PrSupport.PrKey != "" {
		lvEntity.PrKey = event.PrSupport.PrKey
	}
	var reportDiskSize = event.Size
	if lvEntity.Size < reportDiskSize {
		smslog.Infof("VolumeInfo %s is expanded from %d to %d", event.VolumeId, lvEntity.Size, event.Size)
		lvEntity.Size = reportDiskSize
	} else if lvEntity.Size > reportDiskSize {
		smslog.Errorf("can not support shrink the disk %s size from %d to %d, but do not return error",
			event.VolumeId, lvEntity.Size, event.Size)
	}
	lvEntity.FsType = event.FsType
	lvEntity.FsSize = event.FsSize
	lvEntity.UsedSize = event.UsedSize
	lvEntity.AddNodeId(event.NodeId)
	if len(lvEntity.NodeIds) >= (len(config.GetAvailableNodes()) - 1) {
		lvEntity.Status.StatusValue = domain.Success
	}

	_, err := s.lvRepo.Save(lvEntity)
	if err != nil {
		smslog.Errorf("update lv %s err %s", lvEntity.VolumeId, err.Error())
		return err
	}
	return nil
}

func (s *EventUploadService) HandleLvAddEvent(e string) error {
	event := protocol.LvAddEvent{}
	if err := protocol.Decode(e, &event); err != nil {
		smslog.Errorf("LvAddEvent: could not decode event %s: %v", e, err)
		return err
	}
	//TODO fix me
	smslog.Infof("LvAddEvent: start to process: %v", event)
	if event.VolumeId == "" {
		smslog.Debugf("LunAddEvent: ignore to process for id is nil: id: [%s] ", event.VolumeId)
		return nil
	} else {
		smslog.Debugf("LunAddEvent: start to process: id: %s fs_type [%s]", event.VolumeId, event.FsType)
	}

	lvEntity, err := s.lvRepo.FindByVolumeId(event.VolumeId)
	if err != nil {
		smslog.Errorf("find lun by id %s err %s", event.VolumeId, err.Error())
		return err
	}

	if lvEntity == nil {
		err = s.createLvByEvent(&event)
		if err != nil {
			smslog.Errorf("create lun by event %s err %s", event.VolumeId, err.Error())
			return err
		}
	} else {
		if err := s.updateLvByLvAddEvent(lvEntity, &event); err != nil {
			smslog.Errorf("update Lun %s By event $v err %s", lvEntity.GetVolumeId(), event, err.Error())
			return err
		}
	}

	return nil
}

func (s *EventUploadService) HandleLvUpdateEvent(e string) error {
	return nil
}

func (s *EventUploadService) HandleLvRemoveEvent(e string) error {
	return nil
}
