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
	"fmt"
	"polardb-sms/pkg/common"
	"polardb-sms/pkg/device"
	smslog "polardb-sms/pkg/log"
	"polardb-sms/pkg/manager/application/view"
	"polardb-sms/pkg/manager/config"
	"polardb-sms/pkg/manager/domain"
	"polardb-sms/pkg/manager/domain/lv"
	"strconv"
	"strings"
)

const (
	GiB = 1024 * 1024 * 1024
	MiB = 1024 * 1024
)

type ClusterLvAssembler interface {
	ToClusterLvEntity(v *view.ClusterLvCreateRequest) *lv.LogicalVolumeEntity
	ToClusterLvEntities(vs []*view.ClusterLvCreateRequest) []*lv.LogicalVolumeEntity
	ToClusterLvView(e *lv.LogicalVolumeEntity, lvPvcMap map[string]PvcInner) *view.ClusterLvResponse
	ToClusterLvViews(es []*lv.LogicalVolumeEntity) []*view.ClusterLvResponse
}

type ClusterLvAssemblerImpl struct {
	lvRepo lv.LvRepository
}

//TODO fix here
func (as *ClusterLvAssemblerImpl) ToClusterLvEntity(v *view.ClusterLvCreateRequest) *lv.LogicalVolumeEntity {
	e := &lv.LogicalVolumeEntity{
		VolumeInfo: domain.VolumeInfo{
			VolumeName: v.Name,
			VolumeId:   common.DmNamePrefix + v.Name,
			Size:       v.Size,
			Sectors:    v.SectorNum,
			SectorSize: v.SectorSize,
			PrSupport:  &device.PrSupportReport{},
			UsedSize:   0,
		},
		LvType: common.LvType(v.Mode),
		Status: domain.VolumeStatus{
			StatusValue:  domain.Creating,
			ErrorCode:    "",
			ErrorMessage: "",
		},
		Children: &lv.Children{},
		PrInfo:   make(map[string]*lv.PrCheckList),
		Extend:   make(map[string]interface{}, 0),
		NodeIds:  make([]string, 0),
	}
	for _, multipathVolume := range v.Luns {
		item, err := as.lvRepo.FindByVolumeId(multipathVolume.VolumeId)
		if err != nil {
			smslog.Errorf("can not fine pv for %v: %v", multipathVolume, err)
			continue
		}
		e.Children.AddChild(item)
	}
	dmDevice, err := e.GetDmDevice()
	if err == nil {
		e.Size = int64(dmDevice.SectorSize) * dmDevice.SectorNum
		e.SectorSize = dmDevice.SectorSize
		e.Sectors = dmDevice.SectorNum
	}
	for nodeId, _ := range config.GetAvailableNodes() {
		e.NodeIds = append(e.NodeIds, nodeId)
	}

	return e
}

func (as *ClusterLvAssemblerImpl) ToClusterLvEntities(vs []*view.ClusterLvCreateRequest) []*lv.LogicalVolumeEntity {
	var es []*lv.LogicalVolumeEntity
	for _, v := range vs {
		e := as.ToClusterLvEntity(v)
		es = append(es, e)
	}
	return es
}

//TODO fix this
func (as *ClusterLvAssemblerImpl) ToClusterLvView(e *lv.LogicalVolumeEntity, lvPvcMap map[string]PvcInner) *view.ClusterLvResponse {
	v := &view.ClusterLvResponse{
		VolumeName: e.VolumeName,
		VolumeId:   e.VolumeId,
		VolumeType: e.LvType,
		Size:       e.Size,
		SectorSize: e.SectorSize,
		SectorNum:  e.Sectors,
		FsType:     e.FsType,
		FsSize:     e.FsSize,
		UsedSize:   e.UsedSize,
		NodeIds:    strings.Join(e.NodeIds, ";"),
		Status:     e.Status,
		Usable:     e.Usable(),
	}

	var luns []view.MultipathVolumeView
	for _, lunInf := range e.Children.Items {
		luns = append(luns, view.MultipathVolumeView{
			VolumeName: lunInf.GetVolumeName(),
			VolumeId:   lunInf.GetVolumeId(),
			Capacity:   parseCapacityBytesToGiB(lunInf.GetCapacity()),
			SectorNum: lunInf.GetSectors(),
			SectorSize: lunInf.GetSectorSize(),
		})
	}
	v.Children = luns
	innerPvc, ok := lvPvcMap[e.VolumeId]
	if !ok {
		return v
	}
	v.DbClusterName = innerPvc.DbClusterName
	v.PvcName = innerPvc.PvcName
	v.Usable = false
	return v
}

func (as *ClusterLvAssemblerImpl) ToClusterLvViews(es []*lv.LogicalVolumeEntity) []*view.ClusterLvResponse {
	lvPvcMap := GetLvPvcMap()
	var vs []*view.ClusterLvResponse
	for _, e := range es {
		v := as.ToClusterLvView(e, lvPvcMap)
		vs = append(vs, v)
	}
	return vs
}

func NewClusterLvAssembler() ClusterLvAssembler {
	as := &ClusterLvAssemblerImpl{
		lvRepo: lv.GetLvRepository(),
	}
	return as
}

func parseCapacityGiBToBytes(capacityStr string) int64 {
	var (
		capacityBytes int64
		storageUnit   string
		storageSize   string
	)

	for index, value := range capacityStr {
		if len(capacityStr) == 1 && capacityStr[0] == 48 {
			storageSize = "1024"
			break
		}

		/*
			Linux ASCII number 0~9(48-57) and period(46)
		*/
		if (value >= 48 && value <= 57) || value == 46 {
			continue
		} else {
			storageSize = capacityStr[:index]
			storageUnit = capacityStr[index:len(capacityStr)]
			break
		}
	}

	splitCapacity, _ := strconv.ParseFloat(storageSize, 64)

	switch storageUnit {
	case "t", "T", "tb", "TB", "Ti", "TiB":
		capacityBytes = int64(splitCapacity * 1024 * GiB)
	case "g", "G", "gb", "GB", "Gi", "GiB":
		capacityBytes = int64(splitCapacity * GiB)
	case "m", "M", "mb", "MB", "Mi", "MiB":
		capacityBytes = int64(splitCapacity * MiB)
	case "k", "K", "kb", "KB", "Ki", "KiB":
		capacityBytes = int64(splitCapacity * 1024)
	case "B":
		capacityBytes = int64(splitCapacity)
	default:
		capacityBytes = int64(splitCapacity * GiB)
	}

	return capacityBytes
}

func parseCapacityBytesToGiB(capacityBytes int64) string {
	return fmt.Sprintf("%dGiB", capacityBytes/GiB)
}
