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


package protocol

import (
	"k8s.io/apimachinery/pkg/util/uuid"
	"polardb-sms/pkg/common"
	"polardb-sms/pkg/device"
)

type EventType int

const (
	LunAdd EventType = iota
	LunUpdate
	LunRemove
	LunSubPathAdd
	LunSubPathRemove
	LvAdd
	LvUpdate
	LvRemove
)

type BatchEvent struct {
	Events []*Event `json:"events"`
	NodeId string   `json:"node_id"`
}

type Event struct {
	Id        string    `json:"id"`
	EventType EventType `json:"eType"`
	Body      string    `json:"body"`
}

func NewEvent(body string, eventType EventType) *Event {
	return &Event{
		Id:        string(uuid.NewUUID()),
		EventType: eventType,
		Body:      body,
	}
}

type Lun struct {
	Name         string                  `json:"name"`
	VolumeId     string                  `json:"volume_id"`
	LunType      string                  `json:"lun_type"`
	Paths        []string                `json:"paths"`
	PathNum      int                     `json:"path_num"`
	Vendor       string                  `json:"vendor"`
	Size         int64                   `json:"size"`
	Sectors      int64                   `json:"sectors"`
	SectorSize   int                     `json:"sectorSize"`
	FsType       common.FsType           `json:"fsType"`
	FsSize       int64                   `json:"fs_size"`
	NodeId       string                  `json:"nodeId"`
	NodeIp       string                  `json:"nodeIp"`
	PrSupport    *device.PrSupportReport `json:"prSupport"`
	UsedSize     int64                   `json:"used_size"`
	Product      string                  `json:"product"`
	SerialNumber string                  `json:"serial_number"`
}

type LunAddEvent struct {
	Lun
}

type LunSubPathAddEvent struct {
	Lun
	SubPath string `json:"subPath"`
}

type LunRemoveEvent struct {
	Lun
}

type LunSubPathRemoveEvent struct {
	Lun
	SubPath string `json:"subPath"`
}

type LunUpdateEvent struct {
	Lun
}

type Lv struct {
	VolumeId   string                  `json:"volume_id"`
	VolumeType string                  `json:"volume_type"`
	Sectors    int64                   `json:"sectors"`
	SectorSize int                     `json:"sectorSize"`
	Size       int64                   `json:"size"`
	FsType     common.FsType           `json:"fsType"`
	FsSize     int64                   `json:"fs_size"`
	NodeId     string                  `json:"nodeId"`
	NodeIp     string                  `json:"nodeIp"`
	Items      []*device.DmTableItem   `json:"items"`
	PrSupport  *device.PrSupportReport `json:"prSupport"`
	UsedSize   int64                   `json:"used_size"`
	Children   []string                `json:"children"`
}

type LvAddEvent struct {
	Lv
}

type LvUpdateEvent struct {
	Lv
}

type LvRemoveEvent struct {
	Lv
}
