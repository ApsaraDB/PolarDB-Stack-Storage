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
	smslog "polardb-sms/pkg/log"
	"polardb-sms/pkg/manager/application/view"
	"polardb-sms/pkg/manager/domain"
	"polardb-sms/pkg/manager/domain/k8spvc"
	"polardb-sms/pkg/manager/domain/lv"
	"strings"
)

type ClusterLunAssembler interface {
	ToClusterLunEntity(v *view.ClusterLunCreateRequest) *lv.LogicalVolumeEntity
	ToClusterLunEntities(vs []*view.ClusterLunCreateRequest) []*lv.LogicalVolumeEntity
	ToClusterLunView(e *lv.LogicalVolumeEntity, lvPvcMap map[string]PvcInner) *view.ClusterLunResponse
	ToClusterLunViews(es []*lv.LogicalVolumeEntity) []*view.ClusterLunResponse
}

type ClusterLunAssemblerImpl struct {
}

func (as *ClusterLunAssemblerImpl) ToClusterLunEntity(v *view.ClusterLunCreateRequest) *lv.LogicalVolumeEntity {
	e := lv.LogicalVolumeEntity{
		VolumeInfo: domain.VolumeInfo{
			VolumeName: v.Name,
			VolumeId:   v.Wwid,
			Vendor:     v.Vendor,
			Size:       v.Size,
			SectorSize: v.SectorSize,
			Sectors:    v.SectorNum,
			FsType:     v.FsType,
		},
		NodeIds:   strings.Split(v.NodeIds, ";"),
		ClusterId: v.ClusterId,
		Desc:      v.Desc,
	}
	return &e
}

func (as *ClusterLunAssemblerImpl) ToClusterLunEntities(vs []*view.ClusterLunCreateRequest) []*lv.LogicalVolumeEntity {
	var es []*lv.LogicalVolumeEntity
	for _, v := range vs {
		e := as.ToClusterLunEntity(v)
		es = append(es, e)
	}
	return es
}

func (as *ClusterLunAssemblerImpl) ToClusterLunView(e *lv.LogicalVolumeEntity, lvPvcMap map[string]PvcInner) *view.ClusterLunResponse {
	v := view.ClusterLunResponse{
		Name:            e.VolumeName,
		Wwid:            e.VolumeId,
		VolumeType:      e.LvType,
		Vendor:          e.Vendor,
		Product:         e.Product,
		Size:            e.Size,
		SectorSize:      e.SectorSize,
		SectorNum:       e.Sectors,
		FsType:          e.FsType,
		FsSize:          e.FsSize,
		Paths:           nil,
		NodeIds:         strings.Join(e.NodeIds, ";"),
		ClusterId:       e.ClusterId,
		PrSupportStatus: e.PrInfo.String(),
		Desc:            e.Desc,
		Status:          e.Status,
		UsedSize:        e.UsedSize,
		DbClusterName:   "",
		PvcName:         e.GetPvcName(),
		LvName:          e.GetLvName(),
		Usable:          e.Usable(),
	}
	for _, child := range e.Children.Items {
		if v.PathNum < child.GetPathNum() {
			v.PathNum = child.GetPathNum()
		}
	}

	innerPvc, ok := lvPvcMap[e.VolumeId]
	if !ok {
		return &v
	}
	v.DbClusterName = innerPvc.DbClusterName
	v.PvcName = innerPvc.PvcName
	v.Usable = false
	return &v
}

type PvcInner struct {
	PvcName       string `json:"pvc_name"`
	PvcNamespace  string `json:"pvc_namespace"`
	DbClusterName string `json:"db_cluster_name"`
}

func GetLvPvcMap() map[string]PvcInner {
	var ret = make(map[string]PvcInner, 0)
	pvcs, err := k8spvc.GetPvcRepository().QueryAll()
	if err != nil {
		smslog.Errorf("find LvPvcMap err %s", err.Error())
		return ret
	}
	for _, pvc := range pvcs {
		if pvc.DiskStatus.VolumeId != "" {
			ret[pvc.DiskStatus.VolumeId] = PvcInner{
				PvcName:       pvc.Name,
				PvcNamespace:  pvc.Namespace,
				DbClusterName: pvc.DbClusterName,
			}
		}
	}
	return ret
}

func (as *ClusterLunAssemblerImpl) ToClusterLunViews(es []*lv.LogicalVolumeEntity) []*view.ClusterLunResponse {
	lvPvcMap := GetLvPvcMap()
	var vs []*view.ClusterLunResponse
	for _, e := range es {
		v := as.ToClusterLunView(e, lvPvcMap)
		vs = append(vs, v)
	}
	return vs
}

func NewClusterLunAssembler() ClusterLunAssembler {
	as := &ClusterLunAssemblerImpl{}
	return as
}
