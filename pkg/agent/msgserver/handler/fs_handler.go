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
	"fmt"
	"sync"

	"polardb-sms/pkg/agent/device/filesystem"
	"polardb-sms/pkg/common"
	"polardb-sms/pkg/network/message"
)

var processingMap = make(map[string]bool)
var processingLock sync.Mutex

func lock(volumeId string) bool {
	processingLock.Lock()
	defer processingLock.Unlock()
	_, ok := processingMap[volumeId]
	if ok {
		return false
	}
	processingMap[volumeId] = true
	return true
}

func unlock(volumeId string) {
	processingLock.Lock()
	defer processingLock.Unlock()
	delete(processingMap, volumeId)
}

type FsExpandReqHandler struct {
}

func (h *FsExpandReqHandler) Handle(msg *message.SmsMessage) *message.SmsMessage {
	var (
		err           error
		fs            filesystem.Filesystem
		expandCommand message.FsExpandCommand
	)

	if err = json.Unmarshal(msg.Body.Content, &expandCommand); err != nil {
		return message.FailRespMessage(message.SmsMessageHead_CMD_EXPAND_FS_RESP, msg.Head.MsgId, err.Error())
	}
	if ok := lock(expandCommand.VolumeId); !ok {
		return message.FailRespMessage(message.SmsMessageHead_CMD_EXPAND_FS_RESP, msg.Head.MsgId, "volume is locked by another processing")
	}
	defer unlock(expandCommand.VolumeId)
	switch expandCommand.FsType {
	case common.Pfs:
		fs = filesystem.NewPfs()
	case common.Ext4:
		fs = filesystem.NewExt4()
	default:
		return message.FailRespMessage(message.SmsMessageHead_CMD_EXPAND_FS_RESP, msg.Head.MsgId, fmt.Sprintf("not found filesystem type - (%s)", expandCommand.FsType))
	}

	if err = fs.ExpandFilesystem(expandCommand.VolumeId, expandCommand.ReqSize, expandCommand.OriginSize); err != nil {
		return message.FailRespMessage(message.SmsMessageHead_CMD_EXPAND_FS_RESP, msg.Head.MsgId, err.Error())
	}
	return message.SuccessRespMessage(message.SmsMessageHead_CMD_EXPAND_FS_RESP, msg.Head.MsgId, nil)
}

type FsFormatReqHandler struct {
}

func (h *FsFormatReqHandler) Handle(msg *message.SmsMessage) *message.SmsMessage {
	var (
		err           error
		fs            filesystem.Filesystem
		formatCommand message.FsFormatCommand
	)

	if err = json.Unmarshal(msg.Body.Content, &formatCommand); err != nil {
		return message.FailRespMessage(message.SmsMessageHead_CMD_FORMAT_FS_RESP, msg.Head.MsgId, err.Error())
	}

	if ok := lock(formatCommand.VolumeId); !ok {
		return message.FailRespMessage(message.SmsMessageHead_CMD_FORMAT_FS_RESP, msg.Head.MsgId, "volume is locked by another processing")
	}
	defer unlock(formatCommand.VolumeId)
	switch formatCommand.FsType {
	case common.Pfs:
		fs = filesystem.NewPfs()
	case common.Ext4:
		fs = filesystem.NewExt4()
	default:
		return message.FailRespMessage(message.SmsMessageHead_CMD_FORMAT_FS_RESP, msg.Head.MsgId, fmt.Sprintf("not found filesystem type - (%s)", formatCommand.FsType))
	}

	if err = fs.FormatFilesystem(formatCommand.VolumeId); err != nil {
		return message.FailRespMessage(message.SmsMessageHead_CMD_FORMAT_FS_RESP, msg.Head.MsgId, err.Error())
	}
	return message.SuccessRespMessage(message.SmsMessageHead_CMD_EXPAND_FS_RESP, msg.Head.MsgId, nil)
}
