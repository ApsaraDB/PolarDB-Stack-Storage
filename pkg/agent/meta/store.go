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

package meta

import (
	"errors"
)

var ErrNotImplement = errors.New("not implement")

type ListOption int

const (
	ListWithOutData ListOption = 0
)

type DMTableRecord struct {
	Name    string
	Data    string
	Version string
}

/**
* 基于文件系统实现, FileStore

* 基于DB实现，考虑sqlite等嵌入数据库
 */
type DMTableStore interface {
	// 更新dm设备的table定义
	Put(record *DMTableRecord) error

	// 查询dm设备的table定义
	Get(name string) (*DMTableRecord, error)

	// 删除dm设备的table定义
	Delete(name string) error

	// 查询所有dm设备的table定义
	List() (map[string]*DMTableRecord, error)

	// 查询dm设备历史版本table定义
	Versions(name string) ([]*DMTableRecord, error)
}

var _dmStore DMTableStore

func CreateDmStore(dir string) error {
	var err error
	_dmStore, err = NewDBStore(dir)
	return err
}

func GetDmStore() DMTableStore {
	return _dmStore
}
