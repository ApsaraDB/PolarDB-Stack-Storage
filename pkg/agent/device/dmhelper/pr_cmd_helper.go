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


package dmhelper

import (
	"polardb-sms/pkg/agent/device/reservation/mpathpersist"
	"polardb-sms/pkg/agent/device/reservation/nvme"
	"polardb-sms/pkg/agent/utils"
	"polardb-sms/pkg/device"
	smslog "polardb-sms/pkg/log"
)

func getPrSupportParam(name string) *device.PrSupportReport {
	ret := &device.PrSupportReport{
	}
	devicePath := "/dev/mapper/" + name
	var prInfo *mpathpersist.PersistentReserve
	var err error
	if utils.CheckNvmeVolumeStartWith3(name) {
		prInfo, err = nvme.GetPrInfo(devicePath)
	} else {
		prInfo, err = mpathpersist.GetPrInfo(devicePath)
	}
	if err == nil {
		smslog.Debugf("get devicePath %s pr info %v", devicePath, prInfo)
		if prInfo.ReservationType == mpathpersist.PRC_WR_EX {
			if prInfo.ReservationKey != "" {
				if prInfo.ReservationKey != "0x0" {
					ret.PrKey = prInfo.ReservationKey
				} else {
					if len(prInfo.Keys) == 1 {
						for k, _ := range prInfo.Keys {
							ret.PrKey = k
						}
					}
				}
			}
		}
	} else {
		smslog.Infof("failed get pr info for %s, err %s", name, err.Error())
	}
	capability, err := mpathpersist.ReportCapabilities(devicePath)
	if err == nil {
		ret.Pr7Supported = capability.Support(mpathpersist.PRC_WR_EX)
		ret.PrCapacities = capability.String()
	} else {
		smslog.Infof("failed get pr capabilities for %s, err %s", name, err.Error())
	}
	return ret
}
