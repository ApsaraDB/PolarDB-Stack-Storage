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


package message

import "polardb-sms/pkg/common"

type ResultCode int

func (x *MessageBody) IsSuccess() bool {
	return x != nil && (x.ExecStatus == MessageBody_Success)
}

func ExecFail(msg string) MessageBody {
	return MessageBody{
		ExecStatus: MessageBody_Fail,
		ErrMsg:     msg,
	}
}

func ExecSuccess(contents []byte) MessageBody {
	return MessageBody{
		ExecStatus: MessageBody_Success,
		Content:    contents,
	}
}


type PrCheckCmdResult struct {
	CheckType   int           `json:"check_type"`
	CheckResult int           `json:"check_result"`
	VolumeType  common.LvType `json:"volume_type"`
	Name        string        `json:"name"`
}

type PrBatchCheckCmdResult struct {
	Results []*PrCheckCmdResult `json:"results"`
}
