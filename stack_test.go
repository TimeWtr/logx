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
	"testing"
)

func m() string {
	return ""
}

func g() []string {
	return nil
}

func f1() []string {
	return g()
}

func f2() []string {
	return f1()
}

func f3() []string {
	return f2()
}

func f4() []string {
	return f3()
}

func TestStreamline(t *testing.T) {
	fmt.Println(m(), "test content")
}

func TestMultiLevel(t *testing.T) {
	fmt.Println(f4(), "test multi level content")
}

func TestMultiLevel_line(t *testing.T) {
	fmt.Println("test multi level content")
	levels := f4()
	for _, level := range levels {
		fmt.Println(level)
	}
}

func TestCallEntityWrap_Fullname(t *testing.T) {
	cew := newCallEntityWrap()
	for i := 0; i < 10000; i++ {
		t.Logf("fullename: %s", cew.Fullname())
	}
}

func TestCallEntityWrap_Fullnames(t *testing.T) {
	cew := newCallEntityWrap(WithPC(), WithSkip(3))
	for i := 0; i < 10000; i++ {
		t.Logf("fullename: %s", cew.Fullnames())
	}
}

func BenchmarkCallEntityWrap_Fullnames_NotPC(b *testing.B) {
	cew := newCallEntityWrap(WithSkip(5))
	for i := 0; i < 10000; i++ {
		b.Logf("fullename: %s", cew.Fullname())
	}
}

func BenchmarkCallEntityWrap_Fullnames_PC(b *testing.B) {
	cew := newCallEntityWrap(WithPC(), WithSkip(5))
	for i := 0; i < 10000; i++ {
		b.Logf("fullename: %s", cew.Fullnames())
	}
}
