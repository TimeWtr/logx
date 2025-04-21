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

const (
	DebugColor Color = iota + 30
	InfoColor
	WarnColor
	ErrorColor
	PanicColor
	FatalColor
)

type Color uint8

func (c Color) String(s string) string {
	return fmt.Sprintf("\x1b[1;%dm[%s] \x1b[0m", uint8(c), s)
}

// ColorPlugin 日志颜色插件
type ColorPlugin interface {
	Format(enabled bool, level LoggerLevel) string
}

type ANSIColorPlugin struct{}

func NewANSIColorPlugin() ColorPlugin {
	return &ANSIColorPlugin{}
}

func (p *ANSIColorPlugin) Format(enabled bool, level LoggerLevel) string {
	if enabled {
		switch level {
		case DebugLevel:
			return DebugColor.String(level.UpperString())
		case InfoLevel:
			return InfoColor.String(level.UpperString())
		case WarnLevel:
			return WarnColor.String(level.UpperString())
		case ErrorLevel:
			return ErrorColor.String(level.UpperString())
		case PanicLevel:
			return PanicColor.String(level.UpperString())
		case FatalLevel:
			return FatalColor.String(level.UpperString())
		default:
		}
	}

	return fmt.Sprintf("[" + level.UpperString() + "] ")
}
