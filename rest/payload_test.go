/*
Copyright 2014 Workiva, LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

 http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package rest

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Ensures that Get returns nil and an error if the key doesn't exist.
func TestGetBadKey(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{}

	actual, err := payload.Get("foo")

	assert.Nil(actual, "Incorrect return value")
	assert.Equal(fmt.Errorf("No value with key 'foo'"), err, "Incorrect error value")
}

// Ensures that Get returns the correct value.
func TestGet(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{"foo": 1}

	actual, err := payload.Get("foo")

	assert.Equal(1, actual, "Incorrect return value")
	assert.Nil(err, "Error should be nil")
}

// Ensures that GetInt returns zero value and an error if the value isn't an int.
func TestGetIntBadValue(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{"foo": "bar"}

	actual, err := payload.GetInt("foo")

	assert.Equal(0, actual, "Incorrect return value")
	assert.Equal(fmt.Errorf("Value with key 'foo' not an int"), err, "Incorrect error value")
}

// Ensures that GetInt returns the correct value.
func TestGetInt(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{"foo": int(1)}

	actual, err := payload.Get("foo")

	assert.Equal(int(1), actual, "Incorrect return value")
	assert.Nil(err, "Error should be nil")
}

// Ensures that GetInt8 returns zero value and an error if the value isn't an int8.
func TestGetInt8BadValue(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{"foo": 9000}

	actual, err := payload.GetInt8("foo")

	assert.Equal(0, actual, "Incorrect return value")
	assert.Equal(fmt.Errorf("Value with key 'foo' not an int8"), err, "Incorrect error value")
}

// Ensures that GetInt8 returns the correct value.
func TestGetInt8(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{"foo": int8(1)}

	actual, err := payload.GetInt8("foo")

	assert.Equal(int8(1), actual, "Incorrect return value")
	assert.Nil(err, "Error should be nil")
}

// Ensures that GetInt16 returns zero value and an error if the value isn't an int16.
func TestGetInt16BadValue(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{"foo": "bar"}

	actual, err := payload.GetInt16("foo")

	assert.Equal(0, actual, "Incorrect return value")
	assert.Equal(fmt.Errorf("Value with key 'foo' not an int16"), err, "Incorrect error value")
}

// Ensures that GetInt16 returns the correct value.
func TestGetInt16(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{"foo": int16(1)}

	actual, err := payload.GetInt16("foo")

	assert.Equal(int16(1), actual, "Incorrect return value")
	assert.Nil(err, "Error should be nil")
}

// Ensures that GetInt32 returns zero value and an error if the value isn't an int32.
func TestGetInt32BadValue(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{"foo": "bar"}

	actual, err := payload.GetInt32("foo")

	assert.Equal(0, actual, "Incorrect return value")
	assert.Equal(fmt.Errorf("Value with key 'foo' not an int32"), err, "Incorrect error value")
}

// Ensures that GetInt32 returns the correct value.
func TestGetInt32(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{"foo": int32(1)}

	actual, err := payload.GetInt32("foo")

	assert.Equal(int32(1), actual, "Incorrect return value")
	assert.Nil(err, "Error should be nil")
}

// Ensures that GetInt64 returns zero value and an error if the value isn't an int64.
func TestGetInt64BadValue(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{"foo": "bar"}

	actual, err := payload.GetInt64("foo")

	assert.Equal(0, actual, "Incorrect return value")
	assert.Equal(fmt.Errorf("Value with key 'foo' not an int64"), err, "Incorrect error value")
}

// Ensures that GetInt64 returns the correct value.
func TestGetInt64(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{"foo": int64(1)}

	actual, err := payload.GetInt64("foo")

	assert.Equal(int64(1), actual, "Incorrect return value")
	assert.Nil(err, "Error should be nil")
}

// Ensures that GetUint returns zero value and an error if the value isn't an int.
func TestGetUintBadValue(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{"foo": "bar"}

	actual, err := payload.GetUint("foo")

	assert.Equal(0, actual, "Incorrect return value")
	assert.Equal(fmt.Errorf("Value with key 'foo' not a uint"), err, "Incorrect error value")
}

// Ensures that GetUint returns the correct value.
func TestGetUint(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{"foo": uint(1)}

	actual, err := payload.Get("foo")

	assert.Equal(uint(1), actual, "Incorrect return value")
	assert.Nil(err, "Error should be nil")
}

// Ensures that GetUint8 returns zero value and an error if the value isn't an int8.
func TestGetUint8BadValue(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{"foo": 9000}

	actual, err := payload.GetUint8("foo")

	assert.Equal(0, actual, "Incorrect return value")
	assert.Equal(fmt.Errorf("Value with key 'foo' not a uint8"), err, "Incorrect error value")
}

// Ensures that GetUint8 returns the correct value.
func TestGetUint8(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{"foo": uint8(1)}

	actual, err := payload.GetUint8("foo")

	assert.Equal(uint8(1), actual, "Incorrect return value")
	assert.Nil(err, "Error should be nil")
}

// Ensures that GetUint16 returns zero value and an error if the value isn't an int16.
func TestGetUint16BadValue(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{"foo": "bar"}

	actual, err := payload.GetUint16("foo")

	assert.Equal(0, actual, "Incorrect return value")
	assert.Equal(fmt.Errorf("Value with key 'foo' not a uint16"), err, "Incorrect error value")
}

// Ensures that GetUint16 returns the correct value.
func TestGetUint16(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{"foo": uint16(1)}

	actual, err := payload.GetUint16("foo")

	assert.Equal(uint16(1), actual, "Incorrect return value")
	assert.Nil(err, "Error should be nil")
}

// Ensures that GetUint32 returns zero value and an error if the value isn't an int32.
func TestGetUint32BadValue(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{"foo": "bar"}

	actual, err := payload.GetUint32("foo")

	assert.Equal(0, actual, "Incorrect return value")
	assert.Equal(fmt.Errorf("Value with key 'foo' not a uint32"), err, "Incorrect error value")
}

// Ensures that GetUint32 returns the correct value.
func TestGetUint32(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{"foo": uint32(1)}

	actual, err := payload.GetUint32("foo")

	assert.Equal(uint32(1), actual, "Incorrect return value")
	assert.Nil(err, "Error should be nil")
}

// Ensures that GetUint64 returns zero value and an error if the value isn't an int64.
func TestGetUint64BadValue(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{"foo": "bar"}

	actual, err := payload.GetUint64("foo")

	assert.Equal(0, actual, "Incorrect return value")
	assert.Equal(fmt.Errorf("Value with key 'foo' not a uint64"), err, "Incorrect error value")
}

// Ensures that GetUint64 returns the correct value.
func TestGetUint64(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{"foo": uint64(1)}

	actual, err := payload.GetUint64("foo")

	assert.Equal(uint64(1), actual, "Incorrect return value")
	assert.Nil(err, "Error should be nil")
}

// Ensures that GetFloat32 returns zero value and an error if the value isn't a float32.
func TestGetFloat32BadValue(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{"foo": "bar"}

	actual, err := payload.GetFloat32("foo")

	assert.Equal(0, actual, "Incorrect return value")
	assert.Equal(fmt.Errorf("Value with key 'foo' not a float32"), err, "Incorrect error value")
}

// Ensures that GetFloat32 returns the correct value.
func TestGetFloat32(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{"foo": float32(1)}

	actual, err := payload.GetFloat32("foo")

	assert.Equal(float32(1), actual, "Incorrect return value")
	assert.Nil(err, "Error should be nil")
}

// Ensures that GetFloat64 returns zero value and an error if the value isn't a float64.
func TestGetFloat64BadValue(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{"foo": "bar"}

	actual, err := payload.GetFloat64("foo")

	assert.Equal(0, actual, "Incorrect return value")
	assert.Equal(fmt.Errorf("Value with key 'foo' not a float64"), err, "Incorrect error value")
}

// Ensures that GetFloat64 returns the correct value.
func TestGetFloat64(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{"foo": float64(1)}

	actual, err := payload.GetFloat64("foo")

	assert.Equal(float64(1), actual, "Incorrect return value")
	assert.Nil(err, "Error should be nil")
}

// Ensures that GetByte returns zero value and an error if the value isn't a byte.
func TestGetByteBadValue(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{"foo": 1.0}

	actual, err := payload.GetByte("foo")

	assert.Equal(0, actual, "Incorrect return value")
	assert.Equal(fmt.Errorf("Value with key 'foo' not a byte"), err, "Incorrect error value")
}

// Ensures that GetByte returns the correct value.
func TestGetByte(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{"foo": byte(1)}

	actual, err := payload.GetByte("foo")

	assert.Equal(byte(1), actual, "Incorrect return value")
	assert.Nil(err, "Error should be nil")
}

// Ensures that GetString returns zero value and an error if the value isn't a string.
func TestGetStringBadValue(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{"foo": 1.0}

	actual, err := payload.GetString("foo")

	assert.Equal("", actual, "Incorrect return value")
	assert.Equal(fmt.Errorf("Value with key 'foo' not a string"), err, "Incorrect error value")
}

// Ensures that GetString returns the correct value.
func TestGetString(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{"foo": "bar"}

	actual, err := payload.GetString("foo")

	assert.Equal("bar", actual, "Incorrect return value")
	assert.Nil(err, "Error should be nil")
}

// Ensures that GetBool returns false and an error if the value isn't a bool.
func TestGetBoolBadValue(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{"foo": 1.0}

	actual, err := payload.GetBool("foo")

	assert.Equal(false, actual, "Incorrect return value")
	assert.Equal(fmt.Errorf("Value with key 'foo' not a bool"), err, "Incorrect error value")
}

// Ensures that GetBool returns the correct value.
func TestGetBool(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{"foo": true}

	actual, err := payload.GetBool("foo")

	assert.Equal(true, actual, "Incorrect return value")
	assert.Nil(err, "Error should be nil")
}

// Ensures that GetSlice returns nil and an error if the value isn't a slice.
func TestGetSliceBadValue(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{"foo": 1.0}

	actual, err := payload.GetSlice("foo")

	assert.Equal([]interface{}(nil), actual, "Incorrect return value")
	assert.Equal(fmt.Errorf("Value with key 'foo' not a slice"), err, "Incorrect error value")
}

// Ensures that GetSlice returns the correct value.
func TestGetSlice(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{"foo": []interface{}{1, 2, 3}}

	actual, err := payload.GetSlice("foo")

	assert.Equal([]interface{}{1, 2, 3}, actual, "Incorrect return value")
	assert.Nil(err, "Error should be nil")
}

// Ensures that GetMap returns nil and an error if the value isn't a map.
func TestGetMapBadValue(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{"foo": 1.0}

	actual, err := payload.GetMap("foo")

	assert.Equal(map[string]interface{}(nil), actual, "Incorrect return value")
	assert.Equal(fmt.Errorf("Value with key 'foo' not a map"), err, "Incorrect error value")
}

// Ensures that GetMap returns the correct value.
func TestGetMap(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{"foo": map[string]interface{}{"a": 1}}

	actual, err := payload.GetMap("foo")

	assert.Equal(map[string]interface{}{"a": 1}, actual, "Incorrect return value")
	assert.Nil(err, "Error should be nil")
}

// Ensures that GetDuration returns zero value and an error if the value isn't a
// time.Duration.
func TestGetDurationBadValue(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{"foo": "bar"}

	actual, err := payload.GetDuration("foo")

	assert.Equal(0, actual, "Incorrect return value")
	assert.Equal(fmt.Errorf("Value with key 'foo' not a time.Duration"),
		err, "Incorrect error value")
}

// Ensures that GetDuration returns the correct value.
func TestGetDuration(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{"foo": time.Duration(100)}

	actual, err := payload.GetDuration("foo")

	assert.Equal(time.Duration(100), actual, "Incorrect return value")
	assert.Nil(err, "Error should be nil")
}

// Ensures that GetTime returns zero value and an error if the value isn't a time.Time.
func TestGetTimeBadValue(t *testing.T) {
	assert := assert.New(t)
	payload := Payload{"foo": "bar"}

	actual, err := payload.GetTime("foo")

	assert.Equal(time.Time{}, actual, "Incorrect return value")
	assert.Equal(fmt.Errorf("Value with key 'foo' not a time.Time"),
		err, "Incorrect error value")
}

// Ensures that GetTime returns the correct value.
func TestGetTime(t *testing.T) {
	assert := assert.New(t)
	now := time.Now()
	payload := Payload{"foo": now}

	actual, err := payload.GetTime("foo")

	assert.Equal(now, actual, "Incorrect return value")
	assert.Nil(err, "Error should be nil")
}
