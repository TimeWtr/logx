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

package logx

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNewRotateStrategy(t *testing.T) {
	cfg := &Config{
		filePath:         "./logs",
		filename:         "test.log",
		threshold:        200,
		enableCompress:   true,
		compressionLevel: DefaultCompression,
	}
	rs, err := NewRotateStrategy(cfg)
	assert.Nil(t, err)

	for i := 0; i < 100; i++ {
		err = rs.Rotate()
		assert.Nil(t, err)
		rs.SetCurrentSize(40)
	}
}

func TestNewRotateStrategy_Async_Work(t *testing.T) {
	cfg := &Config{
		filePath:         "./logs",
		filename:         "test.log",
		threshold:        200,
		enableCompress:   true,
		compressionLevel: DefaultCompression,
	}
	rs, err := NewRotateStrategy(cfg)
	assert.Nil(t, err)

	go rs.asyncWork()
	for i := 0; i < 100; i++ {
		err = rs.Rotate()
		assert.Nil(t, err)
		rs.SetCurrentSize(40)
	}

	time.Sleep(1 * time.Second)
	rs.Close()
}

//func TestNewRotateStrategy_Async_Clean_Work(t *testing.T) {
//	cfg := &Config{
//		filePath:         "./logs",
//		filename:         "test.log",
//		threshold:        200,
//		enableCompress:   true,
//		compressionLevel: DefaultCompression,
//		period:           3,
//	}
//	rs, err := NewRotateStrategy(cfg)
//	assert.Nil(t, err)
//	defer rs.Close()
//
//	go rs.asyncWork()
//	for i := 0; i < 100; i++ {
//		err = rs.Rotate()
//		assert.Nil(t, err)
//		rs.SetCurrentSize(40)
//	}
//
//	err = os.Rename(filepath.Join("logs", time.Now().Format(Layout)), "./logs/20250416")
//	assert.Nil(t, err)
//}

// ExampleNewRotateStrategy 日志轮转事例
// 1. 初始化日志轮转对象
// 2. 异步开启周期任务
// 3. 写入数据和设置当前写入数据的大小
// 4. 结束后调用关闭方法，停掉定时任务
func ExampleNewRotateStrategy() {
	cfg := &Config{
		filePath:         "./logs",
		filename:         "test.log",
		threshold:        200,
		enableCompress:   true,
		compressionLevel: DefaultCompression,
	}
	rs, err := NewRotateStrategy(cfg)
	if err != nil {
		return
	}

	go rs.asyncWork()
	defer rs.Close()

	for i := 0; i < 100; i++ {
		err = rs.Rotate()
		if err != nil {
			return
		}
		rs.SetCurrentSize(50)
	}

	time.Sleep(1 * time.Second)
}
