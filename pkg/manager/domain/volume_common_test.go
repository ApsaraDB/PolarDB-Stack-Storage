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
	"go.uber.org/zap/zapcore"
	"polardb-sms/pkg/common"
	"polardb-sms/pkg/device"
	smslog "polardb-sms/pkg/log"
	lv2 "polardb-sms/pkg/manager/domain/lv"
	"polardb-sms/pkg/manager/domain/pv"
	"testing"
)

func TestLogicalVolumeEntity_SerializeExtend(t *testing.T) {
	lv := &lv2.LogicalVolumeEntity{
		VolumeInfo: VolumeInfo{},
		LvType:     common.Non,
		ClusterId:  0,
		PrKey:      "",
		Desc:       "",
		RelatedPvc: "",
		Status:     0,
		Children:   nil,
		PrInfo:     nil,
		Extend: map[string]interface{}{
			"dmKey": device.DmDevice{
				Name:            "xxxx",
				DeviceType:      "xxx",
				SectorNum:       100,
				SectorSize:      0,
				FsType:          common.NoFs,
				FsSize:          0,
				PrSupportStatus: nil,
				UsedSize:        0,
				DmTarget:        nil,
			},
		},
	}
	smslog.Info(lv.Extend.String())
}

func TestParseChildren(t *testing.T) {
	smslog.InitLogger("", "temp", zapcore.DebugLevel)
	children := lv2.Children{
		ChildrenType: common.Pv,
		Items: []Volume{
			&pv.PhysicalVolumeEntity{
				Host: pv.Host{},
				VolumeInfo: VolumeInfo{
					VolumeName: "aaa",
					VolumeId:   "cccc",
					Paths:      nil,
					PathNum:    0,
					Vendor:     "",
					Product:    "",
					Size:       0,
					Sectors:    0,
					SectorSize: 0,
					FsType:     common.NoFs,
					FsSize:     0,
					PrSupport:  nil,
					UsedSize:   0,
				},
				PvType:     common.LocalDisk,
				Status:     0,
				UsedBy:     0,
				UsedByName: "",
			},
			&pv.PhysicalVolumeEntity{
				Host: pv.Host{},
				VolumeInfo: VolumeInfo{
					VolumeName: "bbb",
					VolumeId:   "dddd",
					Paths:      nil,
					PathNum:    0,
					Vendor:     "",
					Product:    "",
					Size:       0,
					Sectors:    0,
					SectorSize: 0,
					FsType:     common.NoFs,
					FsSize:     0,
					PrSupport:  nil,
					UsedSize:   0,
				},
				PvType:     common.LocalDisk,
				Status:     0,
				UsedBy:     0,
				UsedByName: "",
			},
		},
	}
	strRet := children.String()
	smslog.Info(strRet)
	decode := lv2.ParseChildren(strRet)
	smslog.Infof("%v", decode)
	bytes, err := common.StructToBytes(&children)
	if err != nil {
		smslog.Errorf(err.Error())
		return
	}
	smslog.Infof("%s", bytes)
}
