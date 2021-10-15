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
	"github.com/stretchr/testify/require"
	"testing"
)

func TestParseDevices(t *testing.T) {
	info := `
fedora-root: 0 35643392 linear 8:2 4196352
fedora-root: 35643392 20963328 linear 8:3 2048
`
	d, err := ParseDMDevice("fedora-root", info)
	require.NoError(t, err)
	require.Equal(t, "fedora-root", d.Id())
}
