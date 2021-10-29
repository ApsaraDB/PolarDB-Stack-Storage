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

//todo refact this !!! for has storage permission
type ClusterLunCreateRequest struct {
	Name            string        `json:"name"`
	Wwid            string        `json:"wwid"`
	Vendor          string        `json:"vendor"`
	Size            int64         `json:"size"`
	SectorSize      int           `json:"sector_size"`
	SectorNum       int64         `json:"sector_num"`
	FsType          common.FsType `json:"fs_type"`
	Path            string        `json:"path"`
	NodeIds         string        `json:"node_ids"`
	ClusterId       int           `json:"cluster_id"`
	PrSupportStatus string        `json:"pr_support_status"`
	Desc            string        `json:"desc"`
}

type ClusterLunSameSanRequest struct {
	VolumeIds []string `json:"volume_ids"`
}

type ClusterLunFormatRequest struct {
	Name   string        `json:"name"`
	Wwid   string        `json:"wwid"`
	FsType common.FsType `json:"fs_type"`
	FsSize int64         `json:"fs_size"`
}

type ClusterLunFsExpandRequest struct {
	Name         string        `json:"name"`
	Wwid         string        `json:"wwid"`
	FsType       common.FsType `json:"fs_type"`
	ExpandFsSize int64         `json:"fs_size"`
	PvName       string        `json:"pv_name"`
}

type LunAccessPermissionRequest struct {
	Wwid      string   `json:"wwid"`
	RwNodeIp  string   `json:"rw_node"`
	RoNodesIp []string `json:"ro_nodes"`
}

type ClusterLunFormatAndLockRequest struct {
	Name      string        `json:"name"`
	Wwid      string        `json:"wwid"`
	FsType    common.FsType `json:"fs_type"`
	FsSize    int64         `json:"fs_size"`
	RwNodeIp  string        `json:"rw_node"`
	RoNodesIp []string      `json:"ro_nodes"`
}

type LvMultipathStatus struct {
	ErrorMessage  string `json:"error_message"`
	CurrentStatus string `json:"current_status"`
}

type ClusterLunResponse struct {
	Name            string              `json:"name"`
	Wwid            string              `json:"wwid"`
	VolumeType      common.LvType       `json:"volume_type"`
	Vendor          string              `json:"vendor"`
	Product         string              `json:"product"`
	Size            int64               `json:"size"`
	SectorSize      int                 `json:"sector_size"`
	SectorNum       int64               `json:"sector_num"`
	FsType          common.FsType       `json:"fs_type"`
	FsSize          int64               `json:"fs_size"`
	Paths           []string            `json:"paths"`
	PathNum         int                 `json:"path_num"`
	NodeIds         string              `json:"node_ids"`
	ClusterId       int                 `json:"cluster_id"`
	PrSupportStatus string              `json:"pr_support_status"`
	Desc            string              `json:"desc"`
	Status          domain.VolumeStatus `json:"status"`
	UsedSize        int64               `json:"used_size"`
	DbClusterName   string              `json:"db_cluster_name"`
	PvcName         string              `json:"pvc_name"`
	LvName          string              `json:"lv_name"`
	Usable          bool                `json:"usable"`
	CreateTime      string              `json:"create_time"`
}
