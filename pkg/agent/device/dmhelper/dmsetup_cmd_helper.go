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
	"polardb-sms/pkg/device"
	smslog "polardb-sms/pkg/log"
	"strings"

	"polardb-sms/pkg/agent/utils"
)

type DmParam struct {
	Name   string                `json:"name"`
	DmType *device.DmDeviceType  `json:"dm_type"`
	Flag   string                `json:"flag"`
	Minor  string                `json:"minor"`
	Items  []*device.DmTableItem `json:"items"`
}

type ReadOptions struct {
	targetType  device.DmDeviceType
	deviceNames []string
}

func getDmParam(name string) (*DmParam, error) {
	dmParamLine, err := execDmSetupCommand(&ReadOptions{deviceNames: []string{name}})
	if err != nil {
		return nil, err
	}
	return parseDmParamByConciseLine(dmParamLine)
}

func getDmParams() ([]*DmParam, error) {
	return getDmParamsByType("")
}

func getDmParamByType(name string, dmType device.DmDeviceType) (*DmParam, error) {
	dmParamLine, err := execDmSetupCommand(
		&ReadOptions{
			deviceNames: []string{name},
			targetType:  dmType,
		})
	if err != nil {
		return nil, err
	}
	return parseDmParamByConciseLine(dmParamLine)
}

func getDmParamsByType(dmType device.DmDeviceType) ([]*DmParam, error) {
	dmParamLine, err := execDmSetupCommand(&ReadOptions{targetType: dmType})
	if err != nil {
		return nil, err
	}
	lines := strings.Split(dmParamLine, device.SemicolonSign)
	var dmParams []*DmParam
	for _, line := range lines {
		dmParam, err := parseDmParamByConciseLine(line)
		if err != nil {
			smslog.Warnf("failed to parse line [%s]: %v", line, err)
			continue
		}
		dmParams = append(dmParams, dmParam)
	}
	return dmParams, nil
}

func execDmSetupCommand(options *ReadOptions) (string, error) {
	var (
		stdout string
		stderr string
		err    error
	)

	cmd := fmt.Sprintf("dmsetup table --concise")
	if options.targetType != "" {
		cmd = fmt.Sprintf("%s --target %s", cmd, options.targetType)
	}
	if len(options.deviceNames) != 0 {
		cmd = fmt.Sprintf("%s %s", cmd, strings.Join(options.deviceNames, " "))
	}
	stdout, stderr, err = utils.ExecCommand(cmd, utils.CmdDefaultTimeout)
	if err != nil && !strings.Contains(stderr, "No such device or address") {
		return "", fmt.Errorf("read dm table failed, stdout: %s, stderr: %s, err: %s", stdout, stderr, err)
	}
	if err == nil && strings.Contains(stdout, "No device found") {
		return "", fmt.Errorf("dm table not found devices")
	}
	return stdout, err
}

func parseDmParamByConciseLine(line string) (*DmParam, error) {
	items := strings.Split(line, device.CommaSign)
	dmParam := &DmParam{
		Name:  items[0],
		Flag:  items[3],
		Minor: items[2],
		Items: make([]*device.DmTableItem, 0),
	}
	targets := items[4:]
	if err := fillDmParam(dmParam, targets); err != nil {
		return nil, err
	}
	return dmParam, nil
}

func parseDmParamByLines(name string, lines []string) (*DmParam, error) {
	dmParam := &DmParam{
		Name:  name,
		Items: make([]*device.DmTableItem, 0),
	}

	err := fillDmParam(dmParam, lines)
	if err != nil {
		return nil, err
	}
	return dmParam, nil
}

func fillDmParam(dmParam *DmParam, lines []string) error {
	if len(lines) == 0 {
		return fmt.Errorf("failed parse deivce %s, contents is empty", dmParam.Name)
	}
	var dmDeviceType *device.DmDeviceType
	var items = make([]*device.DmTableItem, 0)
	for _, line := range lines {
		item, tmpType, err := device.ParseFromLine(line)
		if err != nil {
			return err
		}
		if dmDeviceType == nil {
			dmDeviceType = tmpType
		}
		if *dmDeviceType != *tmpType {
			return fmt.Errorf("device mapper type not equal ,previous type %s, current type %s", *dmDeviceType, *tmpType)
		}
		items = append(items, item)
	}
	if dmDeviceType == nil {
		return fmt.Errorf("can not find device type from contents %v", lines)
	}
	dmParam.DmType = dmDeviceType
	dmParam.Items = items
	return nil
}
