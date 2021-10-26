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
	"polardb-sms/pkg/agent/utils"
	"polardb-sms/pkg/device"
	smslog "polardb-sms/pkg/log"
	"strconv"
	"strings"
	"time"
)

/*
$ df -T /dev/mapper/hchen-thin-volumn-001
Filesystem                        Type 1K-blocks  Used Available Use% Mounted on
/dev/mapper/hchen-thin-volumn-001 ext4    999320  2568    927940   1% /mnt/base
*/
//Todo just Pfs
type FileSystemParam struct {
	Filesystem string
	Type       string
	BlockSize  int64
	Used       int64
	Available  int64
	Use        string
	MountOn    string
}

func getExt4Param(deviceName string) (*FileSystemParam, error) {
	cmd := fmt.Sprintf("df -T /dev/mapper/%s", deviceName)
	stdout, stderr, err := utils.ExecCommand(cmd, utils.CmdDefaultTimeout)
	if err != nil || stderr != "" {
		return nil, fmt.Errorf("print file system type failed: %v", err)
	}

	var df FileSystemParam
	for _, row := range strings.Split(stdout, device.NewLineSign) {
		if strings.Contains(row, "1K-blocks") || row == "" {
			continue
		}
		fields := strings.Fields(row)
		df.Filesystem = fields[0]
		df.Type = fields[1]
		blocks, err := strconv.ParseInt(fields[2], 10, 64)
		if err != nil {
			return nil, err
		}
		df.BlockSize = blocks * 1024
		used, err := strconv.ParseInt(fields[3], 10, 64)
		if err != nil {
			return nil, err
		}
		df.Used = used
		available, err := strconv.ParseInt(fields[4], 10, 64)
		if err != nil {
			return nil, err
		}
		df.Available = available
		df.Use = fields[5]
		df.MountOn = fields[6]
	}
	return &df, nil
}

func getFileSystemParam(deviceName string) (*FileSystemParam, error) {
	if isPfs(deviceName) {
		var df = &FileSystemParam{
			Filesystem: "pfs",
			Type:       "pfs",
			BlockSize:  0,
			Used:       0,
			Available:  0,
			Use:        "",
			MountOn:    "",
		}
		var err error
		df.BlockSize, err = pfsCapacity(deviceName)
		if err != nil {
			smslog.Debugf("getFileSystemParam %s pfsCapacity err %s", deviceName, err.Error())
		}
		df.Used, err = pfsUsed(deviceName)
		if err != nil {
			smslog.Debugf("getFileSystemParam %s pfsUsed err %s", deviceName, err.Error())
		}
		return df, nil
	}
	return nil, fmt.Errorf("not support non pfs")
}

func isPfs(deviceName string) bool {
	devicePath := fmt.Sprintf("/dev/mapper/%s", deviceName)
	checkCmd := fmt.Sprintf("xxd -l 16 %s | grep JCSFP -B 1", devicePath)
	stdout, stderr, err := utils.ExecCommand(checkCmd, utils.CmdDefaultTimeout)
	if err != nil {
		smslog.Debugf("xxd read %s failed, stdout: %s, stderr: %s, err: %s", deviceName, stdout, stderr, err)
		return false
	}
	if len(stdout) >= 1 {
		return true
	}
	return false
}

func pfsCapacity(deviceName string) (int64, error) {
	checkCmd := fmt.Sprintf("pfs -C disk info mapper_%s | grep nchild=", deviceName)
	stdout, stderr, err := utils.ExecCommand(checkCmd, 20*time.Second)
	if err != nil {
		smslog.Debugf("pfs info %s failed, stdout: %s, stderr: %s, err: %s", deviceName, stdout, stderr, err)
		return 0, err
	}
	for _, row := range strings.Split(stdout, device.NewLineSign) {
		fields := strings.Split(row, ",")
		for _, field := range fields {
			if strings.Contains(field, "nchild=") {
				field := strings.TrimSpace(field)
				sizeIn10Gb, err := strconv.ParseInt(strings.TrimPrefix(field,
					"nchild="), 10, 64)
				if err != nil {
					return 0, err
				}
				return sizeIn10Gb * 1024 * 1024 * 1024 * 10, nil
			}
		}
	}
	return 0, fmt.Errorf("failed parse pfsCapcity %s , stdout %s", checkCmd, stdout)
}

func pfsUsed(deviceName string) (int64, error) {
	blockDevBytes, err := GetBlockDevSize(deviceName)
	if err != nil {
		return 0, err
	}
	reqSizeIn100GiB := blockDevBytes / (100 * 1024 * 1024 * 1024)
	checkCmd := fmt.Sprintf("pfs -C disk du  -d 1 /mapper_%s/ | grep /mapper_%s/$",
		deviceName, deviceName)
	stdout, stderr, err := utils.ExecCommand(checkCmd, time.Duration(20+10*reqSizeIn100GiB)*time.Second)
	if err != nil {
		smslog.Debugf("pfs du %s failed, stdout: %s, stderr: %s, err: %s", deviceName, stdout, stderr, err)
		return 0, err
	}
	fields := strings.Fields(stdout)
	usedKB, err := strconv.ParseInt(fields[0], 10, 64)
	if err != nil {
		return 0, err
	}
	return usedKB * 1000, nil
}
