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


package service

import (
	"encoding/json"
	"fmt"
	"polardb-sms/pkg/common"
	"polardb-sms/pkg/device"
	smslog "polardb-sms/pkg/log"
	"polardb-sms/pkg/protocol"
	"sync"
)

type Transformer interface {
	Transform(*device.DmDevice, protocol.EventType) interface{}
}
type NodeInfo struct {
	nodeId string
	nodeIp string
}
type LunTransformer struct {
	NodeInfo
}

func (t *LunTransformer) Transform(d *device.DmDevice, eventType protocol.EventType) interface{} {
	mt := d.DmTarget.(*device.MultipathDeviceTarget)
	lun := &protocol.Lun{
		Name:         mt.Name,
		VolumeId:     mt.Wwid,
		Paths:        mt.Paths,
		PathNum:      mt.PathNum,
		Vendor:       mt.Vendor,
		Size:         d.SectorNum * int64(d.SectorSize),
		Sectors:      d.SectorNum,
		SectorSize:   d.SectorSize,
		FsType:       d.FsType,
		FsSize:       d.FsSize,
		NodeId:       t.nodeId,
		NodeIp:       t.nodeIp,
		PrSupport:    d.PrSupportStatus,
		UsedSize:     d.UsedSize,
		Product:      mt.Product,
		SerialNumber: d.SerialNumber,
	}
	body, err := json.Marshal(lun)
	if err != nil {
		smslog.Errorf("event lun - (%s) marshal body %v: %v", device.Multipath, lun, err)
		return nil
	}
	return protocol.NewEvent(string(body), eventType)
}

type StripLvTransformer struct {
	NodeInfo
}

func (t *StripLvTransformer) Transform(d *device.DmDevice, eventType protocol.EventType) interface{} {
	mt := d.DmTarget.(*device.StripedDeviceTarget)
	lv := &protocol.Lv{
		VolumeId:   d.Name,
		VolumeType: string(common.DmStripVolume),
		Sectors:    d.SectorNum,
		SectorSize: d.SectorSize,
		Size:       int64(d.SectorSize) * d.SectorNum,
		FsType:     d.FsType,
		FsSize:     d.FsSize,
		NodeId:     t.nodeId,
		NodeIp:     t.nodeIp,
		Items:      mt.DmTableItems,
		UsedSize:   d.UsedSize,
		PrSupport:  d.PrSupportStatus,
		Children:   mt.GetChildren(),
	}
	body, err := json.Marshal(lv)
	if err != nil {
		return fmt.Errorf("event lv - (%s) marshal body %v: %v", device.Linear, lv, err)
	}
	return protocol.NewEvent(string(body), eventType)
}

type LinearLvTransformer struct {
	NodeInfo
}

func (t *LinearLvTransformer) Transform(d *device.DmDevice, eventType protocol.EventType) interface{} {
	mt := d.DmTarget.(*device.LinearDeviceTarget)
	lv := &protocol.Lv{
		VolumeId:   d.Name,
		VolumeType: string(common.DmLinearVolume),
		Sectors:    d.SectorNum,
		SectorSize: d.SectorSize,
		Size:       int64(d.SectorSize) * d.SectorNum,
		FsType:     d.FsType,
		FsSize:     d.FsSize,
		NodeId:     t.nodeId,
		NodeIp:     t.nodeIp,
		Items:      mt.DmTableItems,
		UsedSize:   d.UsedSize,
		PrSupport:  d.PrSupportStatus,
		Children:   mt.GetChildren(),
	}

	body, err := json.Marshal(lv)
	if err != nil {
		return fmt.Errorf("event lv - (%s) marshal body %v: %v", device.Linear, lv, err)
	}
	return protocol.NewEvent(string(body), eventType)
}

var _lunTransformer, _stripTransformer, _linearTransformer Transformer
var _lunOnce, _stripOnce, _linearOnce sync.Once

func getTransformer(deviceType device.DmDeviceType, nodeId, nodeIp string) Transformer {
	switch deviceType {
	case device.Multipath:
		return getLunTransformer(nodeId, nodeIp)
	case device.Linear:
		return getLinearTransformer(nodeId, nodeIp)
	default:
		return getStripTransformer(nodeId, nodeIp)
	}
}

func getLunTransformer(nodeId, nodeIp string) Transformer {
	_lunOnce.Do(func() {
		if _lunTransformer == nil {
			_lunTransformer = &LunTransformer{NodeInfo{
				nodeId: nodeId,
				nodeIp: nodeIp,
			}}
		}
	})
	return _lunTransformer
}

func getLinearTransformer(nodeId, nodeIp string) Transformer {
	_linearOnce.Do(func() {
		if _linearTransformer == nil {
			_linearTransformer = &LinearLvTransformer{NodeInfo{
				nodeId: nodeId,
				nodeIp: nodeIp,
			}}
		}
	})
	return _linearTransformer
}

func getStripTransformer(nodeId, nodeIp string) Transformer {
	_stripOnce.Do(func() {
		if _stripTransformer == nil {
			_stripTransformer = &StripLvTransformer{NodeInfo{
				nodeId: nodeId,
				nodeIp: nodeIp,
			}}
		}
	})
	return _stripTransformer
}
