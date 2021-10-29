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

package k8spvc

import (
	"k8s.io/apimachinery/pkg/api/resource"
	smslog "polardb-sms/pkg/log"
	"testing"
)

func TestPersistVolumeClaimEntity_SetRequestSize(t *testing.T) {
	pvcEntity := &PersistVolumeClaimEntity{
		Name:               "123",
		DiskStatus:         &VolumeMeta{},
		ExpectedDiskStatus: &VolumeMeta{},
	}

	pvcEntity.SetRequestSize(123456790989)
	smslog.Infof("%s", pvcEntity.ExpectedDiskStatus.Size)
	q, err := resource.ParseQuantity(pvcEntity.ExpectedDiskStatus.Size)
	if err != nil {
		smslog.Infof(err.Error())
	}
	smslog.Infof("%v", q)
}
