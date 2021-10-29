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
	"polardb-sms/pkg/manager/domain"
	"testing"
)

func TestVolumeInfo(t *testing.T) {
	var vInfo domain.Volume
	vInfo = &PhysicalVolumeEntity{
		Host: Host{
			Hostname:  "aaa",
			HostIp:    "bbb",
			ClusterId: 0,
		},
		VolumeInfo: domain.VolumeInfo{
			VolumeName: "ccc",
			VolumeId:   "dddd",
			Paths:      nil,
			PathNum:    0,
			Vendor:     "",
			Product:    "",
			Size:       0,
			Sectors:    0,
			SectorSize: 0,
			FsType:     "",
			FsSize:     0,
			PrSupport:  nil,
			UsedSize:   0,
		},
		PvType:     "",
		Status:     0,
		UsedBy:     0,
		UsedByName: "",
	}
	println(vInfo.GetVolumeId())
}
