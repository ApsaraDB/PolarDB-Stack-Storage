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


package network

import (
	"encoding/json"
	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/assert"
	smslog "polardb-sms/pkg/log"
	"polardb-sms/pkg/network/message"
	"testing"
)

type DmExecContext struct {
	DmLines string `json:"dm_lines`
}

func TestParse2(t *testing.T) {
	testCase := assert.New(t)
	msg, err := Parse("\nd\u0012$2e951a96-e16d-47be-8fba-ca20054a8d97\u0018\u0002\u001A\u0002{}")
	testCase.NoError(err)
	smslog.Infof("%v", msg)
}

func TestSend(t *testing.T) {
	dmExecCtx := DmExecContext{
		DmLines: "0 1048576 linear /dev/loop0 8  　\n    1048576 1048576 linear /dev/loop1 8",
	}
	//convert to msg
	msgBuilder := message.NewSmsMessageBuilder()
	b, err := json.Marshal(dmExecCtx)
	if err != nil {
		smslog.Infof("%v", err)
	}
	smslog.Infof("contents %s", b)
	msg := msgBuilder.
		WithType(message.SmsMessageHead_CMD_DM_CREAT_REQ).
		WithContent(b).
		Build()
	output, err := proto.Marshal(msg)
	if err != nil {
		//TODO Handle
		smslog.Errorf("send message error %s", err.Error())
		return
	}
	output = append(output, SmsMessageEnd)
	smslog.Infof("send --%s--", output)
	msg1, _ := Parse(string(output))
	smslog.Infof("%v", msg1)
}
