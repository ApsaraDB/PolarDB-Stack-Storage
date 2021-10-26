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
	"polardb-sms/pkg/common"
	"polardb-sms/pkg/device"
)

type PvcReleaseCommand struct {
	Name       string        `json:"name"`
	VolumeType common.LvType `json:"volume_type"`
	VolumeId   string        `json:"volume_id"`
}

type PvcCreateCommand struct {
	Format     bool          `json:"format"`
	VolumeType common.LvType `json:"volume_type"`
	VolumeId   string        `json:"volume_id"`
	FsType     common.FsType `json:"fs_type"`
}

type DmExecCommandType string

const (
	Create DmExecCommandType = "create"
	Delete DmExecCommandType = "delete"
	Expand DmExecCommandType = "expand"
)

//exec dmsetup command
//todo add more data
type DmExecCommand struct {
	CommandType DmExecCommandType    `json:"command_type"`
	DeviceName  string               `json:"device_name"`
	Device      *device.DmDeviceCore `json:"device"`
	Version     string               `json:"version"`
}

//fs expand command
type FsExpandCommand struct {
	VolumeId   string        `json:"volume_id"`
	FsType     common.FsType `json:"fs_type"`
	ReqSize    int64         `json:"req_size"`
	OriginSize int64         `json:"origin_size"`
	VolumeType common.LvType `json:"volume_type"`
}

type LvFormatCommand struct {
	LvName string `json:"lv_name"`
	FsType string `json:"fs_type"`
	Size   int64  `json:"size"`
}

//lun format command
type LunFormatCommand struct {
	LunName string `json:"lv_name"`
	FsType  int    `json:"fs_type"`
	Size    int64  `json:"size"`
}

type FsFormatCommand struct {
	VolumeId   string        `json:"volume_id"`
	VolumeType common.LvType `json:"volume_type"`
	FsType     common.FsType `json:"fs_type"`
	ReqSize    int64         `json:"req_size"`
}

type PrType int

const (
	WE   PrType = 1 //Write Exclusive
	EA   PrType = 3 //Exclusive Access
	WERO PrType = 5 //Write Exclusive, Registrants Only
	EARO PrType = 6 //Exclusive Access, Registrants Only
	WEAR PrType = 7 //Write Exclusive, All Registrants
	EAAR PrType = 8 //Exclusive Access, All Registrants
)
const (
	PrRegister int = iota
	PrReserve
	PrRelease
	PrClear
	PrPreempt
	PrPathNum
	PathCanWrite
	PathCannotWrite
)

const (
	NvmeRegister int = iota
	NvmeReserve
	NvmeRelease
	NvmeClear
	NvmePreempt
	NvmePathNum
)

type PrCmd struct {
	CmdType    int           `json:"cmd_type"`
	VolumeType common.LvType `json:"volume_type"`
	VolumeId   string        `json:"volume_id"`
	CmdParam   interface{}   `json:"cmd_param"`
	//lv format commandParam   []byte `json:"cmd_param"`
}

func (cmd *PrCmd) UnmarshalJSON(b []byte) error {
	type innerPrCmd struct {
		CmdType    int           `json:"cmd_type"`
		VolumeType common.LvType `json:"volume_type"`
		VolumeId   string        `json:"volume_id"`
		CmdParam   interface{}   `json:"cmd_param"`
	}
	tempCmd := &innerPrCmd{}
	err := common.BytesToStruct(b, &tempCmd)
	if err != nil {
		return err
	}
	switch tempCmd.CmdType {
	case PrRegister:
		tempCmd.CmdParam = &PrRegisterCmdParam{}
	case PrReserve:
		tempCmd.CmdParam = &PrReserveCmdParam{}
	case PrPreempt:
		tempCmd.CmdParam = &PrPreemptCmdParam{}
	case PrClear:
		tempCmd.CmdParam = &PrClearCmdParam{}
	case PrRelease:
		tempCmd.CmdParam = &PrReleaseCmdParam{}
	case PrPathNum:
		tempCmd.CmdParam = &PrCheckPathCmdParam{}
	case PathCanWrite:
		tempCmd.CmdParam = &PrCanWriteCmdParam{}
	case PathCannotWrite:
		tempCmd.CmdParam = &PrCannotWriteCmdParam{}
	}
	err = common.BytesToStruct(b, &tempCmd)
	if err != nil {
		return err
	}
	cmd.VolumeId = tempCmd.VolumeId
	cmd.VolumeType = tempCmd.VolumeType
	cmd.CmdType = tempCmd.CmdType
	cmd.CmdParam = tempCmd.CmdParam
	return nil
}

func (cmd *PrCmd) Bytes() ([]byte, error) {
	return common.StructToBytes(cmd)
}

func ParseForPrCheckCmd(bytes []byte) (*PrCmd, error) {
	cmd := &PrCmd{}
	err := common.BytesToStruct(bytes, cmd)
	if err != nil {
		return nil, err
	}
	return cmd, nil
}

type PrRegisterCmdParam struct {
	RegisterKey string `json:"register_key"`
}

