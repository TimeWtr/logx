// Copyright 2025 TimeWtr
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package core

import (
	"fmt"
	"sync"
	"testing"
)

func TestNewBuffer(t *testing.T) {
	bf := NewBuffer(1000)

	ch := bf.Register()
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()

		for {
			select {
			case data, ok := <-ch:
				if !ok {
					return
				}
				t.Logf("【读取日志】日志内容为：%s", data)
			}
		}
	}()

	go func() {
		defer wg.Done()
		defer bf.Close()

		template := "2025-05-12 12:12:00 [Info] 日志写入测试，当前的序号为: %d\n"
		for i := 0; i < 2000; i++ {
			if i == 1000 {
				t.Log("counter: 1000")
			}
			err := bf.Write(fmt.Sprintf(template, i))
			if err != nil {
				t.Logf("写入日志失败，错误：%s\n", err.Error())
				continue
			}
			t.Logf("日志%d写入成功\n", i)
		}
	}()

	wg.Wait()
	t.Log("写入成功")
}
