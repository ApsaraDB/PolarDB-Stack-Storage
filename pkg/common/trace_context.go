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

type TraceContext map[string]string

func (c TraceContext) String() string {
	bytes, err := StructToBytes(c)
	if err != nil {
		return ""
	}
	return string(bytes)
}

func NewTraceContext(value map[string]string) TraceContext {
	return value
}

func ParseForTraceContext(byteStr string) TraceContext {
	var ctx = TraceContext{}
	_ = BytesToStruct([]byte(byteStr), &ctx)
	return ctx
}
