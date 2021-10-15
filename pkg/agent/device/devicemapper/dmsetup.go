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


package devicemapper

import (
	"fmt"
	"github.com/pkg/errors"
	"polardb-sms/pkg/agent/utils"
	"polardb-sms/pkg/common"
	smslog "polardb-sms/pkg/log"
)

const (
	DMLines = "dm-lines"

	DmTableLinear = "linear"
	DmTableStripe = "stripe"
	DmTableMirror = "mirror"
)

type dmSetup struct {
	dmTable     string
	dmTableType string
}

/*
   dmsetup operate 函数簇
   ## DmSetupCreate(deviceName, dmTableType, dmLines string) error
   ## DmSetupMessage(deviceName string) error
   ## DmSetupSuspend(deviceName string) error
   ## DmSetupResume(deviceName string) error
   ## DmSetupLoad(deviceName, dmLines string) error
   ## DmSetupReload(deviceName string) error
   ## DmSetupRemove(deviceName string) error
*/
func (d *dmSetup) DmSetupCreate(deviceName, dmTableType, dmLines string) error {
	var dmFile = fmt.Sprintf("/var/lib/sms-agent/%s", deviceName)
	err := common.WriteToFile(dmFile, dmLines)
	if err != nil {
		smslog.Error("write to file $s err %s", dmFile, err.Error())
		return err
	}
	dmCreateCmd := fmt.Sprintf("dmsetup create %s %s", deviceName, dmFile)
	_, stderr, err := utils.ExecCommand(dmCreateCmd, utils.CmdDefaultTimeout)
	if err != nil {
		smslog.Errorf("exec command %s stderr %s err %s", dmCreateCmd, stderr, err.Error())
		return errors.Wrap(err, stderr)
	}
	smslog.Infof("successfully exec dmsetup create %s as type - (%s)", deviceName, dmTableType)
	return nil
}

func (d *dmSetup) DmSetupSuspend(deviceName string) error {
	dmSuspendCmd := fmt.Sprintf("dmsetup suspend %s", deviceName)
	_, stderr, err := utils.ExecCommand(dmSuspendCmd, utils.CmdDefaultTimeout)
	if err != nil {
		smslog.Errorf("exec command %s stderr %s err %s", dmSuspendCmd, stderr, err.Error())
		return errors.Wrap(err, stderr)
	}
	smslog.Infof("successfully exec dmsetup suspend %s", deviceName)

	return nil
}

func (d *dmSetup) DmSetupResume(deviceName string) error {
	dmResumeCmd := fmt.Sprintf("dmsetup resume %s", deviceName)
	_, stderr, err := utils.ExecCommand(dmResumeCmd, utils.CmdDefaultTimeout)
	if err != nil {
		smslog.Errorf("exec command %s stderr %s err %s", dmResumeCmd, stderr, err.Error())
		return errors.Wrap(err, stderr)
	}
	smslog.Infof("successfully exec dmsetup resume %s", deviceName)

	return nil
}

func (d *dmSetup) DmSetupLoad(deviceName, dmLines string) error {
	dmLoadCmd := fmt.Sprintf("echo '%s' | dmsetup load %s", dmLines, deviceName)
	_, stderr, err := utils.ExecCommand(dmLoadCmd, utils.CmdDefaultTimeout)
	if err != nil {
		smslog.Errorf("exec command %s stderr %s err %s", dmLoadCmd, stderr, err.Error())
		return errors.Wrap(err, stderr)
	}
	smslog.Infof("successfully exec dmsetup load %s", deviceName)

	return nil
}

func (d *dmSetup) DmSetupReload(deviceName, dmTableType, dmLines string) error {
	var dmFile = fmt.Sprintf("/var/lib/sms-agent/%s", deviceName)
	err := common.WriteToFile(dmFile, dmLines)
	if err != nil {
		smslog.Error("write to file $s err %s", dmFile, err.Error())
		return err
	}
	dmReloadCmd := fmt.Sprintf("dmsetup reload %s %s", deviceName, dmFile)
	_, stderr, err := utils.ExecCommand(dmReloadCmd, utils.CmdDefaultTimeout)
	if err != nil {
		smslog.Errorf("exec command %s stderr %s err %s", dmReloadCmd, stderr, err.Error())
		return errors.Wrap(err, stderr)
	}
	smslog.Infof("successfully exec dmsetup reload %s", deviceName)

	return nil
}

func (d *dmSetup) DmSetupRemove(deviceName string) error {
	dmRemoveCmd := fmt.Sprintf("dmsetup remove -f  %s", deviceName)
	_, stderr, err := utils.ExecCommand(dmRemoveCmd, utils.CmdDefaultTimeout)
	if err != nil {
		smslog.Errorf("exec command %s stderr %s err %s", dmRemoveCmd, stderr, err.Error())
		return errors.Wrap(err, stderr)
	}
	smslog.Infof("successfully exec dmsetup remove %s", deviceName)

	return nil
}

func newDmSetup() *dmSetup {
	return &dmSetup{}
}
