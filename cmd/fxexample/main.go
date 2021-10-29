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

package main

import (
	"context"
	"fmt"
	"go.uber.org/fx"
)

func main() {
	type t3 struct {
		Name string
	}

	type t4 struct {
		Age int
	}

	var (
		v1 *t3
		v2 *t4
	)

	app := fx.New(
		fx.Provide(func() *t3 { return &t3{"hello everybody!!!"} }),
		fx.Provide(func() *t4 { return &t4{2019} }),

		fx.Populate(&v1),
		fx.Populate(&v2),
	)

	app.Start(context.Background())
	defer app.Stop(context.Background())

	fmt.Printf("the reulst is %v , %v\n", v1.Name, v2.Age)
}
