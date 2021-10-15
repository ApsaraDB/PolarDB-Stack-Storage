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
	"polardb-sms/pkg/agent/device/devicemapper"
	"polardb-sms/pkg/common"
	"polardb-sms/pkg/device"
	smslog "polardb-sms/pkg/log"
	"strings"
)

const (
	Byte            = 1
	KiB             = 1024 * Byte
	MiB             = 1024 * KiB
	GiB             = 1024 * MiB
	minimalDiskSize = 210 * GiB
)

func convertDmParamToDevice(param *DmParam) (*device.DmDevice, error) {
	var d = &device.DmDevice{
		Name:       param.Name,
		DeviceType: *param.DmType,
		SectorSize: 0,
		SectorNum:  0,
		FsType:     "",
		FsSize:     0,
	}

	switch *param.DmType {
	case device.Multipath:
		d.DmTarget = &device.MultipathDeviceTarget{DmTableItem: param.Items[0]}
		err := assembleMultipathDevice(d)
		if err != nil {
			return nil, err
		}
		return d, nil
	case device.Linear:
		if !strings.HasPrefix(param.Name, common.DmNamePrefix) {
			return nil, fmt.Errorf("invalid name %s, should start %s", param.Name, common.DmNamePrefix)
		}
		d.DmTarget = &device.LinearDeviceTarget{DmTableItems: param.Items}
		err := assembleLinearDevice(d)
		if err != nil {
			return nil, err
		}
		return d, nil
	case device.Striped:
		if !strings.HasPrefix(param.Name, common.DmNamePrefix) {
			return nil, fmt.Errorf("invalid name %s, should start %s", param.Name, common.DmNamePrefix)
		}
		d.DmTarget = &device.StripedDeviceTarget{DmTableItems: param.Items}
		err := assembleStrippedDevice(d)
		if err != nil {
			return nil, err
		}
		return d, nil
	default:
		return nil, fmt.Errorf("not support device type %s", *param.DmType)
	}
}

func validDevice(d *device.DmDevice) error {
	if d != nil {
		diskSize := int64(d.SectorSize) * d.SectorNum
		if diskSize < minimalDiskSize {
			return fmt.Errorf("diskSize %d is less than minimal size %d", diskSize, minimalDiskSize)
		}
		return nil
	}
	return fmt.Errorf("device is empty")
}

func constructDevice(param *DmParam) (*device.DmDevice, error) {
	d, err := convertDmParamToDevice(param)
	if err != nil {
		return nil, err
	}
	err = validDevice(d)
	if err != nil {
		return nil, err
	}
	return d, nil
}

func QueryDMDevice(name string) (*device.DmDevice, error) {
	dmParam, err := getDmParam(name)
	if err != nil {
		return nil, err
	}
	return constructDevice(dmParam)
}

func QueryDMDevices() (map[string]*device.DmDevice, error) {
	TryRepair()
	dmParams, err := getDmParams()
	smslog.Debugf("Finished get device mapper params %v", dmParams)
	if err != nil {
		return nil, err
	}
	var deviceMap = make(map[string]*device.DmDevice)

	for _, param := range dmParams {
		if BlackListFiltered(param.Name) {
			smslog.Debugf("%s disk is filtered", param.Name)
			continue
		}
		smslog.Debugf("Start construct device %s", param.Name)
		d, err := constructDevice(param)
		smslog.Debugf("Finish construct device %s", param.Name)
		if err != nil {
			smslog.Debugf("construct err %s for param %v", err.Error(), param)
			continue
		}
		deviceMap[d.Id()] = d
	}
	return deviceMap, nil
}

func ParseDMDevice(name, table string) (*device.DmDevice, error) {
	dmParam, err := parseDmParamByLines(name, strings.Split(strings.TrimSpace(table), device.NewLineSign))
	if err != nil {
		return nil, fmt.Errorf("faild to parse table for %s, table: %s, err: %s", name, table, err)
	}
	d, err := constructDevice(dmParam)

	//if !d.Validate() {
	//	return nil, fmt.Errorf("validate table for %s failed, table: %s, err: %s", name, table, err)
	//}

	return d, nil
}

func BatchQueryDevicesByName(names []string) ([]*device.DmDevice, error) {
	return nil, fmt.Errorf("unimplemented")
}

func FindDeviceByNameAndType(name string, deviceType device.DmDeviceType) (*device.DmDevice, error) {
	return nil, fmt.Errorf("unimplement FindDeviceByNameAndType")
}

func TryRepair() {
	_ = devicemapper.GetDeviceMapper().ScanDeviceFcHost()
	_ = devicemapper.GetDeviceMapper().ScanDeviceIscsiHost()
	_ = devicemapper.GetDeviceMapper().ScsiDeviceRescan()
	_ = devicemapper.GetDeviceMapper().MultiPathReload()
}
