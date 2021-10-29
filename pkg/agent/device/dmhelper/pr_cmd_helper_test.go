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
	"polardb-sms/pkg/device"
	smslog "polardb-sms/pkg/log"
	"testing"
)

func TestParsePrParam(t *testing.T) {
	ret := &device.PrSupportReport{}
	prInfo := &mpathpersist.PersistentReserve{
		Keys:            map[string]int{"0xc6134001": 2},
		Generation:      "0x6e",
		ReservationKey:  "0x0",
		ReservationType: "Write Exclusive, all registrants",
	}
	smslog.Infof("get device %s pr info %v", "", prInfo)
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

	smslog.Infof("%v", ret)
}
