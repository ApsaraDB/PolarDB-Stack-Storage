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

package common

import "time"

type FsType string

const (
	NoFs FsType = ""
	Ext4 FsType = "ext4"
	Pfs  FsType = "pfs"
)

func ParseFsType(fsTypeStr string) FsType {
	switch fsTypeStr {
	case "ext4":
		return Ext4
	case "pfs":
		return Pfs
	default:
		return NoFs
	}
}

func (f FsType) Int() int {
	switch f {
	case Ext4:
		return 1
	case Pfs:
		return 2
	default:
		return 0
	}
}

type VolumeType string

const (
	Lv VolumeType = "logical-volume"
	Pv VolumeType = "physical-volume"
)

func ParseLvTypes(str string) []LvType {
	switch str {
	case "lun":
		return []LvType{MultipathVolume}
	case "lv":
		return []LvType{DmLinearVolume, DmStripVolume, DmMirrorVolume}
	}
	return nil
}

type VolumeClass string

const (
	LvClass  VolumeClass = "lv"
	LunClass VolumeClass = "lun"
)

type LvType string

const (
	//未定义
	Non LvType = ""
	//本地盘
	LocalVolume LvType = "local"
	//传统的san存储直接做volume
	MultipathVolume LvType = "multipath"
	//自定义的linear, mirror等设备
	DmLinearVolume LvType = "dm-linear"
	DmStripVolume  LvType = "dm-stripe"
	DmMirrorVolume LvType = "dm-mirror"
)

func (t LvType) ToVolumeClass() VolumeClass {
	switch t {
	case MultipathVolume:
		return LunClass
	default:
		return LvClass
	}
}

type PvType string

const (
	LocalDisk       PvType = "local"
	SanDisk         PvType = "san"
	DistributedDisk PvType = "distributed" //extend for different
)

const DmNamePrefix = "lvid-"

// constants of SSH dial config
const (
	SSHDialAuthPath       = "/home/"
	SSHDialAuthCert       = "/.ssh/id_rsa"
	SSHDialRootUser       = "root"
	SSHDialRootPath       = "/"
	SSHDialDefaultPort    = 22
	RESTDefaultPort       = 8088
	SSHDialDefaultTimeout = 25 * time.Second
)

// constants of SAN request access method
const (
	CLI         = "cli"
	Restful     = "rest"
	PreStopPort = "9809"
	PreStopPath = "/v1/restful/logout"
)
