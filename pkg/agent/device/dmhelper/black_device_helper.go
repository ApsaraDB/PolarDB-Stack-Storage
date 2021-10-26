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
	"io/ioutil"
	"polardb-sms/pkg/common"
	"strings"
)

var AgentFile = "/etc/polardb-sms-agent.conf"

type BlackList struct {
	Wwids []string `json:"wwids"`
	Paths []string `json:"paths"`
}

var blackList = BlackList{
	Wwids: make([]string, 0),
	Paths: make([]string, 0),
}

func InitBlacklist() {
	if !common.PathExists(AgentFile) {
		data, err := common.StructToBytes(&blackList)
		if err == nil {
			common.WriteToFile(AgentFile, string(data))
		}
	} else {
		bytes, err := ioutil.ReadFile(AgentFile)
		if err != nil {
			return
		}
		common.BytesToStruct(bytes, &blackList)
	}
}

func BlackListFiltered(disk string) bool {
	for _, wwid := range blackList.Wwids {
		if strings.Contains(disk, wwid) {
			return true
		}
	}
	for _, path := range blackList.Paths {
		if strings.Contains(disk, path) {
			return true
		}
	}
	return false
}
