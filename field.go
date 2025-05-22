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

// FType 字段类型
type FType uint8

const (
	// ArrTypeField 数组格式的字段类型
	ArrTypeField FType = iota + 1
	// ObjectTypeField 对象格式的字段类型
	ObjectTypeField
	// BinaryTypeField 二进制格式的字段类型
	BinaryTypeField
	// JSONTypeField Json格式的字段类型
	JSONTypeField
	// BoolTypeField 布尔格式的字段类型
	BoolTypeField
	// IntTypeField 数值格式的字段类型
	IntTypeField
	// StringTypeField 字符串格式的字段类型
	StringTypeField
	// FloatTypeField 浮点格式的字段类型
	FloatTypeField
	// DatetimeTypeField 时间格式的字段类型
	DatetimeTypeField
)

type Field struct {
	// 存储的字段名
	Key string
	// 字段类型
	Type FType
	// 存储的复杂对象
	Value any
}
