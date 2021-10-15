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


package lv

import "time"

type LogicalVolume struct {
	Id           int       `xorm:"not null pk autoincr INT"`
	VolumeId     string    `xorm:"VARCHAR(45)"`
	VolumeName   string    `xorm:"not null unique VARCHAR(45)"`
	Children     string    `xorm:"TEXT"`
	LvType       string    `xorm:"VARCHAR(45)"`
	NodeIds      string    `xorm:"TEXT"`
	RelatedPvc   string    `xorm:"VARCHAR(45)"`
	FsSize       int64     `xorm:"BIGINT"`
	FsType       string    `xorm:"INT"`
	SectorNum    int64     `xorm:"BIGINT"`
	SectorSize   int       `xorm:"INT"`
	Size         int64     `xorm:"BIGINT"`
	UsedSize     int64     `xorm:"BIGINT"`
	Status       string    `xorm:"VARCHAR(45)"`
	PrStatus     string    `xorm:"MEDIUMTEXT"`
	PrNodeId     string    `xorm:"VARCHAR(45)"`
	Vendor       string    `xorm:"VARCHAR(45)"`
	Product      string    `xorm:"VARCHAR(45)"`
	Extend       string    `xorm:"MEDIUMTEXT"`
	ClusterId    int       `xorm:"INT"`
	UsedByType   int       `xorm:"INT"`
	UsedByName   string    `xorm:"VARCHAR(45)"`
	SerialNumber string    `xorm:"VARCHAR(45)"`
	Updated      time.Time `xorm:"DATETIME updated"`
	Created      time.Time `xorm:"DATETIME created"`
	DeletedAt    time.Time `xorm:"DATETIME deleted"`
}
