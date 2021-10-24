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

/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package version

import (
	smslog "polardb-sms/pkg/log"
)

// These are set during build time via -ldflags
var (
	GitBranch string
	GitCommit string
	BuildDate string
	Module    string
)

func LogVersion() {
	smslog.Infof("--------------------------------------------------------------------------------------------")
	smslog.Infof("|                                                                                           |")
	smslog.Infof("|       polarbox branch: %v commit: %v   |", GitBranch, GitCommit)
	smslog.Infof("|       polarbox repo: %v                   |", Module)
	smslog.Infof("|       polarbox date: %v                                                 |", BuildDate)
	smslog.Infof("|                                                                                           |")
	smslog.Infof("---------------------------------------------------------------------------------------------")
}
