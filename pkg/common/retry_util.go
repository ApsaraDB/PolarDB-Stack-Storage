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

package common

import (
	smslog "polardb-sms/pkg/log"
	"time"
)

func RunWithRetry(max int, retryInterval time.Duration, f func(retryTimes int) error) error {
	var err error
	for i := 0; i < max; i++ {
		if err = f(i); err == nil {
			return nil
		}
		smslog.Error(err.Error())
		time.Sleep(retryInterval)
	}
	return err
}
