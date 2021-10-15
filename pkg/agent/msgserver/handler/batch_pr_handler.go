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
	"fmt"
	"polardb-sms/pkg/agent/device/dmhelper"
	"polardb-sms/pkg/agent/device/reservation/mpathpersist"
	"polardb-sms/pkg/agent/device/reservation/nvme"
	"polardb-sms/pkg/agent/utils"
	"polardb-sms/pkg/common"
	smslog "polardb-sms/pkg/log"
	"polardb-sms/pkg/network/message"
	"time"
)

type BatchPrExecReqHandler struct {
	nvmeExecWrapper  Processor
	prExecWrapper Processor
}

func NewBatchPrExecReqHandler() *BatchPrExecReqHandler {
	return &BatchPrExecReqHandler{
		nvmeExecWrapper: nvme.NewNvmeExecWrapper(),
		prExecWrapper: mpathpersist.NewPrExecWrapper(),
	}
}

func (h *BatchPrExecReqHandler) execPrCmdForLv(prCmd *message.PrCmd) ([]*message.PrCheckCmdResult, error) {
	switch prCmd.VolumeType {
	case common.MultipathVolume:
		var ret *message.PrCheckCmdResult
		var err error
		if h.nvmeExecWrapper != nil && utils.CheckNvmeVolumeStartWith3(prCmd.VolumeId) {
			ret, err = h.nvmeExecWrapper.Process(prCmd)
		} else {
			ret, err = h.prExecWrapper.Process(prCmd)
		}
		if err != nil {
			return nil, err
		}
		return []*message.PrCheckCmdResult{ret}, nil
	case common.DmStripVolume, common.DmLinearVolume:
		lvDevice, err := dmhelper.QueryDMDevice(prCmd.VolumeId)
		if err != nil {
			retErr := fmt.Errorf("QueryDMDevice %s err %s", prCmd.VolumeId, err.Error())
			smslog.Error(retErr.Error())
			return nil, retErr
		}
		rets := make([]*message.PrCheckCmdResult, 0)
		for _, deviceId := range lvDevice.Children() {
			newPrExecCmd := &message.PrCmd{
				CmdType:    prCmd.CmdType,
				VolumeType: prCmd.VolumeType,
				VolumeId:   deviceId,
				CmdParam:   prCmd.CmdParam,
			}
			var prCheckResult *message.PrCheckCmdResult
			var err error
			if h.nvmeExecWrapper != nil && utils.CheckNvmeVolumeStartWith3(prCmd.VolumeId) {
				prCheckResult, err = h.nvmeExecWrapper.Process(newPrExecCmd)
			} else {
				prCheckResult, err = h.prExecWrapper.Process(newPrExecCmd)
			}
			if err != nil {
				smslog.Errorf("execPrCmdForLv cmd %v err %s", newPrExecCmd, err.Error())
				return nil, err
			}
			rets = append(rets, prCheckResult)
		}
		return rets, nil
	default:
		return nil, fmt.Errorf("not support volume type %s ", prCmd.VolumeType)
	}
}

func (h *BatchPrExecReqHandler) Handle(msg *message.SmsMessage) *message.SmsMessage {
	var prCmds = &message.BatchPrCheckCmd{}
	err := common.BytesToStruct(msg.Body.Content, prCmds)
	if err != nil {
		return message.FailRespMessage(message.SmsMessageHead_CMD_PR_BATCH_RESP, msg.Head.MsgId, err.Error())
	}
	var results = &message.PrBatchCheckCmdResult{
		Results: make([]*message.PrCheckCmdResult, 0),
	}

	err = common.RunWithRetry(3, 1*time.Second, func(retryTimes int) error {
		var tempResults []*message.PrCheckCmdResult
		for _, prCmd := range prCmds.Cmds {
			smslog.Infof("exec prCmd %v", *prCmd)
			var result *message.PrCheckCmdResult
			var err error
			if h.nvmeExecWrapper != nil && utils.CheckNvmeVolumeStartWith3(prCmd.VolumeId) {
				result, err = h.nvmeExecWrapper.Process(prCmd)
			} else {
				result, err = h.prExecWrapper.Process(prCmd)
			}
			if err != nil {
				return err
			}
			tempResults = append(tempResults, result)
		}
		results.Results = tempResults
		return nil
	})

	if err != nil {
		return message.FailRespMessage(message.SmsMessageHead_CMD_PR_BATCH_RESP, msg.Head.MsgId, err.Error())
	}

	contents, err := common.StructToBytes(results)
	if err != nil {
		return message.FailRespMessage(message.SmsMessageHead_CMD_PR_BATCH_RESP, msg.Head.MsgId, err.Error())
	}
	return message.SuccessRespMessage(message.SmsMessageHead_CMD_PR_BATCH_RESP, msg.Head.MsgId, contents)
}
