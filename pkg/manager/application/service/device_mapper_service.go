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
	"polardb-sms/pkg/device"
	smslog "polardb-sms/pkg/log"
	"polardb-sms/pkg/manager/application/view"
	"polardb-sms/pkg/manager/domain/lv"
	"polardb-sms/pkg/manager/domain/pv"
)

type DeviceMapperService struct {
	pvRepo pv.PhysicalVolumeRepository
	lvRepo lv.LvRepository
}

func (s *DeviceMapperService) GenerateConf(v *view.GenerateDmCreateCmdRequest) (*view.DmCreateCmdResponse, error) {
	switch v.LvType {
	case common.DmLinearVolume:
		return s.generateLinearConf(v.LunIdsInOrder)
	case common.DmStripVolume:
		return s.generateStripedConf(v.LunIdsInOrder)
	default:
		return nil, fmt.Errorf("current do not support the type %s", v.LvType)
	}
}

func (s *DeviceMapperService) generateLinearConf(lunIdsInOrder []string) (*view.DmCreateCmdResponse, error) {
	lvEntities, err := s.lvRepo.FindByVolumeIds(lunIdsInOrder)
	if err != nil {
		return nil, fmt.Errorf("can not gen Linear Conf for %v", err)
	}
	dmDevice, err := ParseLinearDevice(lvEntities)
	if err != nil {
		return nil, err
	}
	smslog.Debugf("parse dmDevice %s", dmDevice.String())
	return &view.DmCreateCmdResponse{
		LvName:      dmDevice.Name,
		LvType:      common.DmLinearVolume,
		Size:        int64(dmDevice.SectorSize) * dmDevice.SectorNum,
		SectorSize:  dmDevice.SectorSize,
		Sectors:     dmDevice.SectorNum,
		PreviewConf: dmDevice.String(),
	}, nil
}

func (s *DeviceMapperService) generateStripedConf(lunIdsInOrder []string) (*view.DmCreateCmdResponse, error) {
	lvEntities, err := s.lvRepo.FindByVolumeIds(lunIdsInOrder)
	if err != nil {
		return nil, fmt.Errorf("can not gen Linear Conf for %v", err)
	}
	dmDevice, err := ParseStripedDevice(lvEntities)
	if err != nil {
		return nil, err
	}
	return &view.DmCreateCmdResponse{
		LvName:      dmDevice.Name,
		LvType:      common.DmStripVolume,
		Size:        int64(dmDevice.SectorSize) * dmDevice.SectorNum,
		SectorSize:  dmDevice.SectorSize,
		Sectors:     dmDevice.SectorNum,
		PreviewConf: dmDevice.String(),
	}, nil
}

func ParseLinearDevice(lvEntities []*lv.LogicalVolumeEntity) (*device.DmDevice, error) {
	linearDevice := &device.DmDevice{
		DeviceType: device.Linear,
		DmTarget: &device.LinearDeviceTarget{
			DmTableItems: make([]*device.DmTableItem, 0),
		},
	}
	var (
		totalSectorNum int64
		sectorSize     int
		dmTableItems   = make([]*device.DmTableItem, 0)
	)
	for _, pvEntity := range lvEntities {
		numSectors := pvEntity.Sectors - device.DefaultOffsetSector
		dmTableItems = append(dmTableItems, device.NewLinearDmItem(totalSectorNum, numSectors, pvEntity.VolumeId, device.DefaultOffsetSector))
		totalSectorNum += numSectors
		if sectorSize != 0 && sectorSize != pvEntity.SectorSize {
			return nil, fmt.Errorf("sector size not equal %d, %d", sectorSize, pvEntity.SectorSize)
		}
		sectorSize = pvEntity.SectorSize
	}
	linearDevice.SectorNum = totalSectorNum
	linearDevice.SectorSize = sectorSize
	linearDevice.DmTarget.SetValue(device.DmTableItemsKey, dmTableItems)
	return linearDevice, nil
}

func ParseStripedDevice(lvEntities []*lv.LogicalVolumeEntity) (*device.DmDevice, error) {
	stripedDevice := &device.DmDevice{
		DeviceType: device.Striped,
		DmTarget: &device.StripedDeviceTarget{
			DmTableItems: make([]*device.DmTableItem, 0),
		},
	}
	var (
		totalSectorNum int64
		sectorSize     int
		sectorNum      int64
		dmTableItems   = make([]*device.DmTableItem, 0)
	)

	subDevices := make([]string, 0)
	for _, lvEntity := range lvEntities {
		sectorNum = 0
		if sectorSize != 0 && sectorSize != lvEntity.SectorSize {
			return nil, fmt.Errorf("sector size not equal device 1 [%d], device 2 [%d]", sectorSize, lvEntity.SectorSize)
		}
		sectorNum = lvEntity.Sectors
		sectorSize = lvEntity.SectorSize
		subDevices = append(subDevices, lvEntity.VolumeId)
		totalSectorNum += sectorNum
	}
	dmTableItems = append(dmTableItems, device.NewStripedDmItem(0, totalSectorNum, sectorSize, subDevices))
	stripedDevice.SectorNum = totalSectorNum
	stripedDevice.SectorSize = sectorSize
	stripedDevice.DmTarget.SetValue(device.DmTableItemsKey, dmTableItems)
	return stripedDevice, nil
}

func NewDeviceMapperService() *DeviceMapperService {
	return &DeviceMapperService{
		pvRepo: pv.GetPhysicalVolumeRepository(),
		lvRepo: lv.GetLvRepository(),
	}
}
