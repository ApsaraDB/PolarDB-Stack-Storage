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

package assembler

import (
	"polardb-sms/pkg/common"
	"polardb-sms/pkg/manager/domain"
	"polardb-sms/pkg/manager/domain/pv"
	"polardb-sms/pkg/protocol"
)

type PvAssemblerForEvent interface {
	ToClusterLunEntityByEvent(v *protocol.Lun) *pv.PhysicalVolumeEntity
}

type HostLunFullInfoAssemblerForEventImpl struct {
}

func (as *HostLunFullInfoAssemblerForEventImpl) ToClusterLunEntityByEvent(v *protocol.Lun) *pv.PhysicalVolumeEntity {
	e := &pv.PhysicalVolumeEntity{
		Host: pv.Host{
			Hostname:  v.NodeId,
			HostIp:    v.NodeIp,
			ClusterId: 0,
		},
		VolumeInfo: domain.VolumeInfo{
			VolumeName:   v.Name,
			VolumeId:     v.VolumeId,
			Paths:        v.Paths,
			PathNum:      v.PathNum,
			Vendor:       v.Vendor,
			Product:      v.Product,
			Size:         v.Size,
			Sectors:      v.Sectors,
			SectorSize:   v.SectorSize,
			FsType:       v.FsType,
			FsSize:       v.FsSize,
			PrSupport:    v.PrSupport,
			UsedSize:     v.UsedSize,
			SerialNumber: v.SerialNumber,
		},
		PvType: common.PvType(v.LunType),
		Status: *domain.NonStatus,
	}
	return e
}

func NewHostLunFullInfoAssemblerForEvent() PvAssemblerForEvent {
	as := &HostLunFullInfoAssemblerForEventImpl{}
	return as
}
