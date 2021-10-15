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


package pv

import "time"

type PhysicalVolume struct {
	Id              int       `xorm:"not null pk autoincr INT"`
	VolumeId        string    `xorm:"not null default '“”' unique VARCHAR(45)"`
	VolumeName      string    `xorm:"VARCHAR(45)"`
	PvType          string    `xorm:"INT"`
	NodeId          string    `xorm:"VARCHAR(45)"`
	NodeIp          string    `xorm:"VARCHAR(45)"`
	ClusterId       int       `xorm:"not null default 0 INT"`
	SectorNum       int64     `xorm:"BIGINT"`
	SectorSize      int       `xorm:"INT"`
	Size            int64     `xorm:"BIGINT"`
	FsSize          int64     `xorm:"BIGINT"`
	FsType          string    `xorm:"VARCHAR(45)"`
	UsedSize        int64     `xorm:"BIGINT"`
	Status          string    `xorm:"VARCHAR(255)"`
	PrSupportStatus string    `xorm:"LONGTEXT"`
	PathNum         int       `xorm:"INT"`
	Paths           string    `xorm:"LONGTEXT"`
	Vendor          string    `xorm:"VARCHAR(45)"`
	Product         string    `xorm:"VARCHAR(45)"`
	Desc            string    `xorm:"VARCHAR(45)"`
	Extend          string    `xorm:"MEDIUMTEXT"`
	UsedByType      int       `xorm:"INT"`
	UsedByName      string    `xorm:"VARCHAR(45)"`
	SerialNumber    string    `xorm:"VARCHAR(45)"`
	Updated         time.Time `xorm:"DATETIME updated"`
	Created         time.Time `xorm:"DATETIME created"`
	DeletedAt       time.Time `xorm:"DATETIME deleted"`
}
