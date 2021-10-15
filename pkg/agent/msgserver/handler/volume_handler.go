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
	"polardb-sms/pkg/agent/device/devicemapper"
	"polardb-sms/pkg/network/message"
)

type ScsiReqHandler struct {
}

/*
  Rescan Entire Host SCSI Bus
  ## echo "- - -" > /sys/class/scsi_host/$HOST/scan
  Rescan Specific SCSI Device
  ## echo 1 > /sys/block/$DEVICE/device/rescan
*/
func (h *ScsiReqHandler) Handle(msg *message.SmsMessage) *message.SmsMessage {
	var (
		err error
		dm  = devicemapper.GetDeviceMapper()
	)
	if err = dm.ScanDeviceFcHost(); err != nil {
		return message.FailRespMessage(message.SmsMessageHead_CMD_RESCAN_RESP, msg.Head.MsgId, err.Error())
	}
	if err = dm.ScanDeviceIscsiHost(); err != nil {
		return message.FailRespMessage(message.SmsMessageHead_CMD_RESCAN_RESP, msg.Head.MsgId, err.Error())
	}
	if err = dm.ScsiDeviceRescan(); err != nil {
		return message.FailRespMessage(message.SmsMessageHead_CMD_RESCAN_RESP, msg.Head.MsgId, err.Error())
	}
	return message.SuccessRespMessage(message.SmsMessageHead_CMD_RESCAN_RESP, msg.Head.MsgId, nil)
}
