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
	"os"
	"runtime"
	"strings"
)

const parts = 4

// captureStack 捕获日志记录时文件调用的基本信息：程序文件、所在行数
func captureStack(skip int) string {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		return ""
	}

	return fmt.Sprintf("%s line:%d", file, line)
}

// Streamline 对捕获到的文件路径进行精简，只取最后四个部分的路径进行拼接返回，不取完整路径。
func Streamline() string {
	const skips = 2
	msg := captureStack(skips)
	sli := strings.Split(msg, string(os.PathSeparator))
	if len(sli) < parts {
		return msg
	}

	return strings.Join(sli[len(sli)-parts:], string(os.PathSeparator))
}

// MultiLevel 当出现异常情况时，比如发现错误，捕获4级调用关系
func MultiLevel(skips int) []string {
	sli := make([]string, 0, skips)
	for skip := 3; skip < skips; skip++ {
		msg := captureStack(skip)
		partSli := strings.Split(msg, string(os.PathSeparator))
		if len(partSli) < parts {
			sli = append(sli, msg)
			continue
		}

		sli = append(sli, strings.Join(partSli[len(partSli)-parts:], string(os.PathSeparator)))
	}

	return sli
}
