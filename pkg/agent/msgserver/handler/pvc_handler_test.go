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


package handler

import (
	"github.com/stretchr/testify/assert"
	"os"
	smslog "polardb-sms/pkg/log"
	"testing"
)

func TestNewPvcCreateHandler(t *testing.T) {
	var source = "/dev/disk0"
	var target = "/Users/jimmy/Documents/testdisk"
	err := os.Symlink(source, target)
	smslog.Infof("xxx %v", err)

	ori, _ := os.Readlink(target)
	smslog.Infof("xxx:  %s", ori)
	testCase := assert.New(t)
	testCase.Equal(source, ori)
}
