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
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestConcurrentSyncWrites(t *testing.T) {
	bw, _ := NewBufferWriter("./logs", time.Second)
	defer bw.Close()

	const layout = "2006-01-02 15:04:05"

	const template = "[INFO] logs/test.go line:23  this is a test log, Log entry: "
	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			var builder strings.Builder
			builder.WriteString(time.Now().Format(layout))
			builder.WriteString(template)
			builder.WriteString(strconv.Itoa(i))
			builder.WriteString("\n")
			msg := []byte(builder.String())
			if err := bw.SyncWrite(msg); err != nil {
				t.Errorf("Write failed: %v", err)
			}
		}(i)
	}
	wg.Wait()
}

func TestConcurrent_Async_Writes(t *testing.T) {
	bw, _ := NewBufferWriter("./logs", time.Second)
	defer bw.Close()

	const layout = "2006-01-02 15:04:05"

	var wg sync.WaitGroup
	for i := 0; i < 100000; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			msg := []byte(fmt.Sprintf("%s [INFO] logs/test.go line:23  this is a test log, Log entry %d\n",
				time.Now().Format(layout), i))
			if err := bw.AsyncWrite(msg); err != nil {
				t.Errorf("Write failed: %v", err)
			}
		}(i)
	}
	wg.Wait()
}

func TestNotConcurrentWrites(t *testing.T) {
	bw, _ := NewBufferWriter("./logs", time.Second)
	defer bw.Close()

	const layout = "2006-01-02 15:04:05"
	msg := []byte(fmt.Sprintf("%s [INFO] logs/test.go line:23  this is a test log, Log entry\n",
		time.Now().Format(layout)))
	for i := 0; i < 100000; i++ {
		if err := bw.AsyncWrite(msg); err != nil {
			t.Errorf("Write failed: %v", err)
		}
	}
}

func BenchmarkConcurrentWrites(b *testing.B) {
	bw, _ := NewBufferWriter("./logs", time.Second)
	defer bw.Close()

	const layout = "2006-01-02 15:04:05"

	msg := []byte(fmt.Sprintf("%s [INFO] logs/test.go line:23  this is a test log, Log entry\n",
		time.Now().Format(layout)))
	var wg sync.WaitGroup
	for i := 0; i < b.N; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_ = bw.AsyncWrite(msg)
		}(i)
	}
	wg.Wait()
}

func BenchmarkNotConcurrentWrites(b *testing.B) {
	bw, _ := NewBufferWriter("./logs", time.Second)
	defer bw.Close()

	const layout = "2006-01-02 15:04:05"

	msg := []byte(fmt.Sprintf("%s [INFO] logs/test.go line:23  this is a test log, Log entry\n",
		time.Now().Format(layout)))
	for i := 0; i < b.N; i++ {
		_ = bw.AsyncWrite(msg)
	}
}
