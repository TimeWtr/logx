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
	"strings"
	"sync"
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

const (
	DefaultErrCoreSkip = 3
	DeaultLogSize      = 100 * 1024 * 1024
	DefaultPeriod      = 30
	DefaultLocation    = "Asia/Shanghai"
	DefaultFilename    = "server.log"
)

type Log struct {
	// 配置信息
	cfg *Config
	// 并发保护
	mu *sync.Mutex
	// 轮转策略
	rs *RotateStrategy
	// 日志加颜色输出
	cp ColorPlugin
}

func NewLog(filePath string, opts ...Options) (Logger, error) {
	cfg := &Config{
		filePath:   filePath,
		filename:   DefaultFilename,
		level:      InfoLevel,
		location:   DefaultLocation,
		enableLine: true,
		callSkip:   DefaultErrCoreSkip,
		threshold:  DeaultLogSize,
		period:     DefaultPeriod,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	rs, err := NewRotateStrategy(cfg.filename, cfg.threshold, cfg.enableCompress)
	if err != nil {
		return nil, err
	}

	l := &Log{
		cfg: cfg,
		mu:  new(sync.Mutex),
		rs:  rs,
		cp:  NewANSIColorPlugin(cfg.enableColor),
	}

	return l, nil
}

func (l *Log) prefix(level LoggerLevel, v ...any) string {
	var builder strings.Builder
	builder.WriteString(l.cp.Format(level))
	builder.WriteString(Streamline())
	builder.WriteString(fmt.Sprint(v...))
	return builder.String()
}

func (l *Log) prefixf(level LoggerLevel, format string, v ...any) string {
	var builder strings.Builder
	builder.WriteString(l.cp.Format(level))
	if level.check(InfoLevel) {
		builder.WriteString(Streamline())
	}
	builder.WriteString(fmt.Sprintf(format, v...))
	return builder.String()
}

func (l *Log) Debug(v ...any) {
	if l.cfg.level.check(DebugLevel) {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	l.rs.lg.Println(l.prefix(DebugLevel, v...))
}

func (l *Log) Info(v ...any) {
	if l.cfg.level.check(InfoLevel) {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	l.rs.lg.Println(l.prefix(InfoLevel, v...))
}

func (l *Log) Warn(v ...any) {
	if l.cfg.level.check(WarnLevel) {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	l.rs.lg.Println(l.prefix(WarnLevel, v...))
}

func (l *Log) Error(v ...any) {
	if l.cfg.level.check(ErrorLevel) {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	l.rs.lg.Println(l.prefix(ErrorLevel, v...))
	l.abnormalStack()
}

func (l *Log) Panic(v ...any) {
	if l.cfg.level.check(PanicLevel) {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	l.rs.lg.Println(l.prefix(PanicLevel, v...))
	l.abnormalStack()
}

func (l *Log) Fatal(v ...any) {
	if l.cfg.level.check(FatalLevel) {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	l.rs.lg.Println(l.prefix(FatalLevel, v...))
	l.abnormalStack()
}

func (l *Log) Debugf(format string, v ...any) {
	if l.cfg.level.check(DebugLevel) {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	l.rs.lg.Println(l.prefixf(DebugLevel, format, v...))
}

func (l *Log) Infof(format string, v ...any) {
	if l.cfg.level.check(InfoLevel) {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	l.rs.lg.Println(l.prefixf(InfoLevel, format, v...))
}

func (l *Log) Warnf(format string, v ...any) {
	if l.cfg.level.check(WarnLevel) {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	l.rs.lg.Println(l.prefixf(WarnLevel, format, v...))
}

func (l *Log) Errorf(format string, v ...any) {
	if l.cfg.level.check(ErrorLevel) {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	l.rs.lg.Println(l.prefixf(ErrorLevel, format, v...))
	l.abnormalStack()
}

func (l *Log) Panicf(format string, v ...any) {
	if l.cfg.level.check(PanicLevel) {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	l.rs.lg.Println(l.prefixf(PanicLevel, format, v...))
	l.abnormalStack()
}

func (l *Log) Fatalf(format string, v ...any) {
	if l.cfg.level.check(FatalLevel) {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	msg := l.prefixf(FatalLevel, format, v...)
	l.rs.lg.Println(msg)
	size := l.abnormalStack() + len(msg)
	l.rs.SetCurrentSize(int64(size))
}

// abnormalStack 用于打印异常情况下的多行堆栈信息，特殊处理，Debug、Info级别不需要
// 返回写入的数据大小
func (l *Log) abnormalStack() int {
	var builder strings.Builder
	for _, s := range MultiLevel(l.cfg.callSkip) {
		str := "\t" + s + "\n"
		builder.WriteString(str)
	}

	res := builder.String()
	_, _ = l.rs.logout.WriteString(res)
	return len(res)
}
