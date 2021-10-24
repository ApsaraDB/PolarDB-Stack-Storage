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

package mpathpersist

import (
	"polardb-sms/pkg/common"
	smslog "polardb-sms/pkg/log"
	"polardb-sms/pkg/network/message"
	"testing"
)

func TestPrExecWrapper_Process(t *testing.T) {
	cmd := &message.PrCmd{
		CmdType:    message.PrRelease,
		VolumeType: common.Non,
		VolumeId:   "xxx",
		CmdParam:   nil,
	}

	_, err := NewPrExecWrapper().Process(cmd)
	if err != nil {
		smslog.Errorf("%v", err)
	}
}
