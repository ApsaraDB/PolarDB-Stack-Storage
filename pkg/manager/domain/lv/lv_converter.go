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


package lv

import (
	"polardb-sms/pkg/common"
	"polardb-sms/pkg/manager/domain"
	"strings"
)

var _ domain.Converter = &LvConverter{}

type LvConverter struct {
}

func (l LvConverter) ToModel(t interface{}) (interface{}, error) {
	lvEntity := t.(*LogicalVolumeEntity)
	return &LogicalVolume{
		VolumeId:     lvEntity.VolumeId,
		VolumeName:   lvEntity.VolumeName,
		Children:     lvEntity.GetChildrenString(),
		LvType:       string(lvEntity.LvType),
		NodeIds:      strings.Join(lvEntity.NodeIds, ","),
		RelatedPvc:   lvEntity.GetPvcName(),
		FsSize:       lvEntity.FsSize,
		FsType:       string(lvEntity.FsType),
		SectorNum:    lvEntity.Sectors,
		SectorSize:   lvEntity.SectorSize,
		Size:         lvEntity.Size,
		UsedSize:     lvEntity.UsedSize,
		Status:       lvEntity.Status.String(),
		PrStatus:     lvEntity.PrInfo.String(),
		PrNodeId:     lvEntity.PrKey,
		Vendor:       lvEntity.Vendor,
		Product:      lvEntity.Product,
		Extend:       lvEntity.Extend.String(),
		UsedByType:   int(lvEntity.UsedByType),
		UsedByName:   lvEntity.UsedByName,
		SerialNumber: lvEntity.SerialNumber,
	}, nil
}

func (l LvConverter) ToEntity(t interface{}) (interface{}, error) {
	lvModel := t.(*LogicalVolume)
	ret := &LogicalVolumeEntity{
		VolumeInfo: domain.VolumeInfo{
			VolumeName:   lvModel.VolumeName,
			VolumeId:     lvModel.VolumeId,
			Vendor:       lvModel.Vendor,
			Product:      lvModel.Product,
			Size:         lvModel.Size,
			Sectors:      lvModel.SectorNum,
			SectorSize:   lvModel.SectorSize,
			FsType:       common.FsType(lvModel.FsType),
			FsSize:       lvModel.FsSize,
			PrSupport:    nil,
			UsedSize:     lvModel.UsedSize,
			SerialNumber: lvModel.SerialNumber,
		},
		LvType:     common.LvType(lvModel.LvType),
		ClusterId:  lvModel.ClusterId,
		PrKey:      lvModel.PrNodeId,
		Status:     domain.ParseVolumeStatus(lvModel.Status),
		Children:   &Children{Items: ParseChildren(lvModel.Children)},
		PrInfo:     ParsePrInfo(lvModel.PrStatus),
		Extend:     ParseExtend(lvModel.Extend),
		NodeIds:    strings.Split(lvModel.NodeIds, ","),
		UsedByType: domain.UsedByType(lvModel.UsedByType),
		UsedByName: lvModel.UsedByName,
	}
	return ret, nil
}

func (l LvConverter) ToEntities(t []interface{}) ([]interface{}, error) {
	var ret []interface{}
	for _, lvModelInf := range t {
		lvEntity, _ := l.ToEntity(lvModelInf)
		ret = append(ret, lvEntity)
	}
	return ret, nil
}

func NewLvConverter() domain.Converter {
	return &LvConverter{}
}
