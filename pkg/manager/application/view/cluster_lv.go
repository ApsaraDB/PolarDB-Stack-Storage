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


package view

import (
	"polardb-sms/pkg/common"
	"polardb-sms/pkg/manager/domain"
)

type MultipathVolumeView struct {
	VolumeName string `json:"name"`
	VolumeId   string `json:"volume_id"`
	Capacity   string `json:"capacity"`
	SectorNum  int64  `json:"sector_num"`
	SectorSize int    `json:"sector_size"`
}

type ClusterLvCreateRequest struct {
	Name       string                `json:"name"`
	Vendor     string                `json:"vendor"`
	Luns       []MultipathVolumeView `json:"luns"`
	Mode       string                `json:"mode"`
	Size       int64                 `json:"size"`
	SectorSize int                   `json:"sector_size"`
	SectorNum  int64                 `json:"sector_num"`
	DmTable    string                `json:"dm_table"`
}

type ClusterLvExpandRequest struct {
	*ClusterLvCreateRequest
}

type ClusterLvFormatRequest struct {
	VolumeName string        `json:"volume_name"`
	VolumeId   string        `json:"volume_id"`
	FsType     common.FsType `json:"fs_type"`
	FsSize     int64         `json:"fs_size"`
}

type ClusterLvFsExpandRequest struct {
	VolumeName string        `json:"volume_name"`
	VolumeId   string        `json:"volume_id"`
	FsType     common.FsType `json:"fs_type"`
	ReqSize    int64         `json:"req_size"`
}

type LvDmDeviceStatus struct {
	CurrentStatus string `json:"current_status"`
	ErrorMessage  string `json:"error_message"`
}

type ClusterLvResponse struct {
	VolumeName      string                `json:"volume_name"`
	VolumeId        string                `json:"volume_id"`
	VolumeType      common.LvType         `json:"volume_type"`
	Size            int64                 `json:"size"`
	SectorSize      int                   `json:"sector_size"`
	SectorNum       int64                 `json:"sector_num"`
	FsType          common.FsType         `json:"fs_type"`
	FsSize          int64                 `json:"fs_size"`
	NodeIds         string                `json:"node_ids"`
	ClusterId       int                   `json:"cluster_id"`
	PrSupportStatus string                `json:"pr_support_status"`
	Desc            string                `json:"desc"`
	Status          domain.VolumeStatus   `json:"status"`
	UsedSize        int64                 `json:"used_size"`
	DbClusterName   string                `json:"db_cluster_name"`
	PvcName         string                `json:"pvc_name"`
	LvName          string                `json:"lv_name"`
	Children        []MultipathVolumeView `json:"children"`
	Usable          bool                  `json:"usable"`
	CreateTime      string                `json:"create_time"`
}
