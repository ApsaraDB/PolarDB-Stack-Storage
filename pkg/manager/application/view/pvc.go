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

type PvcRequest struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

type PvcCreateRequest struct {
	PvcRequest
	RequestSizeInByte string `json:"request_size_in_byte"`
}

type PvcWriteLockRequest struct {
	PvcRequest
	WriteLockNodeId string `json:"write_lock_node_id"`
	WriteLockNodeIp string `json:"write_lock_node_ip"`
}

//volume id: lun wwid or lv name
type PvcBindVolumeRequest struct {
	PvcRequest
	LvType       common.LvType `json:"volume_type"`
	VolumeId     string        `json:"volume_id"`
	NeedFormat   bool          `json:"need_format"`
	ResourceId   string        `json:"resource_id"` //pvc对应的clusterID
	StorageClass string        `json:"storage_class"`
}

type PvcCreateWithVolumeRequest struct {
	PvcRequest
	LvType     common.LvType `json:"volume_type"`
	VolumeId   string        `json:"volume_id"`
	NeedFormat bool          `json:"need_format"`
}

type PvcFormatRequest struct {
	PvcRequest
}

type PvcExpandFsRequest struct {
	PvcRequest
	LvType   common.LvType `json:"volume_type"`
	VolumeId string        `json:"volume_id"`
	FsType   common.FsType `json:"fs_type"`
	ReqSize  int64         `json:"req_size"`
}

type PvcExpandFsResponse struct {
	Status     domain.VolumeStatusValue `json:"status"`
	ErrMessage string                   `json:"err_message"`
}

type PvcIsReadyResponse struct {
	IsReady bool `json:"is_ready"`
}

type Node struct {
	NodeId string `json:"node_id"`
	NodeIp string `json:"node_ip"`
}

type PvcVolumePermissionTopoResponse struct {
	WriteNode Node   `json:"write_node"`
	ReadNodes []Node `json:"read_nodes"`
}

type PvcResponse struct {
	Name          string              `json:"name"`
	Namespace     string              `json:"namespace"`
	LvType        common.VolumeClass  `json:"volume_type"`
	VolumeId      string              `json:"volume_id"`
	VolumeName    string              `json:"volume_name"`
	SizeInByte    int64               `json:"size_in_byte"`
	FsType        common.FsType       `json:"fs_type"`
	Usable        bool                `json:"usable"`
	DbClusterName string              `json:"db_cluster_name"`
	Status        domain.VolumeStatus `json:"status"`
	CreateTime    string              `json:"create_time"`
}
