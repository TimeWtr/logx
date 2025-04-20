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

type Options func(*Config)

// WithColor 是否开启日志输出颜色
func WithColor() Options {
	return func(l *Config) {
		l.enableColor = true
	}
}

// WithLevel 设置日志级别，如果不设置，默认级别是InfoLevel
func WithLevel(level LoggerLevel) Options {
	return func(l *Config) {
		l.level = level
	}
}

// WithFileName 配置日志文件名称
func WithFileName(fileName string) Options {
	return func(l *Config) {
		l.filename = fileName
	}
}

// WithLine 开启日志打印行号
func WithLine(line int) Options {
	return func(l *Config) {
		l.enableLine = true
	}
}

// WithCallSkip 设置ErrorLevel、PanicLevel和FatalLevel日志级别时打印的堆栈信息层级
func WithCallSkip(skip int) Options {
	return func(l *Config) {
		l.callSkip = skip
	}
}

// WithAsync 开启异步写入日志文件
func WithAsync() Options {
	return func(l *Config) {
		l.enableAsync = true
	}
}

func WithLocation(location string) Options {
	return func(l *Config) {
		l.location = location
	}
}
