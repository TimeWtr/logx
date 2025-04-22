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
)

func f(lg Logger) {
	sf(lg)
}

func sf(lg Logger) {
	for i := 0; i < 10000; i++ {
		//lg.Info("hello world")
		//lg.Debugf("hello world")
		//lg.Warn("hello world")
		//lg.Error("test error, err is: ", errors.New("this is a test error"))
	}
}

func TestNewLog(t *testing.T) {
	lg, err := NewLog(
		"./logs",
		WithColor(),
		WithAsync(),
		WithThreshold(1024*100),
		WithCallSkip(3))
	assert.NoError(t, err)
	assert.NotNil(t, lg)
	f(lg)
}
