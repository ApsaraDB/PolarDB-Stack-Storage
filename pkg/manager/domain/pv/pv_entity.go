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
	"polardb-sms/pkg/common"
	"polardb-sms/pkg/manager/domain"
)

type Host struct {
	Hostname  string `json:"hostname"`
	HostIp    string `json:"host_ip"`
	ClusterId int    `json:"cluster_id"`
}

//represent a disk in host
type PhysicalVolumeEntity struct {
	Host
	domain.VolumeInfo
	PvType     common.PvType       `json:"disk_type"`
	Status     domain.VolumeStatus `json:"status"`
	UsedBy     domain.UsedByType   `json:"used_by"`
	UsedByName string              `json:"used_by_name"`
}

func (p *PhysicalVolumeEntity) GetVolumeName() string {
	return p.VolumeInfo.VolumeName + "#" + p.Hostname
}

func (p *PhysicalVolumeEntity) Type() common.VolumeType {
	return common.Pv
}

func (p *PhysicalVolumeEntity) GetVolumeId() string {
	return p.VolumeInfo.VolumeId + "#" + p.Hostname
}
