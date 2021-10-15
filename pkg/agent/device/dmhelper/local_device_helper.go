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

//func QueryLocalDevices(prefix string) (map[string]*device.DmDevice, error) {
//	var deviceMap = make(map[string]*device.DmDevice)
//	ret, err := path.ReadDirNoStat(prefix)
//	if err != nil {
//		return nil, err
//	}
//	for _, subPath := range ret {
//		smslog.Debugf("Start construct device %s", subPath)
//		d, err := constructLocalDevice(subPath)
//		smslog.Debugf("Finish construct device %s", param.Name)
//		if err != nil {
//			smslog.Debugf("construct err %s for param %v", err.Error(), param)
//			continue
//		}
//		deviceMap[d.Id()] = d
//	}
//	return deviceMap, nil
//}
//
//func constructLocalDevice(diskPath string) (*device.DmDevice, error) {
//	ret := &device.DmDevice{
//		Name:            "",
//		VolumeId:        "",
//		DeviceType:      "",
//		SectorNum:       0,
//		SectorSize:      0,
//		FsType:          "",
//		FsSize:          0,
//		PrSupportStatus: nil,
//		UsedSize:        0,
//		DmTarget:        nil,
//	}
//}
