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

package pv

import (
	"fmt"
	"polardb-sms/pkg/common"
	"polardb-sms/pkg/device"
	"polardb-sms/pkg/manager/domain"
	"strings"
)

var _ domain.Converter = &PhysicalVolumeConverter{}

type PhysicalVolumeConverter struct {
}

func NewPhysicalVolumeConverter() domain.Converter {
	return &PhysicalVolumeConverter{}
}
func (p PhysicalVolumeConverter) ToModel(t interface{}) (interface{}, error) {
	pvEntity, ok := t.(*PhysicalVolumeEntity)
	if !ok {
		return nil, fmt.Errorf("entity can not convert to PhysicalVolumeEntity model %v", t)
	}
	return &PhysicalVolume{
		VolumeId:        pvEntity.VolumeId,
		VolumeName:      pvEntity.VolumeName,
		PvType:          string(pvEntity.PvType),
		ClusterId:       pvEntity.ClusterId,
		FsSize:          pvEntity.FsSize,
		FsType:          string(pvEntity.FsType),
		NodeId:          pvEntity.Hostname,
		NodeIp:          pvEntity.HostIp,
		PathNum:         pvEntity.PathNum,
		Paths:           strings.Join(pvEntity.Paths, ","),
		PrSupportStatus: pvEntity.PrSupport.String(),
		SectorNum:       pvEntity.Sectors,
		SectorSize:      pvEntity.SectorSize,
		Size:            pvEntity.Size,
		Status:          pvEntity.Status.String(),
		UsedSize:        pvEntity.UsedSize,
		Vendor:          pvEntity.Vendor,
		Product:         pvEntity.Product,
		SerialNumber:    pvEntity.SerialNumber,
	}, nil
}

func (p PhysicalVolumeConverter) ToEntity(t interface{}) (interface{}, error) {
	pvModel, ok := t.(*PhysicalVolume)
	if !ok {
		return nil, fmt.Errorf("convert model to PhysicalVolumeEntity err %v", t)
	}
	return &PhysicalVolumeEntity{
		Host: Host{
			Hostname:  pvModel.NodeId,
			HostIp:    pvModel.NodeIp,
			ClusterId: pvModel.ClusterId,
		},
		VolumeInfo: domain.VolumeInfo{
			VolumeName:   pvModel.VolumeName,
			VolumeId:     pvModel.VolumeId,
			Paths:        strings.Split(pvModel.Paths, ","),
			PathNum:      pvModel.PathNum,
			Vendor:       pvModel.Vendor,
			Product:      pvModel.Product,
			Size:         pvModel.Size,
			Sectors:      pvModel.SectorNum,
			SectorSize:   pvModel.SectorSize,
			FsType:       common.FsType(pvModel.FsType),
			FsSize:       pvModel.FsSize,
			PrSupport:    device.PrSupportReportFromString(pvModel.PrSupportStatus),
			UsedSize:     pvModel.UsedSize,
			SerialNumber: pvModel.SerialNumber,
		},
		PvType: common.PvType(pvModel.PvType),
		Status: domain.ParseVolumeStatus(pvModel.Status),
	}, nil
}

func (p PhysicalVolumeConverter) ToEntities(t []interface{}) ([]interface{}, error) {
	var ret []interface{}
	for _, pvModelInf := range t {
		pvEntityInf, err := p.ToEntity(pvModelInf)
		if err != nil {
			continue
		}
		pvEntity, ok := pvEntityInf.(PhysicalVolume)
		if !ok {
			continue
		}
		ret = append(ret, pvEntity)
	}
	return ret, nil
}
