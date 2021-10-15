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
	"encoding/json"
	smslog "polardb-sms/pkg/log"
	"polardb-sms/pkg/manager/application/view"
	"testing"
)

func TestToClusterLvEntity(t *testing.T) {
	v := &view.ClusterLvCreateRequest{
		Name: "test",
		Luns: []view.MultipathVolumeView{
			{
				VolumeName:     "lun01",
				VolumeId:     "36e00084100ee7ec972e7047800001f6e",
				Capacity: "210G",
			},
		},
		Mode:       0,
		Size:       51200,
		SectorSize: 512,
		SectorNum:  100,
		DmTable:    "0 1048576 linear /dev/loop0 8  　\n    1048576 1048576 linear /dev/loop1 8",
	}

	as := NewClusterLvAssembler()
	e := as.ToClusterLvEntity(v)
	ret, _ := json.Marshal(e)
	smslog.Infof(string(ret))
}
