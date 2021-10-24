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
	"encoding/json"
	"polardb-sms/pkg/agent/device/devicemapper"
	"polardb-sms/pkg/agent/device/dmhelper"
	"polardb-sms/pkg/agent/meta"
	"polardb-sms/pkg/device"
	smslog "polardb-sms/pkg/log"
	"polardb-sms/pkg/network/message"
)

// TODO 细化错误处理，所有操作支持幂等
type DmCreateReqHandler struct {
}

/*
   msg.Content: `{"dm_lines":"0 1048576 linear /dev/loop0 8  　\n    1048576 1048576 linear /dev/loop1 8"}`
*/
func (h *DmCreateReqHandler) Handle(msg *message.SmsMessage) *message.SmsMessage {
	var (
		err       error
		dmType    string
		dmCommand message.DmExecCommand
	)

	if err = json.Unmarshal(msg.Body.Content, &dmCommand); err != nil {
		return message.FailRespMessage(message.SmsMessageHead_CMD_DM_CREAT_RESP, msg.Head.MsgId, err.Error())
	}

	switch dmCommand.Device.DeviceType {
	case device.Linear:
		dmType = devicemapper.DmTableLinear
	case device.Striped:
		dmType = devicemapper.DmTableStripe
	case device.Mirror:
		dmType = devicemapper.DmTableMirror
	default:
		return message.FailRespMessage(message.SmsMessageHead_CMD_DM_CREAT_RESP, msg.Head.MsgId, "not support this dm type")
	}
	createDevice, err := dmhelper.QueryDMDevice(dmCommand.DeviceName)
	if err != nil {
		smslog.Errorf("query device %s err %s", dmCommand.DeviceName, err.Error())
	}
	dm := devicemapper.GetDeviceMapper()
	if err == nil && createDevice != nil {
		if err = dm.DmSetupRemove(dmCommand.DeviceName); err != nil {
			return message.FailRespMessage(message.SmsMessageHead_CMD_DM_CREAT_RESP, msg.Head.MsgId, err.Error())
		}
	}
	tableStr, err := dmCommand.Device.GetDmTableString()
	if err != nil {
		return message.FailRespMessage(message.SmsMessageHead_CMD_DM_CREAT_RESP, msg.Head.MsgId, err.Error())
	}
	if err = dm.DmSetupCreate(dmCommand.DeviceName, dmType, tableStr); err != nil {
		return message.FailRespMessage(message.SmsMessageHead_CMD_DM_CREAT_RESP, msg.Head.MsgId, err.Error())
	}

	if err := meta.GetDmStore().Put(&meta.DMTableRecord{
		Name: dmCommand.DeviceName,
		Data: tableStr,
	}); err != nil {
		return message.FailRespMessage(message.SmsMessageHead_CMD_DM_CREAT_RESP, msg.Head.MsgId, err.Error())
	}

	return message.SuccessRespMessage(message.SmsMessageHead_CMD_DM_CREAT_RESP, msg.Head.MsgId, nil)
}

/*
   # echo 0 4194304 thin 253:2 0 | dmsetup load hchen-thin-volumn-001
   # dmsetup resume hchen-thin-volumn-001
*/
type DmExpandReqHandler struct {
}

func (h *DmExpandReqHandler) Handle(msg *message.SmsMessage) *message.SmsMessage {
	var (
		err       error
		dmCommand message.DmExecCommand
	)

	if err = json.Unmarshal(msg.Body.Content, &dmCommand); err != nil {
		return message.FailRespMessage(message.SmsMessageHead_CMD_DM_UPDATE_RESP, msg.Head.MsgId, err.Error())
	}

	var dm = devicemapper.GetDeviceMapper()
	tableStr, err := dmCommand.Device.GetDmTableString()
	if err != nil {
		return message.FailRespMessage(message.SmsMessageHead_CMD_DM_CREAT_RESP, msg.Head.MsgId, err.Error())
	}
	if err = dm.DmSetupLoad(dmCommand.DeviceName, tableStr); err != nil {
		return message.FailRespMessage(message.SmsMessageHead_CMD_DM_UPDATE_RESP, msg.Head.MsgId, err.Error())
	}

	if err = dm.DmSetupResume(dmCommand.DeviceName); err != nil {
		return message.FailRespMessage(message.SmsMessageHead_CMD_DM_UPDATE_RESP, msg.Head.MsgId, err.Error())
	}

	if err = meta.GetDmStore().Put(&meta.DMTableRecord{
		Name: dmCommand.DeviceName,
		Data: tableStr,
	}); err != nil {
		return message.FailRespMessage(message.SmsMessageHead_CMD_DM_UPDATE_RESP, msg.Head.MsgId, err.Error())
	}

	return message.SuccessRespMessage(message.SmsMessageHead_CMD_DM_UPDATE_RESP, msg.Head.MsgId, nil)
}

type DmRemoveReqHandler struct {
}

func (h *DmRemoveReqHandler) Handle(msg *message.SmsMessage) *message.SmsMessage {
	var (
		err       error
		dmCommand message.DmExecCommand
	)

	if err = json.Unmarshal(msg.Body.Content, &dmCommand); err != nil {
		return message.FailRespMessage(message.SmsMessageHead_CMD_DM_DELETE_RESP, msg.Head.MsgId, err.Error())
	}

	deleteDevice, err := dmhelper.QueryDMDevice(dmCommand.DeviceName)
	if err != nil {
		smslog.Errorf("query device %s err %s", dmCommand.DeviceName, err.Error())
	}
	if err == nil && deleteDevice != nil {
		dm := devicemapper.GetDeviceMapper()
		if err = dm.DmSetupRemove(dmCommand.DeviceName); err != nil {
			return message.FailRespMessage(message.SmsMessageHead_CMD_DM_DELETE_RESP, msg.Head.MsgId, err.Error())
		}
	}

	if err := meta.GetDmStore().Delete(dmCommand.DeviceName); err != nil {
		return message.FailRespMessage(message.SmsMessageHead_CMD_DM_DELETE_RESP, msg.Head.MsgId, err.Error())
	}

	return message.SuccessRespMessage(message.SmsMessageHead_CMD_DM_DELETE_RESP, msg.Head.MsgId, nil)
}
