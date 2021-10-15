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
	"go.uber.org/zap/zapcore"
	"polardb-sms/pkg/common"
	smslog "polardb-sms/pkg/log"
	"testing"
)

func TestBatchPrCheckCmd_Bytes(t *testing.T) {
	smslog.InitLogger("/Users/jimmy/", "temp", zapcore.DebugLevel)
	reg, _ := newPrRegisterCmd("abc", common.MultipathVolume, "aba")
	res, _ := newPrReserveCmd("abc", common.MultipathVolume, "aba", EA)
	batchCmd := BatchPrCheckCmd{
		Cmds: []*PrCmd{
			reg, res,
		},
	}
	by, _ := batchCmd.Bytes()
	smslog.Infof(string(by))
	batchCmdRet := &BatchPrCheckCmd{}
	err := common.BytesToStruct(by, batchCmdRet)
	if err != nil {
		smslog.Error(err.Error())
	}
	smslog.Infof("%v, %v",
		batchCmdRet.Cmds[0].CmdParam.(*PrRegisterCmdParam),
		batchCmdRet.Cmds[1].CmdParam.(*PrReserveCmdParam))
}

func TestPrCmd_UnmarshalJSON(t *testing.T) {
	type fields struct {
		CmdType    int
		VolumeType common.LvType
		Name       string
		CmdParam   interface {
		}
	}
	type args struct {
		b []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &PrCmd{
				CmdType:    tt.fields.CmdType,
				VolumeType: tt.fields.VolumeType,
				VolumeId:   tt.fields.Name,
				CmdParam:   tt.fields.CmdParam,
			}
			if err := cmd.UnmarshalJSON(tt.args.b); (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
