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
	"encoding/json"
	"io"
	"io/ioutil"
	smslog "polardb-sms/pkg/log"
)

func StructToBytes(src interface{}) ([]byte, error) {
	ctxBytes, err := json.Marshal(src)
	if err != nil {
		smslog.Infof("Error to Marshal struct: %v, err: %v", src, err)
		return nil, err
	}
	return ctxBytes, nil
}

func BytesToStruct(contents []byte, tgt interface{}) error {
	err := json.Unmarshal(contents, tgt)
	if err != nil {
		return err
	}
	return nil
}

func IoStreamToStruct(r io.Reader, obj interface{}) error {
	x, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	return BytesToStruct(x, obj)
}

func MapToStruct(m map[string]interface{}, tgt interface{}) error {
	ctxBytes, err := StructToBytes(m)
	if err != nil {
		return err
	}
	if err := BytesToStruct(ctxBytes, tgt); err != nil {
		return err
	}
	return nil
}
