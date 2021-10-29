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

package device

import (
	"github.com/stretchr/testify/assert"
	"polardb-sms/pkg/common"
	"testing"
)

func TestParseFromLine(t *testing.T) {
	testCase := assert.New(t)
	line := "0 1048576 linear /dev/loop0 8"
	ret, _, err := ParseFromLine(line)
	testCase.NoError(err)
	smslog.Infof("%v", ret)

	line = "1048576 1048576 linear /dev/loop1 8"
	ret, _, err = ParseFromLine(line)
	testCase.NoError(err)
	smslog.Infof("%v", ret)

	line = "0 4194304 striped 2 32 /dev/loop0 8 /dev/loop1 8"
	ret, _, err = ParseFromLine(line)
	testCase.NoError(err)
	smslog.Infof("%v", ret)

	line = "4194304 4194304 striped 2 32 /dev/loop2 8 /dev/loop3 8"
	ret, _, err = ParseFromLine(line)
	testCase.NoError(err)
	smslog.Infof("%v", ret)
}

func TestMultipathDeviceTarget_GetValue(t *testing.T) {
	testCase := assert.New(t)
	multipathDevice := &DmDevice{
		Name:       "xx",
		DeviceType: Multipath,
		SectorNum:  0,
		SectorSize: 100,
		FsType:     common.Ext4,
		FsSize:     20,
		DmTarget: &MultipathDeviceTarget{
			Wwid:            "3xxx",
			Vendor:          "huawei",
			Paths:           []string{"aaa", "bbb"},
			PathNum:         2,
			PrSupportStatus: "",
			Extend:          "",
			DmTableItem: &DmTableItem{
				LogicalStartSector: 0,
				NumSectors:         100,
				TargetArgs: &MultipathArgs{
					PathNum: 2,
					Paths:   []string{"aaaa", "bbbb"},
				},
			},
		},
	}
	smslog.Infof(multipathDevice.DmTarget.(*MultipathDeviceTarget).Vendor)
	vendor, _ := multipathDevice.DmTarget.GetValue(VendorKey)
	testCase.Equal(multipathDevice.DmTarget.(*MultipathDeviceTarget).Vendor, vendor.(string))
}
