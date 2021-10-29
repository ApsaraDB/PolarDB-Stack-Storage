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

package repository

import (
	smslog "polardb-sms/pkg/log"
	"polardb-sms/pkg/manager/domain/repository/mysql"
	"sync"
	"xorm.io/xorm"
	"xorm.io/xorm/caches"
)

var _dbIns *BaseDB
var _dbInsOnce sync.Once

type BaseDB struct {
	Engine *xorm.Engine
	Cacher caches.Cacher
}

func GetBaseDB() *BaseDB {
	_dbInsOnce.Do(func() {
		_dbIns = &BaseDB{
			Engine: mysql.GetDBEngine(),
			Cacher: caches.NewLRUCacher(caches.NewMemoryStore(), 1024),
		}
	})
	return _dbIns
}

func (db *BaseDB) CacheTable(tableStruct interface{}) {
	err := db.Engine.MapCacher(tableStruct, db.Cacher)
	if err != nil {
		smslog.Errorf("cache table err %s", tableStruct)
	}
}
