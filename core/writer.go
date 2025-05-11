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
	"io"
)

//// Writer 写入器抽象接口
//type Writer interface {
//	// LevelChecker 校验是否允许打印日志
//	LevelChecker
//	// AsyncWriter 异步缓冲队列接口，用户设置缓冲区大小和刷新
//	AsyncWriter
//	// Write 执行写入操作的方法
//	Write([]byte) error
//	// Close 关闭方法，用于资源的释放
//	Close()
//}
//
//// AsyncWriter 异步缓冲队列接口
//type AsyncWriter interface {
//	// Flush 刷盘
//	Flush() error
//	// SetBufferSize 设置缓冲区大小
//	SetBufferSize(int)
//}

// Entity 结构化日志数据格式
type Entity struct {
	// 日志时间戳
	Timestamp int64
	// 日志级别
	Level LoggerLevel
	// 分布式追踪ID
	TraceID string
	// 服务名称
	Service string
	// 消息主体
	Message string
	// 结构化信息
	Fields map[string]any
	// 堆栈数据，可以是单条，也可以是多条，多条对应的是ErrorLevel、PanicLevel和FatalLevel级别
	CE []CallerEntity
}

// Writer 定义抽象的Writer接口，支持文件、网络、终端和消息队列(Kafka)的写入/输出
type Writer interface {
	io.Writer
	//Flush 刷新缓冲区(文件、网络)
	Flush() error
	// Close 释放资源
	Close() error
}

type FileWriter struct {
	w io.Writer
}

func NewFileWriter() Writer {
	return &FileWriter{}
}

func (f *FileWriter) Write(p []byte) (n int, err error) {
	//TODO implement me
	panic("implement me")
}

func (f *FileWriter) Flush() error {
	//TODO implement me
	panic("implement me")
}

func (f *FileWriter) Close() error {
	//TODO implement me
	panic("implement me")
}
