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
	"polardb-sms/pkg/common"
	smslog "polardb-sms/pkg/log"
	"polardb-sms/pkg/manager/config"
	"polardb-sms/pkg/network/message"
)

func NewPrLockStage(node config.Node, nodeIp, wwid, currentPrNodeIp string, volumeType common.LvType) (*PrBatchStageRunner, error) {
	var (
		cmdTypes []int
		preemptedKey string
	)
	if currentPrNodeIp != "" {
		cmdTypes = []int{message.PrPreempt}
		preemptedKey = common.IpV4ToPrKey(currentPrNodeIp)
	} else {
		cmdTypes = []int{message.PrRegister, message.PrReserve}
	}

	if volumeType.ToVolumeClass() == common.LvClass {
		cmdTypes = append(cmdTypes, message.PathCanWrite)
	}
	//name string, volumeType common.LvType, registerKey, preemptedKey string, reserveType PrType
	cmd, err := message.NewBatchCmd(wwid, volumeType, common.IpV4ToPrKey(nodeIp), preemptedKey, message.WEAR, cmdTypes)
	if err != nil {
		return nil, err
	}
	return NewPrBatchStage(cmd, node), nil
}

func NewCheckPathNumCmdStage(node config.Node, volumeId string, volumeType common.LvType) (*PrStageRunner, error) {
	cmd, err := message.NewCheckPathNumCmd(volumeId, volumeType)
	if err != nil {
		return nil, err
	}
	return NewPrStage(cmd, node), nil
}

func NewRegisterAndReserveCmdStage(node config.Node, volumeId string, volumeType common.LvType, registerKey string, reserveType message.PrType) (*PrBatchStageRunner, error) {
	batchCmd, err := message.NewBatchCmd(volumeId, volumeType, registerKey, "", reserveType,
		[]int{message.PrRegister, message.PrReserve, message.PathCanWrite})
	if err != nil {
		return nil, err
	}
	return NewPrBatchStage(batchCmd, node), nil
}

func NewReleaseAndClearCmdStage(node config.Node, volumeId string, volumeType common.LvType, registerKey string, reserveType message.PrType) (*PrBatchStageRunner, error) {
	batchCmd, err := message.NewBatchCmd(volumeId, volumeType, registerKey, "", reserveType,
		[]int{message.PrRegister, message.PrReserve, message.PrClear})
	if err != nil {
		return nil, err
	}
	return NewPrBatchStage(batchCmd, node), nil
}

func NewRegAndPreemptCmdStage(node config.Node, volumeId string, volumeType common.LvType, registerKey, preemptedKey string, reserveType message.PrType) (*PrBatchStageRunner, error) {
	batchCmd, err := message.NewBatchCmd(volumeId, volumeType, registerKey, preemptedKey, reserveType,
		[]int{message.PrRegister, message.PrPreempt, message.PathCanWrite})
	if err != nil {
		return nil, err
	}
	return NewPrBatchStage(batchCmd, node), nil
}

func NewPreemptAndClearCmdStage(node config.Node, volumeId string, volumeType common.LvType, registerKey, preemptedKey string, reserveType message.PrType) (*PrBatchStageRunner, error) {
	batchCmd, err := message.NewBatchCmd(volumeId, volumeType, registerKey, preemptedKey, reserveType,
		[]int{message.PrRegister, message.PrPreempt, message.PathCanWrite, message.PrClear})
	if err != nil {
		return nil, err
	}
	return NewPrBatchStage(batchCmd, node), nil
}

type PrStageRunner struct {
	*Stage
	TargetNode config.Node
}

func (s *PrStageRunner) Run(ctx common.TraceContext) *StageExecResult {
	msg, err := message.NewMessage(message.SmsMessageHead_CMD_PR_EXEC_REQ, s.Content, ctx)
	if err != nil {
		return StageExecFail(err.Error())
	}
	ret := sendAndWait(msg, s.TargetNode.Name, 5)
	s.Result = ret
	smslog.Infof("finish run pr cmd %v, result %s", s.Content, string(s.Result.Content))
	return ret
}

func (s *PrStageRunner) Rollback(ctx common.TraceContext) *StageExecResult {
	return StageExecFail(fmt.Errorf("umimplement fs expand rollback").Error())
}

func NewPrStage(cmd *message.PrCmd,
	execNode config.Node) *PrStageRunner {
	return &PrStageRunner{
		Stage: &Stage{
			Content:   cmd,
			SType:     PrStage,
			StartTime: 0,
			Result:    nil,
		},
		TargetNode: execNode,
	}
}

type PrStageConstructor struct {
}

func (c *PrStageConstructor) Construct() interface{} {
	return &PrStageRunner{
		Stage: &Stage{
			Content: &message.PrCmd{},
		},
		TargetNode: config.Node{},
	}
}

type PrBatchStageRunner struct {
	*Stage
	TargetNode config.Node
}

func (s *PrBatchStageRunner) Run(ctx common.TraceContext) *StageExecResult {
	msg, err := message.NewMessage(message.SmsMessageHead_CMD_PR_BATCH_REQ, s.Content, ctx)
	if err != nil {
		return StageExecFail(err.Error())
	}
	ret := sendAndWait(msg, s.TargetNode.Name, 10)
	s.Result = ret
	return ret
}

func (s *PrBatchStageRunner) Rollback(ctx common.TraceContext) *StageExecResult {
	return StageExecFail(fmt.Errorf("umimplement fs expand rollback").Error())
}

func NewPrBatchStage(batchPrCmd *message.BatchPrCheckCmd,
	execNode config.Node) *PrBatchStageRunner {
	return &PrBatchStageRunner{
		Stage: &Stage{
			Content:   batchPrCmd,
			SType:     PrBatchStage,
			StartTime: 0,
			Result:    nil,
		},
		TargetNode: execNode,
	}
}

type PrBatchStageConstructor struct {
}

func (c *PrBatchStageConstructor) Construct() interface{} {
	return &PrBatchStageRunner{
		Stage: &Stage{
			Content: &message.BatchPrCheckCmd{
				Cmds: []*message.PrCmd{},
			},
		},
	}
}
