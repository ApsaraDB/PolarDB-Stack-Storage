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


package common

import (
	"fmt"
	"os"
	smslog "polardb-sms/pkg/log"
	"strings"
)

const (
	NotFoundPath = "can not find volume path"
)

func GetDevicePath(deviceId string) (string, error) {
	device := fmt.Sprintf("/dev/mapper/%s", deviceId)
	if !PathExists(device) {
		device = fmt.Sprintf("/dev/mapper/pv-%s", deviceId)
		if !PathExists(device) {
			err := fmt.Errorf("%s %s", NotFoundPath, deviceId)
			smslog.Infof(err.Error())
			return "", err
		}
	}
	return device, nil
}

func GetPBDName(volumeId string) (string, error) {
	if strings.HasPrefix(volumeId, "mapper_") {
		return volumeId, nil
	}
	device := fmt.Sprintf("/dev/mapper/%s", volumeId)
	if PathExists(device) {
		return fmt.Sprintf("mapper_%s", volumeId), nil
	}
	device = fmt.Sprintf("/dev/mapper/pv-%s", volumeId)
	if PathExists(device) {
		return fmt.Sprintf("mapper_pv-%s", volumeId), nil
	}

	err := fmt.Errorf("%s %s", NotFoundPath, volumeId)
	smslog.Infof(err.Error())
	return "", err
}

func PathExists(path string) bool {
	if _, err := os.Stat(path); err != nil {
		smslog.Debugf("PathExists %s err %s", path, err.Error())
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

func PathNotFoundError(err error) bool {
	if !strings.Contains(err.Error(), NotFoundPath) {
		return false
	}
	return true
}
