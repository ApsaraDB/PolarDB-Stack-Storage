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


package nvme

import (
	"fmt"
	"polardb-sms/pkg/agent/device/reservation"
	"polardb-sms/pkg/network/message"
)

type NvmeExecWrapper struct {
	executors map[int]reservation.PrCmdExecutor
}

func NewNvmeExecWrapper() *NvmeExecWrapper {
	wrapper := &NvmeExecWrapper{executors: make(map[int]reservation.PrCmdExecutor)}
	wrapper.executors[message.NvmeRegister] = NewNvmeRegisterExecutor()
	wrapper.executors[message.NvmeReserve] = NewNvmeReserveExecutor()
	wrapper.executors[message.NvmePreempt] = NewNvmePreemptExecutor()
	wrapper.executors[message.NvmeRelease] = NewNvmeReleaseExecutor()
	wrapper.executors[message.NvmeClear] = NewNvmeClearExecutor()
	return wrapper
}

func (w *NvmeExecWrapper) Process(cmd *message.PrCmd) (*message.PrCheckCmdResult, error) {
	executor, exit := w.executors[cmd.CmdType]
	if !exit {
		return nil, fmt.Errorf("do not support cmd type %d", cmd.CmdType)
	}
	return reservation.ExecuteCmd(executor, cmd)
}
