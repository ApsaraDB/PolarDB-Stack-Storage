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

package handler

import (
	mpath "polardb-sms/pkg/agent/device/reservation/mpathpersist"
	"polardb-sms/pkg/agent/device/reservation/nvme"
	"polardb-sms/pkg/agent/utils"
	"polardb-sms/pkg/common"
	"polardb-sms/pkg/network/message"
)

type Processor interface {
	Process(cmd *message.PrCmd) (*message.PrCheckCmdResult, error)
}

type PrExecReqHandler struct {
	nvmeExecWrapper  Processor
	mpathExecWrapper Processor
}

func NewPrExecReqHandler() *PrExecReqHandler {
	return &PrExecReqHandler{
		nvme.NewNvmeExecWrapper(),
		mpath.NewPrExecWrapper(),
	}
}

func (h *PrExecReqHandler) Handle(msg *message.SmsMessage) *message.SmsMessage {
	var (
		err    error
		prCmd  = &message.PrCmd{}
		result *message.PrCheckCmdResult
	)

	if err = common.BytesToStruct(msg.Body.Content, prCmd); err != nil {
		return message.FailRespMessage(message.SmsMessageHead_CMD_PR_EXEC_RESP, msg.Head.MsgId, err.Error())
	}

	if h.nvmeExecWrapper != nil && utils.CheckNvmeVolume(prCmd.VolumeId) {
		result, err = h.nvmeExecWrapper.Process(prCmd)
	} else {
		result, err = h.mpathExecWrapper.Process(prCmd)
	}
	if err != nil {
		return message.FailRespMessage(message.SmsMessageHead_CMD_PR_EXEC_RESP, msg.Head.MsgId, err.Error())
	}

	contents, err := common.StructToBytes(result)
	if err != nil {
		return message.FailRespMessage(message.SmsMessageHead_CMD_PR_EXEC_RESP, msg.Head.MsgId, err.Error())
	}
	return message.SuccessRespMessage(message.SmsMessageHead_CMD_PR_EXEC_RESP, msg.Head.MsgId, contents)
}
