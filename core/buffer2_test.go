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
	"strconv"
	"strings"
	"sync"
	"testing"
)

func TestNewBuffer(t *testing.T) {
	bf := NewBuffer(2000)

	ch := bf.Register()
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()

		counter := 0
		for {
			select {
			case data, ok := <-ch:
				if !ok {
					t.Log("通道关闭")
					t.Logf("接收到日志数据条数: %d", counter)
					return
				}
				t.Logf("【读取日志】日志内容为：%s", data)
				counter++
			default:

			}
		}
	}()

	go func() {
		defer wg.Done()
		defer bf.Close()

		template := "2025-05-12 12:12:00 [Info] 日志写入测试，当前的序号为: %d\n"
		for i := 0; i < 500000; i++ {
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

func BenchmarkNewBuffer(b *testing.B) {
	bf := NewBuffer(5000)

	ch := bf.Register()
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()

		counter := 0
		for {
			select {
			case data, ok := <-ch:
				if !ok {
					b.Log("通道关闭")
					b.Logf("接收到日志数据条数: %d", counter)
					return
				}
				b.Logf("【读取日志】日志内容为：%s", data)
				counter++
			default:

			}
		}
	}()

	go func() {
		defer wg.Done()
		defer bf.Close()

		template := "2025-05-12 12:12:00 [Info] 日志写入测试，当前的序号为: "
		for i := 0; i < b.N; i++ {
			var builder strings.Builder
			builder.WriteString(template)
			builder.WriteString(strconv.Itoa(i))
			builder.WriteString("\n")
			err := bf.Write(builder.String())
			if err != nil {
				b.Logf("写入日志失败，错误：%s\n", err.Error())
				continue
			}
			b.Logf("日志%d写入成功\n", i)
		}
	}()

	wg.Wait()
	b.Log("写入成功")
}

func BenchmarkNewBuffer_No_Log(b *testing.B) {
	bf := NewBuffer(5000)

	ch := bf.Register()
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()

		counter := 0
		for {
			select {
			case _, ok := <-ch:
				if !ok {
					b.Log("收到日志条数: ", counter)
					return
				}
				counter++
			default:

			}
		}
	}()

	go func() {
		defer wg.Done()
		defer bf.Close()

		template := "2025-05-12 12:12:00 [Info] 日志写入测试，当前的序号为: "
		for i := 0; i < b.N; i++ {
			var builder strings.Builder
			builder.WriteString(template)
			builder.WriteString(strconv.Itoa(i))
			builder.WriteString("\n")
			err := bf.Write(builder.String())
			if err != nil {
				b.Logf("写入日志失败，错误：%s\n", err.Error())
				continue
			}
		}
	}()

	wg.Wait()
	b.Log("写入成功")
}
