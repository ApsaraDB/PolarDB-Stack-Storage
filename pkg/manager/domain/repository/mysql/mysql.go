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


package mysql

import (
	_ "github.com/go-sql-driver/mysql"
	smslog "polardb-sms/pkg/log"
	dbConf "polardb-sms/pkg/manager/config"
	"sync"
	"time"
	"xorm.io/xorm"
)

var _engine *xorm.Engine
var _dbEngineOnce sync.Once

func GetDBEngine() *xorm.Engine {
	_dbEngineOnce.Do(func() {
		var err error
		//localConn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local",
		//	"root", "passw0rd", "127.0.0.1", "3306", "polardb_sms")
		for {
			_engine, err = xorm.NewEngine("mysql", dbConf.DBConnStr())
			if err != nil {
				smslog.Fatalf("mysql connect error %v, retry later", err)
				time.Sleep(2 * time.Second)
				continue
			}
			break
		}
		//_engine.ShowSQL(true)
		smslog.Infof("connect to DB %s successful", dbConf.DBConnStr())
	})
	return _engine
}
