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
	"fmt"
	"log"
	"os"
	"strings"
)

type Logger interface {
	Debug(v ...any)
	Info(v ...any)
	Warn(v ...any)
	Error(v ...any)
	Panic(v ...any)
	Fatal(v ...any)
	Debugf(format string, v ...any)
	Infof(format string, v ...any)
	Warnf(format string, v ...any)
	Errorf(format string, v ...any)
	Panicf(format string, v ...any)
	Fatalf(format string, v ...any)
}

type Log struct {
	// 日志级别
	level LoggerLevel
	// 日志输出文件
	loggerFile string
	// 集成原生日志包
	lg *log.Logger
}

func NewLog(filePath string, level LoggerLevel) Logger {
	logout, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		_, _ = os.Stderr.WriteString(fmt.Sprintf("Failed to open log file: %s, err: %v\n", filePath, err))
	}

	return &Log{
		level:      level,
		loggerFile: filePath,
		lg:         log.New(logout, "", log.Ldate|log.Lmicroseconds),
	}
}

func (l *Log) prefix(level LoggerLevel, v ...any) string {
	var builder strings.Builder
	builder.WriteString("[")
	builder.WriteString(level.UpperString())
	builder.WriteString("] ")
	builder.WriteString(fmt.Sprint(v...))
	return builder.String()
}

func (l *Log) prefixf(level LoggerLevel, format string, v ...any) string {
	var builder strings.Builder
	builder.WriteString("[")
	builder.WriteString(level.UpperString())
	builder.WriteString("] ")
	builder.WriteString(fmt.Sprintf(format, v...))
	return builder.String()
}

func (l *Log) Debug(v ...any) {
	if l.level.check(DebugLevel) {
		l.lg.Println(l.prefix(DebugLevel, v...))
	}
}

func (l *Log) Info(v ...any) {
	if l.level.check(InfoLevel) {
		l.lg.Println(l.prefix(InfoLevel, v...))
	}
}

func (l *Log) Warn(v ...any) {
	if l.level.check(WarnLevel) {
		l.lg.Println(l.prefix(WarnLevel, v...))
	}
}

func (l *Log) Error(v ...any) {
	if l.level.check(ErrorLevel) {
		l.lg.Println(l.prefix(ErrorLevel, v...))
	}
}

func (l *Log) Panic(v ...any) {
	if l.level.check(PanicLevel) {
		l.lg.Println(l.prefix(PanicLevel, v...))
	}
}

func (l *Log) Fatal(v ...any) {
	if l.level.check(FatalLevel) {
		l.lg.Println(l.prefix(FatalLevel, v...))
	}
}

func (l *Log) Debugf(format string, v ...any) {
	if l.level.check(DebugLevel) {
		l.lg.Println(l.prefixf(DebugLevel, format, v...))
	}
}

func (l *Log) Infof(format string, v ...any) {
	if l.level.check(InfoLevel) {
		l.lg.Println(l.prefixf(InfoLevel, format, v...))
	}
}

func (l *Log) Warnf(format string, v ...any) {
	if l.level.check(WarnLevel) {
		l.lg.Println(l.prefixf(WarnLevel, format, v...))
	}
}

func (l *Log) Errorf(format string, v ...any) {
	if l.level.check(ErrorLevel) {
		l.lg.Println(l.prefixf(ErrorLevel, format, v...))
	}
}

func (l *Log) Panicf(format string, v ...any) {
	if l.level.check(PanicLevel) {
		l.lg.Println(l.prefixf(PanicLevel, format, v...))
	}
}

func (l *Log) Fatalf(format string, v ...any) {
	if l.level.check(FatalLevel) {
		l.lg.Println(l.prefixf(FatalLevel, format, v...))
	}
}
