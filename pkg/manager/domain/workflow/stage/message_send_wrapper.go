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


package stage

import (
	"fmt"
	smslog "polardb-sms/pkg/log"
	"polardb-sms/pkg/manager/config"
	"polardb-sms/pkg/manager/msgserver"
	"polardb-sms/pkg/network/message"
	"time"
)

func sendAndWait(msg *message.SmsMessage, toId string, timeout int64) *StageExecResult {
	smslog.Debugf("sendAndWait to %s timeout %d sec", toId, timeout)
	ch := make(chan interface{})
	defer close(ch)
	msgserver.MsgServer.SendMessageTo(toId, msg, ch)
	select {
	case result := <-ch:
		return FromMessageBody(result.(*message.MessageBody))
	case <-time.After(time.Duration(timeout) * time.Second):
		return StageExecFail(fmt.Sprintf("Timeout when agent process msg [%v], toId %s", msg, toId))
	}
}

func sendToAllSequential(msg *message.SmsMessage,
	timeout int64, minSuc int) *StageExecResult {
	var rets = make(chan *StageExecResult)
	defer close(rets)
	var sucCnt = 0
	var returnVal *StageExecResult

	for nodeName, _ := range config.GetAvailableNodes() {
		go func() {
			defer smslog.LogPanic()
			var toMsg = &message.SmsMessage{
				Head: &message.SmsMessageHead{
					MsgType:      msg.Head.MsgType,
					MsgId:        fmt.Sprintf("%s#%s", nodeName, msg.Head.MsgId),
					MsgLen:       msg.Head.MsgLen,
					AckMsgId:     "",
					TraceContext: msg.Head.TraceContext,
				},
				Body: msg.Body,
			}
			ret := sendAndWait(toMsg, nodeName, timeout)
			rets <- ret
		}()
		select {
		case ret := <-rets:
			if ret.ExecStatus == StageSuccess {
				sucCnt += 1
				returnVal = ret
			} else {
				smslog.Debugf("sendToAllSequential: stage exec err %s", ret.ErrMsg)
				returnVal = ret
			}
		case <-time.After(time.Duration(timeout) * time.Second):
			returnVal = StageExecFail("time out")
		}
		if sucCnt >= minSuc {
			break
		}
	}
	return returnVal
}

func sendToAllParallel(msg *message.SmsMessage,
	timeout int64, minSuc int) *StageExecResult {
	var rets = make(chan *StageExecResult)
	defer close(rets)
	for nodeName, _ := range config.GetAvailableNodes() {
		nodeName := nodeName
		go func() {
			defer smslog.LogPanic()
			var toMsg = &message.SmsMessage{
				Head: &message.SmsMessageHead{
					MsgType:      msg.Head.MsgType,
					MsgId:        fmt.Sprintf("%s#%s", nodeName, msg.Head.MsgId),
					MsgLen:       msg.Head.MsgLen,
					AckMsgId:     "",
					TraceContext: msg.Head.TraceContext,
				},
				Body: msg.Body,
			}
			ret := sendAndWait(toMsg, nodeName, timeout)
			rets <- ret
		}()
	}

	var sucCnt = 0
	var returnVal *StageExecResult
	for {
		select {
		case ret := <-rets:
			if ret.ExecStatus == StageSuccess {
				sucCnt += 1
				returnVal = ret
			} else {
				smslog.Debugf("sendToAllParallel: stage exec err %s", ret.ErrMsg)
			}
		case <-time.After(time.Duration(timeout) * time.Second):
			returnVal = StageExecFail("time out")
			return returnVal
		}
		if sucCnt >= minSuc {
			break
		}
	}
	return returnVal
}
