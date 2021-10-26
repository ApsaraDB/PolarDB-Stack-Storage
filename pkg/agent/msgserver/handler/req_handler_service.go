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
	"polardb-sms/pkg/network/message"
)

type ReqHandler interface {
	Handle(msg *message.SmsMessage) *message.SmsMessage
}

type ReqMsgHandlerService interface {
	Process(msg *message.SmsMessage) (*message.SmsMessage, error)
}

type ReqMsgHandleServiceImpl struct {
	handlers map[message.SmsMessageHead_SmsMsgType]ReqHandler
}

func (s *ReqMsgHandleServiceImpl) Register(msgType message.SmsMessageHead_SmsMsgType, handler ReqHandler) {
	s.handlers[msgType] = handler
}

func (s *ReqMsgHandleServiceImpl) Process(msg *message.SmsMessage) (*message.SmsMessage, error) {
	handler, ok := s.handlers[msg.Head.MsgType]
	if !ok {
		return nil, fmt.Errorf("can not find the handler for msg %v", msg)
	}
	return handler.Handle(msg), nil
}

/*
	SmsMessageHead_CMD_PR_EXEC_REQ     SmsMessageHead_SmsMsgType = 10
	SmsMessageHead_CMD_PR_BATCH_REQ    SmsMessageHead_SmsMsgType = 20
	SmsMessageHead_CMD_DM_CREAT_REQ    SmsMessageHead_SmsMsgType = 100
	SmsMessageHead_CMD_DM_UPDATE_REQ   SmsMessageHead_SmsMsgType = 102
	SmsMessageHead_CMD_DM_DELETE_REQ   SmsMessageHead_SmsMsgType = 104
	SmsMessageHead_CMD_RESCAN_REQ      SmsMessageHead_SmsMsgType = 300
	SmsMessageHead_CMD_EXPAND_FS_REQ   SmsMessageHead_SmsMsgType = 400
	SmsMessageHead_CMD_FORMAT_FS_REQ   SmsMessageHead_SmsMsgType = 500
	SmsMessageHead_CMD_LUN_CREATE_REQ  SmsMessageHead_SmsMsgType = 600
	SmsMessageHead_CMD_LUN_EXPAND_REQ  SmsMessageHead_SmsMsgType = 700
*/
func NewReqMsgHandlerService(nodeIp string) ReqMsgHandlerService {
	service := &ReqMsgHandleServiceImpl{
		make(map[message.SmsMessageHead_SmsMsgType]ReqHandler),
	}
	//TODO fixme
	service.Register(message.SmsMessageHead_CMD_PR_EXEC_REQ, NewPrExecReqHandler())
	service.Register(message.SmsMessageHead_CMD_PR_BATCH_REQ, NewBatchPrExecReqHandler())
	service.Register(message.SmsMessageHead_CMD_DM_CREAT_REQ, &DmCreateReqHandler{})
	service.Register(message.SmsMessageHead_CMD_DM_DELETE_REQ, &DmRemoveReqHandler{})
	service.Register(message.SmsMessageHead_CMD_DM_UPDATE_REQ, &DmExpandReqHandler{})
	service.Register(message.SmsMessageHead_CMD_RESCAN_REQ, &ScsiReqHandler{})
	service.Register(message.SmsMessageHead_CMD_EXPAND_FS_REQ, &FsExpandReqHandler{})
	service.Register(message.SmsMessageHead_CMD_FORMAT_FS_REQ, &FsFormatReqHandler{})
	service.Register(message.SmsMessageHead_CMD_PVC_CREATE_REQ, NewPvcCreateHandler(nodeIp))
	service.Register(message.SmsMessageHead_CMD_PVC_RELEASE_REQ, NewPvcReleaseHandler(nodeIp))
	return service
}
