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

package domain

import (
	"polardb-sms/pkg/common"
	"polardb-sms/pkg/device"
	smslog "polardb-sms/pkg/log"
)

type UsedByType int

const (
	Non UsedByType = iota
	DBUsed
	LvUsed
)

type Volume interface {
	Type() common.VolumeType
	GetVolumeId() string
	GetVolumeName() string
	GetCapacity() int64
	GetUsedSize() int64
	GetFsType() common.FsType
	GetFsSize() int64
	GetSectors() int64
	GetSectorSize() int
	GetPathNum() int
}

type VolumeInfo struct {
	VolumeName   string                  `json:"volume_name"`
	VolumeId     string                  `json:"volume_id"`
	Paths        []string                `json:"paths"`
	PathNum      int                     `json:"path_num"`
	Vendor       string                  `json:"vendor"`
	Product      string                  `json:"product"`
	Size         int64                   `json:"size"`
	Sectors      int64                   `json:"sectors"`
	SectorSize   int                     `json:"sectorSize"`
	FsType       common.FsType           `json:"fsType"`
	FsSize       int64                   `json:"fs_size"`
	PrSupport    *device.PrSupportReport `json:"prSupport"`
	UsedSize     int64                   `json:"used_size"`
	SerialNumber string                  `json:"serial_number"`
}

func (v *VolumeInfo) GetVolumeName() string {
	return v.VolumeName
}

func (v *VolumeInfo) GetVolumeId() string {
	return v.VolumeId
}

func (v *VolumeInfo) GetCapacity() int64 {
	return v.Size
}

func (v *VolumeInfo) GetUsedSize() int64 {
	return v.UsedSize
}

func (v *VolumeInfo) GetFsType() common.FsType {
	return v.FsType
}

func (v *VolumeInfo) GetFsSize() int64 {
	return v.FsSize
}

func (v *VolumeInfo) GetSectors() int64 {
	return v.Sectors
}

func (v *VolumeInfo) GetSectorSize() int {
	return v.SectorSize
}

func (v *VolumeInfo) GetPathNum() int {
	return v.PathNum
}

func (v *VolumeInfo) GetSerialNumber() string {
	return v.SerialNumber
}

type VolumeStatusValue int

const (
	Success VolumeStatusValue = iota
	Fail
	Expanding
	Creating
	Deleting
	Formatting
	Releasing
	PrLocking
	NoAction
)

type ErrorCode string

const (
	CreateError ErrorCode = "Create_Err"
	ExpandError ErrorCode = "Expand_Err"
	FormatError ErrorCode = "Format_Err"
	DeleteError ErrorCode = "Delete_Err"
	NoError     ErrorCode = ""
)

type VolumeStatus struct {
	StatusValue  VolumeStatusValue `json:"status_value"`
	ErrorCode    ErrorCode         `json:"error_code"`
	ErrorMessage string            `json:"error_message"`
}

var (
	NonStatus = &VolumeStatus{
		StatusValue:  NoAction,
		ErrorCode:    NoError,
		ErrorMessage: "",
	}
)

func (vs *VolumeStatus) String() string {
	bytes, err := common.StructToBytes(vs)
	if err != nil {
		return ""
	}
	return string(bytes)
}

func ParseVolumeStatus(str string) VolumeStatus {
	vs := VolumeStatus{}
	err := common.BytesToStruct([]byte(str), &vs)
	if err != nil {
		smslog.Debugf("parse VolumeStatus err %s", err.Error())
	}
	return vs
}
