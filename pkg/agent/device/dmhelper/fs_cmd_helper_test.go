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
	"polardb-sms/pkg/device"
	"strconv"
	"strings"
	"testing"
)

func TestGiBToByte(t *testing.T) {
	stdout := " (0)allocnode: id 0, shift 0, nchild=21, nall 53760, nfree 50977, next 0\n (0)allocnode: id 0, shift 0, nchild=21, nall 43008, nfree 40748, next 0\n (0)allocnode: id 0, shift 0, nchild=21, nall 43008, nfree 40748, next 0"
	for _, row := range strings.Split(stdout, device.NewLineSign) {
		fields := strings.Split(row, ",")
		for _, field := range fields {
			if strings.Contains(field, "nchild=") {
				field := strings.TrimSpace(field)
				sizeInGb, err := strconv.ParseInt(strings.TrimPrefix(field,
					"nchild="), 10, 64)
				if err != nil {
					println(err.Error())
				}
				println(sizeInGb * 1024 * 1024 * 1024)
			}
		}
	}
}

func TestPfsUsed(t *testing.T) {
	stdout := "4438144\t/mapper_36e00084100ee7ec9d83b384b000008d8/"
	fields := strings.Fields(stdout)
	usedKB, err := strconv.ParseInt(fields[0], 10, 64)
	if err != nil {
		println(err.Error())
	}
	println(usedKB * 1000)
}
