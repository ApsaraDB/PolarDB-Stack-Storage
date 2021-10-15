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


package devicemapper

import (
	"sync"
)

type DeviceMapper struct {
	*scsi
	*dmSetup
	*multiPath
}

var _deviceMapper *DeviceMapper
var _deviceMapperOnce sync.Once

func GetDeviceMapper() *DeviceMapper {
	_deviceMapperOnce.Do(func() {
		if _deviceMapper == nil {
			_deviceMapper = &DeviceMapper{
				newScsi(),
				newDmSetup(),
				newMultiPath(),
			}
		}
	})
	return _deviceMapper
}
