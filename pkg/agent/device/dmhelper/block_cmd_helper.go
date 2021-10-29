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

package dmhelper

import (
	"fmt"
	"github.com/pkg/errors"
	"polardb-sms/pkg/agent/utils"
	"polardb-sms/pkg/common"
	"polardb-sms/pkg/device"
	smslog "polardb-sms/pkg/log"
	"strconv"
	"strings"
)

type BlockParam struct {
	SectorSize int `json:"sector_size"`
}

func getBlockDevParam(name string) (*BlockParam, error) {
	var (
		stdout string
		stderr string
		err    error
	)
	cmd := fmt.Sprintf("blockdev --getss /dev/mapper/%s", name)
	stdout, stderr, err = utils.ExecCommand(cmd, utils.CmdDefaultTimeout)
	if err != nil {
		return nil, fmt.Errorf("read dm table failed, stdout: %s, stderr: %s, err: %s", stdout, stderr, err)
	}
	sectorSize, err := strconv.Atoi(strings.Split(stdout, device.NewLineSign)[0])
	if err != nil {
		smslog.Errorf("err: %s", err.Error())
		return nil, err
	}
	return &BlockParam{SectorSize: sectorSize}, nil
}

func GetBlockDevSize(deviceName string) (int64, error) {
	devicePath, err := common.GetDevicePath(deviceName)
	if err != nil {
		return 0, errors.Wrap(err, fmt.Sprintf("failed get DevicePath by deviceName %s", deviceName))
	}
	blockDevCmd := fmt.Sprintf("blockdev --getsize64 %s", devicePath)
	outInfo, stderr, err := utils.ExecCommand(blockDevCmd, utils.CmdDefaultTimeout)
	if err != nil {
		smslog.Errorf("exec command %s stderr %s err %s", blockDevCmd, stderr, err.Error())
		return 0, errors.Wrap(err, stderr)
	}

	var blockDevBytes int64
	blockDevBytes, err = strconv.ParseInt(strings.Trim(outInfo, device.NewLineSign), 10, 64)
	if err != nil {
		return 0, errors.Wrap(err, fmt.Sprintf("failed get blockDevBytes by deviceName %s", deviceName))
	}
	return blockDevBytes, nil
}
