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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLevel(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name    string
		level   LoggerLevel
		wantRes bool
	}{
		{
			name:    "合法level",
			level:   DebugLevel,
			wantRes: true,
		},
		{
			name:    "不合法level_1",
			level:   100,
			wantRes: false,
		},
		{
			name:    "不合法level_2",
			level:   DebugLevel - 1,
			wantRes: false,
		},
	}

	for _, tcs := range testCases {
		tc := tcs
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			res := tc.level.valid()
			assert.Equal(t, tc.wantRes, res)
			t.Log(tc.level.String())
			t.Log(tc.level.UpperString())
		})
	}
}

func TestProhibit(t *testing.T) {
	t.Parallel()
	// 当前的日志级别
	level := InfoLevel
	testCases := []struct {
		name    string
		level   LoggerLevel
		input   LoggerLevel
		valid   bool
		wantRes bool
	}{
		{
			name:    "不允许输出_DebugLevel",
			level:   level,
			input:   DebugLevel,
			valid:   true,
			wantRes: false,
		},
		{
			name:    "允许输出_InfoLevel",
			level:   level,
			input:   InfoLevel,
			valid:   true,
			wantRes: true,
		},
		{
			name:    "允许输出_ErrorLevel",
			level:   level,
			input:   ErrorLevel,
			valid:   true,
			wantRes: true,
		},
		{
			name:    "允许输出_FatalLevel",
			level:   level,
			input:   FatalLevel,
			valid:   true,
			wantRes: true,
		},
	}

	for _, tcs := range testCases {
		tc := tcs
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			res := tc.level.valid()
			assert.Equal(t, tc.valid, res)
			allow := tc.level.Prohibit(tc.input)
			assert.Equal(t, tc.wantRes, allow)
			t.Log(tc.level.String())
			t.Log(tc.level.UpperString())
		})
	}
}
