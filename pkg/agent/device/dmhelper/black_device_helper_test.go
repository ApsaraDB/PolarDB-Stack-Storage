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

import (
	"fmt"
	"os"
	"testing"
)

func TestBlacklist(t *testing.T) {
	testLines := "blacklist {\n    attachlist {\n    }\n    locallist {\n        centos-home\n        centos-swap\n        centos-root\n    }\n}"
	AgentFile = "test_local_file"
	file, _ := os.Create(AgentFile)
	_, _ = file.WriteString(testLines)
	_ = file.Sync()
	ret, err := List()
	if err != nil {
		println(err.Error())
	}
	println(fmt.Sprintf("%v", ret))
	println(len(ret))
	println(BlackListFiltered("centos-home"))
	_ = os.Remove(AgentFile)
}
