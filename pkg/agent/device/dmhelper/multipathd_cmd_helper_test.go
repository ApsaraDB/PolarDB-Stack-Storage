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
	"encoding/json"
	"fmt"
	"polardb-sms/pkg/device"
	smslog "polardb-sms/pkg/log"
	"regexp"
	"strings"
	"testing"
)

func TestMultipath(t *testing.T) {
	jsonStr := "{\n   \"major_version\": 0,\n   \"minor_version\": 1,\n   \"map\":{\n      \"name\" : \"36e00084100ee7ec969ce59cd00000c95\",\n      \"uuid\" : \"36e00084100ee7ec969ce59cd00000c95\",\n      \"sysfs\" : \"dm-79\",\n      \"failback\" : \"immediate\",\n      \"queueing\" : \"-\",\n      \"paths\" : 2,\n      \"write_prot\" : \"rw\",\n      \"dm_st\" : \"active\",\n      \"features\" : \"0\",\n      \"hwhandler\" : \"0\",\n      \"action\" : \"\",\n      \"path_faults\" : 0,\n      \"vend\" : \"HUAWEI  \",\n      \"prod\" : \"XSG1            \",\n      \"rev\" : \"6000\",\n      \"switch_grp\" : 0,\n      \"map_loads\" : 1,\n      \"total_q_time\" : 0,\n      \"q_timeouts\" : 0,\n      \"path_groups\": [{\n         \"selector\" : \"service-time 0\",\n         \"pri\" : 1,\n         \"dm_st\" : \"active\",\n         \"group\" : 1,\n         \"paths\": [{\n            \"dev\" : \"sdes\",\n            \"dev_t\" : \"129:64\",\n            \"dm_st\" : \"active\",\n            \"dev_st\" : \"running\",\n            \"chk_st\" : \"ready\",\n            \"checker\" : \"tur\",\n            \"pri\" : 1,\n            \"host_wwnn\" : \"0x200000109b5a6ac9\",\n            \"target_wwnn\" : \"0x2100e00084ee7ec9\",\n            \"host_wwpn\" : \"0x100000109b5a6ac9\",\n            \"target_wwpn\" : \"0x2018e00084ee7ec9\",\n            \"host_adapter\" : \"0000:85:00.0\"\n         },{\n            \"dev\" : \"sdnt\",\n            \"dev_t\" : \"71:496\",\n            \"dm_st\" : \"active\",\n            \"dev_st\" : \"running\",\n            \"chk_st\" : \"ready\",\n            \"checker\" : \"tur\",\n            \"pri\" : 1,\n            \"host_wwnn\" : \"0x200000109b5a6ac9\",\n            \"target_wwnn\" : \"0x2100e00084ee7ec9\",\n            \"host_wwpn\" : \"0x100000109b5a6ac9\",\n            \"target_wwpn\" : \"0x2008e00084ee7ec9\",\n            \"host_adapter\" : \"0000:85:00.0\"\n         }]\n      }]\n   }\n}"
	var retMap = RawMultipathParam{}

	err := json.Unmarshal([]byte(jsonStr), &retMap)
	if err != nil {
		smslog.Errorf(err.Error())
		return
	}

	paths := make([]string, 0)
	for _, pathGroups := range retMap.ParamMap.PathGroups {
		for _, path := range pathGroups.Paths {
			paths = append(paths, path.Dev)
		}
	}
	smslog.Infof("%v", paths)
}

/**
pv-360050767088081329800000000000105 (360050767088081329800000000000105) dm-5 ALIBABA ,MCS
size=300G features='1 queue_if_no_path' hwhandler='0' wp=rw
|-+- policy='round-robin 0' prio=50 status=active
| |- 14:0:1:3 sdl     8:176  active ready running
| |- 15:0:1:3 sdar    66:176 active ready running
| |- 14:0:2:3 sdt     65:48  active ready running
| `- 15:0:2:3 sdaz    67:48  active ready running
`-+- policy='round-robin 0' prio=10 status=enabled
  |- 14:0:0:3 sdd     8:48   active ready running
  |- 15:0:0:3 sdag    66:0   active ready running
  |- 14:0:3:3 sdac    65:192 active ready running
  `- 15:0:3:3 sdbh    67:176 active ready running
*/
func TestMultipathP(t *testing.T) {
	var ret = MultipathParam{}
	output := "pv-360050767088081329800000000000105 (360050767088081329800000000000105) dm-5 ALIBABA ,MCS\nsize=300G features='1 queue_if_no_path' hwhandler='0' wp=rw\n|-+- policy='round-robin 0' prio=50 status=active\n| |- 14:0:1:3 sdl     8:176  active ready running\n| |- 15:0:1:3 sdar    66:176 active ready running\n| |- 14:0:2:3 sdt     65:48  active ready running\n| `- 15:0:2:3 sdaz    67:48  active ready running\n`-+- policy='round-robin 0' prio=10 status=enabled\n  |- 14:0:0:3 sdd     8:48   active ready running\n  |- 15:0:0:3 sdag    66:0   active ready running\n  |- 14:0:3:3 sdac    65:192 active ready running\n  `- 15:0:3:3 sdbh    67:176 active ready running"
	//output1 := "360050767088081329800000000000108 dm-9 ALIBABA ,MCS\nsize=505G features='1 queue_if_no_path' hwhandler='0' wp=rw\n|-+- policy='round-robin 0' prio=50 status=active\n| |- 14:0:0:6 sdax 67:16  active ready  running\n| |- 15:0:1:6 sdbm 68:0   active ready  running\n| |- 14:0:3:6 sdbg 67:160 active ready  running\n| `- 15:0:3:6 sdbs 68:96  active ready  running\n`-+- policy='round-robin 0' prio=10 status=enabled\n  |- 14:0:1:6 sdba 67:64  active ready  running\n  |- 15:0:0:6 sdbj 67:208 active ready  running\n  |- 14:0:2:6 sdbd 67:112 active ready  running\n  `- 15:0:2:6 sdbp 68:48  active ready  running"
	lines := strings.Split(output, device.NewLineSign)
	firstLineFields := strings.Fields(lines[0])
	lenFields := len(firstLineFields)
	ret.Name = firstLineFields[0]
	ret.Wwid = firstLineFields[lenFields-4]
	ret.Wwid = strings.TrimLeft(ret.Wwid, "(")
	ret.Wwid = strings.TrimRight(ret.Wwid, ")")
	ret.Vendor = firstLineFields[lenFields-2]
	ret.Product = firstLineFields[lenFields-1]
	ret.Product = strings.TrimPrefix(ret.Product, device.CommaSign)
	remainLines := lines[1:]
	subPathPattern := regexp.MustCompile(`\d:\d:\d:\d`)
	var paths []string
	for _, line := range remainLines {
		if subPathPattern.MatchString(line) {
			subPathFields := strings.Fields(line)
			if strings.HasPrefix(line, "| ") {
				paths = append(paths, subPathFields[3])
			} else {
				paths = append(paths, subPathFields[2])
			}
		}
	}
	ret.Paths = paths
	ret.PathNum = len(paths)
	println(fmt.Sprintf("%v", ret))
}
