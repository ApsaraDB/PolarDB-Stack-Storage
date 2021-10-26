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

package config

import (
	"github.com/Unknwon/goconfig"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParse(t *testing.T) {
	testCase := assert.New(t)
	config, err := goconfig.LoadConfigFile("./manager.conf")
	ServerConf = ServerConfig{Port: DefaultPort}
	if err != nil {
		return
	}
	parse(config)
	testCase.Equal(1, len(GetAvailableNodes()))
}

func TestParseDbConfig(t *testing.T) {
	dbstr := "metabase:\n  host: 198.19.66.88\n  port: 3306\n  user: polar\n  password: polarstack2020\n  type: mysql\n  version: 5.7.31"
	parseDbConfig(dbstr)
}
