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

type RespHandler interface {
	Handle(msg *message.SmsMessage) *message.MessageBody
}

type RespMsgHandleService struct {
	handlers map[message.SmsMessageHead_SmsMsgType]RespHandler
}

func (s *RespMsgHandleService) register(msgType message.SmsMessageHead_SmsMsgType, handler RespHandler) {
	s.handlers[msgType] = handler
}

func (s *RespMsgHandleService) Handle(msg *message.SmsMessage) (*message.MessageBody, error) {
	handler, ok := s.handlers[msg.Head.MsgType]
	if !ok {
		return nil, fmt.Errorf("can not find the handler for msg %v", msg)
	}
	return handler.Handle(msg), nil
}
func NewRespMsgHandleService() *RespMsgHandleService {
	service := &RespMsgHandleService{
		handlers: make(map[message.SmsMessageHead_SmsMsgType]RespHandler),
	}
	//TODO fixme
	service.register(message.SmsMessageHead_CMD_DM_CREAT_RESP, GetDefaultHandler())
	service.register(message.SmsMessageHead_CMD_DM_UPDATE_RESP, GetDefaultHandler())
	service.register(message.SmsMessageHead_CMD_DM_DELETE_RESP, GetDefaultHandler())
	service.register(message.SmsMessageHead_CMD_RESCAN_RESP, GetDefaultHandler())
	service.register(message.SmsMessageHead_CMD_EXPAND_FS_RESP, GetDefaultHandler())
	service.register(message.SmsMessageHead_CMD_FORMAT_FS_RESP, GetDefaultHandler())
	service.register(message.SmsMessageHead_CMD_PR_BATCH_RESP, GetDefaultHandler())
	service.register(message.SmsMessageHead_CMD_PR_EXEC_RESP, GetDefaultHandler())
	service.register(message.SmsMessageHead_CMD_PVC_CREATE_RESP, GetDefaultHandler())
	service.register(message.SmsMessageHead_CMD_PVC_RELEASE_RESP, GetDefaultHandler())
	return service
}
