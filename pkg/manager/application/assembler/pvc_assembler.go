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
	smslog "polardb-sms/pkg/log"
	"polardb-sms/pkg/manager/application/view"
	"polardb-sms/pkg/manager/config"
	"polardb-sms/pkg/manager/domain/k8spvc"
	"polardb-sms/pkg/manager/domain/lv"
	"sort"
	"strconv"
)

const (
	DefaultNamespace = "polardb-sms"
)

type PvcAssembler interface {
	ToPvcByCreateWithVolumeRequest(v *view.PvcCreateWithVolumeRequest) *k8spvc.PersistVolumeClaimEntity
	ToPvcByBindVolumeRequest(v *view.PvcBindVolumeRequest) *k8spvc.PersistVolumeClaimEntity
	ToPvcResponses(pvcs []*k8spvc.PersistVolumeClaimEntity) []*view.PvcResponse
	ToPvcResponse(pvcEntity *k8spvc.PersistVolumeClaimEntity) *view.PvcResponse
	ToPvcVolumePermTopo(prKey string) *view.PvcVolumePermissionTopoResponse
}

type PvcAssemblerImpl struct {
}

func (a *PvcAssemblerImpl) ToPvcByCreateWithVolumeRequest(v *view.PvcCreateWithVolumeRequest) *k8spvc.PersistVolumeClaimEntity {
	pvcEntity := &k8spvc.PersistVolumeClaimEntity{
		Name:      v.Name,
		Namespace: v.Namespace,
		PvName:    fmt.Sprintf("pv-%s", v.VolumeId),
		DiskStatus: &k8spvc.VolumeMeta{
			VolumeId:   v.VolumeId,
			VolumeType: v.LvType,
			VolumeMode: k8spvc.Block,
			NeedFormat: v.NeedFormat,
		},
		ExpectedDiskStatus: &k8spvc.VolumeMeta{
			VolumeId:   v.VolumeId,
			VolumeType: v.LvType,
			VolumeMode: k8spvc.Block,
			NeedFormat: v.NeedFormat,
		},
		StorageClassName: "local-storage",
	}
	return pvcEntity
}

func (a *PvcAssemblerImpl) ToPvcByBindVolumeRequest(v *view.PvcBindVolumeRequest) *k8spvc.PersistVolumeClaimEntity {
	pvcEntity := &k8spvc.PersistVolumeClaimEntity{
		Name:      v.Name,
		Namespace: v.Namespace,
		PvName:    fmt.Sprintf("pv-%s", v.VolumeId),
		DiskStatus: &k8spvc.VolumeMeta{
			VolumeId:   v.VolumeId,
			VolumeType: v.LvType,
			VolumeMode: k8spvc.Block,
			NeedFormat: v.NeedFormat,
		},
		ExpectedDiskStatus: &k8spvc.VolumeMeta{
			VolumeId:   v.VolumeId,
			VolumeType: v.LvType,
			VolumeMode: k8spvc.Block,
			NeedFormat: v.NeedFormat,
		},
		StorageClassName: v.StorageClass,
	}
	return pvcEntity
}

func (a *PvcAssemblerImpl) ToPvcResponses(pvcs []*k8spvc.PersistVolumeClaimEntity) []*view.PvcResponse {
	var result []*view.PvcResponse

	for _, pvc := range pvcs {
		pvcResp := a.ToPvcResponse(pvc)
		result = append(result, pvcResp)
	}
	sort.SliceStable(result, func(i, j int) bool {
		createTimeI, _ := strconv.ParseInt(result[i].CreateTime, 10, 64)
		createTimeJ, _ := strconv.ParseInt(result[j].CreateTime, 10, 64)
		return createTimeI > createTimeJ
	})
	return result
}

func getLvNameById(volumeId string) string {
	lvEntity, err := lv.GetLvRepository().FindByVolumeId(volumeId)
	if err != nil {
		smslog.Errorf("getLvNameById %s err %s", volumeId, err.Error())
		return ""
	}
	if lvEntity != nil {
		return lvEntity.VolumeName
	}
	return ""
}

func (a *PvcAssemblerImpl) ToPvcResponse(pvc *k8spvc.PersistVolumeClaimEntity) *view.PvcResponse {
	response := &view.PvcResponse{
		Name:          pvc.Name,
		Namespace:     pvc.Namespace,
		LvType:        pvc.DiskStatus.VolumeType.ToVolumeClass(),
		VolumeId:      pvc.DiskStatus.VolumeId,
		VolumeName:    "",
		SizeInByte:    pvc.DiskStatus.Size,
		FsType:        common.Pfs,
		Usable:        pvc.Usable(),
		DbClusterName: pvc.DbClusterName,
		Status:        pvc.PvcStatus,
		CreateTime:    pvc.CreateTime,
	}
	if lvEntity, err := lv.GetLvRepository().FindByVolumeId(pvc.DiskStatus.VolumeId); err == nil && lvEntity != nil {
		response.VolumeName = lvEntity.VolumeName
		response.SizeInByte = lvEntity.Size
	}
	return response
}

func (a *PvcAssemblerImpl) ToPvcVolumePermTopo(prKey string) *view.PvcVolumePermissionTopoResponse {
	prNodeIp := common.PrKeyToIpV4(prKey)
	prNode := config.GetNodeByIp(prNodeIp)
	if prNode == nil {
		smslog.Infof("PrKey: %s, can not find the prNode by ip %s", prKey, prNodeIp)
		return nil
	}
	prNodeId := prNode.Name
	var writeNode view.Node
	var readNodes []view.Node
	for name, node := range config.GetAvailableNodes() {
		if name == prNodeId {
			writeNode.NodeIp = node.Ip
			writeNode.NodeId = prNodeId
		} else {
			readNodes = append(readNodes, view.Node{
				NodeId: node.Name,
				NodeIp: node.Ip,
			})
		}
	}
	return &view.PvcVolumePermissionTopoResponse{
		WriteNode: writeNode,
		ReadNodes: readNodes,
	}
}

func NewPvcAssembler() PvcAssembler {
	return &PvcAssemblerImpl{}
}