type PrReserveCmdParam struct {
	RegisterKey string `json:"register_key"`
	ReserveType PrType `json:"reserve_type"`
}
type PrReleaseCmdParam struct {
	RegisterKey string `json:"register_key"`
	ReserveType PrType `json:"reserve_type"`
}
type PrClearCmdParam struct {
	RegisterKey string `json:"register_key"`
}
type PrPreemptCmdParam struct {
	RegisterKey  string `json:"register_key"`
	ReserveType  PrType `json:"reserve_type"`
	PreemptedKey string `json:"preempted_key"`
}
type PrCheckPathCmdParam struct {
}
type PrCanWriteCmdParam struct {
}
type PrCannotWriteCmdParam struct {
}

func newPrCmd(volumeId string, volumeType common.LvType, cmdType int, param interface{}) *PrCmd {
	cmd := &PrCmd{
		CmdType:    cmdType,
		VolumeType: volumeType,
		VolumeId:   volumeId,
		CmdParam:   param,
	}
	return cmd
}

func newPrRegisterCmd(volumeId string, volumeType common.LvType, registerKey string) (*PrCmd, error) {
	param := &PrRegisterCmdParam{
		RegisterKey: registerKey,
	}
	return newPrCmd(volumeId, volumeType, PrRegister, param), nil
}

func newPrReserveCmd(volumeId string, volumeType common.LvType, registerKey string, reserveType PrType) (*PrCmd, error) {
	param := &PrReserveCmdParam{
		RegisterKey: registerKey,
		ReserveType: reserveType,
	}

	return newPrCmd(volumeId, volumeType, PrReserve, param), nil
}

func newPrReleaseCmd(volumeId string, volumeType common.LvType, registerKey string, reserveType PrType) (*PrCmd, error) {
	param := &PrReleaseCmdParam{
		RegisterKey: registerKey,
		ReserveType: reserveType,
	}

	return newPrCmd(volumeId, volumeType, PrRelease, param), nil
}

func newPrClearCmd(volumeId string, volumeType common.LvType, registerKey string) (*PrCmd, error) {
	param := &PrClearCmdParam{
		RegisterKey: registerKey,
	}

	return newPrCmd(volumeId, volumeType, PrClear, param), nil
}

func NewPrPreemptCmd(volumeId string, volumeType common.LvType, registerKey, preemptedKey string, reserveType PrType) (*PrCmd, error) {
	param := &PrPreemptCmdParam{
		RegisterKey:  registerKey,
		ReserveType:  reserveType,
		PreemptedKey: preemptedKey,
	}

	return newPrCmd(volumeId, volumeType, PrPreempt, param), nil
}

func newPathCanWriteCmd(volumeId string, volumeType common.LvType) (*PrCmd, error) {
	return newPrCmd(volumeId, volumeType, PathCanWrite, nil), nil
}

func newPathCannotWriteCmd(volumeId string, volumeType common.LvType) (*PrCmd, error) {
	return newPrCmd(volumeId, volumeType, PathCannotWrite, nil), nil
}

func NewCheckPathNumCmd(volumeId string, volumeType common.LvType) (*PrCmd, error) {
	return newPrCmd(volumeId, volumeType, PrPathNum, nil), nil
}

type BatchPrCheckCmd struct {
	Cmds []*PrCmd `json:"cmds"`
}

func newBatchPrCmd(cmds []*PrCmd) *BatchPrCheckCmd {
	batchCmd := &BatchPrCheckCmd{
		cmds,
	}
	return batchCmd
}

func (bc *BatchPrCheckCmd) Bytes() ([]byte, error) {
	return common.StructToBytes(bc)
}

func ParseForBatchPrCmd(bytes []byte) (*BatchPrCheckCmd, error) {
	cmd := &BatchPrCheckCmd{}
	err := common.BytesToStruct(bytes, cmd)
	if err != nil {
		return nil, err
	}
	return cmd, nil
}

//todo refactor this code
func NewBatchCmd(volumeId string, volumeType common.LvType, registerKey, preemptedKey string, reserveType PrType, cmdTypes []int) (*BatchPrCheckCmd, error) {
	var cmds = make([]*PrCmd, 0)
	for _, cmdType := range cmdTypes {
		switch cmdType {
		case PrRegister:
			cmd, err := newPrRegisterCmd(volumeId, volumeType, registerKey)
			if err != nil {
				return nil, err
			}
			cmds = append(cmds, cmd)
		case PrReserve:
			cmd, err := newPrReserveCmd(volumeId, volumeType, registerKey, reserveType)
			if err != nil {
				return nil, err
			}
			cmds = append(cmds, cmd)
		case PrRelease:
			cmd, err := newPrReleaseCmd(volumeId, volumeType, registerKey, reserveType)
			if err != nil {
				return nil, err
			}
			cmds = append(cmds, cmd)
		case PrClear:
			cmd, err := newPrClearCmd(volumeId, volumeType, registerKey)
			if err != nil {
				return nil, err
			}
			cmds = append(cmds, cmd)
		case PrPathNum:
			cmd, err := NewCheckPathNumCmd(volumeId, volumeType)
			if err != nil {
				return nil, err
			}
			cmds = append(cmds, cmd)
		case PathCanWrite:
			cmd, err := newPathCanWriteCmd(volumeId, volumeType)
			if err != nil {
				return nil, err
			}
			cmds = append(cmds, cmd)
		case PrPreempt:
			cmd, err := NewPrPreemptCmd(volumeId, volumeType, registerKey, preemptedKey, reserveType)
			if err != nil {
				return nil, err
			}
			cmds = append(cmds, cmd)
		}
	}

	return newBatchPrCmd(cmds), nil
}
