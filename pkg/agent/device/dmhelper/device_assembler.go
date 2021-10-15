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
	"polardb-sms/pkg/common"
	"polardb-sms/pkg/device"
	smslog "polardb-sms/pkg/log"
)

func assembleMultipathDevice(d *device.DmDevice) error {
	multipathParam, err := getMultipathParam(d.Name)
	if err == nil {
		d.DmTarget.(*device.MultipathDeviceTarget).Wwid = multipathParam.Wwid
		d.DmTarget.(*device.MultipathDeviceTarget).Vendor = multipathParam.Vendor
		d.DmTarget.(*device.MultipathDeviceTarget).Paths = multipathParam.Paths
		d.DmTarget.(*device.MultipathDeviceTarget).PathNum = multipathParam.PathNum
		d.DmTarget.(*device.MultipathDeviceTarget).Product = multipathParam.Product
		d.DmTarget.(*device.MultipathDeviceTarget).Name = multipathParam.Name
	} else {
		return err
	}

	d.SectorNum = d.DmTarget.(*device.MultipathDeviceTarget).DmTableItem.NumSectors
	fsParam, err := getFileSystemParam(d.Name)
	if err == nil {
		d.FsType = common.ParseFsType(fsParam.Type)
		d.FsSize = fsParam.BlockSize
		d.UsedSize = fsParam.Used
	} else {
		smslog.Debugf("getFileSystemParam err %s", err.Error())
	}
	blockParam, err := getBlockDevParam(d.Name)
	if err == nil {
		d.SectorSize = blockParam.SectorSize
	}
	d.PrSupportStatus = getPrSupportParam(d.Name)
	serialNumber, err := getSqInqParams(d.Name)
	if err == nil {
		d.SerialNumber = serialNumber
	}
	return nil
}

func assembleLinearDevice(d *device.DmDevice) error {
	for _, item := range d.DmTarget.(*device.LinearDeviceTarget).DmTableItems {
		d.SectorNum += item.NumSectors
	}
	fsParam, err := getFileSystemParam(d.Name)
	if err == nil {
		d.FsType = common.ParseFsType(fsParam.Type)
		d.FsSize = fsParam.BlockSize
		d.UsedSize = fsParam.Used
	} else {
		smslog.Debugf("getFileSystemParam err %s", err.Error())
	}
	blockParam, err := getBlockDevParam(d.Name)
	if err == nil {
		d.SectorSize = blockParam.SectorSize
	}
	d.PrSupportStatus = getPrSupportParam(d.Name)
	return nil
}

func assembleStrippedDevice(d *device.DmDevice) error {
	d.SectorNum = d.DmTarget.(*device.StripedDeviceTarget).DmTableItems[0].NumSectors
	fsParam, err := getFileSystemParam(d.Name)
	if err == nil {
		d.FsType = common.ParseFsType(fsParam.Type)
		d.FsSize = fsParam.BlockSize
		d.UsedSize = fsParam.Used
	} else {
		smslog.Debugf("getFileSystemParam err %s", err.Error())
	}
	blockParam, err := getBlockDevParam(d.Name)
	if err == nil {
		d.SectorSize = blockParam.SectorSize
	}
	d.PrSupportStatus = getPrSupportParam(d.Name)
	return nil
}
