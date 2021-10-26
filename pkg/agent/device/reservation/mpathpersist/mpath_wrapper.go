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

package mpathpersist

import (
	"fmt"

	"polardb-sms/pkg/agent/device/reservation"
	"polardb-sms/pkg/network/message"
)

const (
	BaseDir = "/dev/mapper"
)

type PrExecWrapper struct {
	executors map[int]reservation.PrCmdExecutor
}

func NewPrExecWrapper() *PrExecWrapper {
	wrapper := &PrExecWrapper{executors: make(map[int]reservation.PrCmdExecutor)}
	wrapper.executors[message.PrRegister] = NewPrRegisterExecutor()
	wrapper.executors[message.PrReserve] = NewPrReserveExecutor()
	wrapper.executors[message.PrRelease] = NewPrReleaseExecutor()
	wrapper.executors[message.PrClear] = NewPrClearExecutor()
	wrapper.executors[message.PrPreempt] = NewPrPreemptExecutor()
	wrapper.executors[message.PrPathNum] = NewPrPathNumExecutor()
	wrapper.executors[message.PathCanWrite] = NewPrPathCanWriteExecutor()
	wrapper.executors[message.PathCannotWrite] = NewPrPathCannotWriteExecutor()
	return wrapper
}

func (w *PrExecWrapper) Process(cmd *message.PrCmd) (*message.PrCheckCmdResult, error) {
	executor, exit := w.executors[cmd.CmdType]
	if !exit {
		return nil, fmt.Errorf("do not support cmd type %d", cmd.CmdType)
	}
	return reservation.ExecuteCmd(executor, cmd)
}
