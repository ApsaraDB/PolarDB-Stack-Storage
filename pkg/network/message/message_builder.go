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

import (
	"k8s.io/apimachinery/pkg/util/uuid"
	"polardb-sms/pkg/common"
)

type SmsMessageBuilder struct {
	m *SmsMessage
}

func NewSmsMessageBuilder() *SmsMessageBuilder {
	return &SmsMessageBuilder{
		m: &SmsMessage{
			Head: &SmsMessageHead{},
			Body: &MessageBody{
				ExecStatus: MessageBody_Success,
			},
		}}
}

func (b *SmsMessageBuilder) Build() *SmsMessage {
	b.m.Head.MsgId = string(uuid.NewUUID())
	if b.m.Head.TraceContext == nil {
		b.m.Head.TraceContext = map[string]string{}
	}
	b.m.Head.TraceContext["msgId"] = b.m.Head.MsgId
	b.m.Head.TraceContext["askMsgId"] = b.m.Head.AckMsgId
	return b.m
}

func (b *SmsMessageBuilder) WithType(t SmsMessageHead_SmsMsgType) *SmsMessageBuilder {
	b.m.Head.MsgType = t
	return b
}

func (b *SmsMessageBuilder) WithContent(c []byte) *SmsMessageBuilder {
	b.m.Body.Content = make([]byte, len(c))
	copy(b.m.Body.Content, c)
	return b
}

func (b *SmsMessageBuilder) WithErrMsg(errMsg string) *SmsMessageBuilder {
	b.m.Body.ExecStatus = MessageBody_Fail
	b.m.Body.ErrMsg = errMsg
	return b
}

func (b *SmsMessageBuilder) WithTraceContext(ctx common.TraceContext) *SmsMessageBuilder {
	b.m.Head.TraceContext = ctx
	return b
}

func (b *SmsMessageBuilder) WithAckMsgId(ackMsgId string) *SmsMessageBuilder {
	b.m.Head.AckMsgId = ackMsgId
	return b
}

func NewMessage(t SmsMessageHead_SmsMsgType,
	body interface{},
	ctx common.TraceContext) (*SmsMessage, error) {
	content, err := common.StructToBytes(body)
	if err != nil {
		return nil, err
	}
	return NewSmsMessageBuilder().
		WithContent(content).
		WithType(t).
		WithTraceContext(ctx).Build(), nil
}

func SuccessRespMessage(t SmsMessageHead_SmsMsgType, ackMsgId string, contents []byte) *SmsMessage {
	return NewSmsMessageBuilder().
		WithType(t).
		WithContent(contents).
		WithAckMsgId(ackMsgId).
		Build()
}

func FailRespMessage(t SmsMessageHead_SmsMsgType, ackMsgId, errMsg string) *SmsMessage {
	return NewSmsMessageBuilder().
		WithType(t).
		WithErrMsg(errMsg).
		WithAckMsgId(ackMsgId).
		Build()
}