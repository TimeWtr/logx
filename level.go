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

import "fmt"

type LoggerLevel uint8

const (
	// DebugLevel 用于开发环境调试的日志级别，生产环境中需要切换其他的级别
	DebugLevel LoggerLevel = iota + 1
	// InfoLevel 默认的日志级别
	InfoLevel
	// WarnLevel 出现了危险的情况需要打印日志，存在危险，但不影响系统的正常运行
	WarnLevel
	// ErrorLevel 比WarnLevel更严重，业务出现了明显的错误，系统仍可正常运行
	ErrorLevel
	// PanicLevel 比ErrorLevel严重，出现的错误影响到了系统的正常运行，记录日志后Panic
	PanicLevel
	// FatalLevel 记录日志后，直接调用os.Exit(1)
	FatalLevel

	_minLevel = DebugLevel
	_maxLevel = FatalLevel
)

// String 用于校验并返回日志级别的小写格式的字符串内容
func (l LoggerLevel) String() string {
	switch l {
	case DebugLevel:
		return "debug"
	case InfoLevel:
		return "info"
	case WarnLevel:
		return "warn"
	case ErrorLevel:
		return "error"
	case PanicLevel:
		return "panic"
	case FatalLevel:
		return "fatal"
	default:
		return fmt.Sprintf("unknown level(%d)", l)
	}
}

// UpperString 用于校验并返回日志级别大写格式的字符串内容
func (l LoggerLevel) UpperString() string {
	switch l {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	case PanicLevel:
		return "PANIC"
	case FatalLevel:
		return "FATAL"
	default:
		return fmt.Sprintf("unknown level(%d)", l)
	}
}

// valid 校验是否是合法的日志级别
func (l LoggerLevel) valid() bool {
	return l > _maxLevel || l < _minLevel
}

// prohibit 校验日志级别，如果当前的日志级别比允许的级别高就返回为false，
// 允许打印日志，返回返回为true，禁止打印日志
func (l LoggerLevel) prohibit(level LoggerLevel) bool {
	return l > level
}

type LevelChecker interface {
	// 是否允许打印对应级别的日志
	check(LoggerLevel) bool
}
