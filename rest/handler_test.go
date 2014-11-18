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
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Ensures that decodePayload returns an empty map for empty payloads.
func TestDecodePayloadEmpty(t *testing.T) {
	assert := assert.New(t)
	payload := bytes.NewBufferString("")

	decoded, err := decodePayload(payload.Bytes())

	assert.Equal(Payload{}, decoded)
	assert.Nil(err)
}

// Ensures that decodePayload returns a nil and an error for invalid JSON payloads.
func TestDecodePayloadBadJSON(t *testing.T) {
	assert := assert.New(t)
	body := `{"foo": "bar", "baz": 1`
	payload := bytes.NewBufferString(body)

	decoded, err := decodePayload(payload.Bytes())

	assert.Nil(decoded)
	assert.NotNil(err)
}

// Ensures that decodePayload returns a decoded map for JSON payloads.
func TestDecodePayloadHappyPath(t *testing.T) {
	assert := assert.New(t)
	body := `{"foo": "bar", "baz": 1}`
	payload := bytes.NewBufferString(body)

	decoded, err := decodePayload(payload.Bytes())

	assert.Equal(Payload{"foo": "bar", "baz": float64(1)}, decoded)
	assert.Nil(err)
}

// Ensures that decodePayloadSlice returns an empty slice for empty payloads.
func TestDecodePayloadSliceEmpty(t *testing.T) {
	assert := assert.New(t)
	payload := bytes.NewBufferString("")

	decoded, err := decodePayloadSlice(payload.Bytes())

	assert.Equal([]Payload{}, decoded)
	assert.Nil(err)
}

// Ensures that decodePayloadSlice returns a nil and an error for invalid JSON payloads.
func TestDecodePayloadSliceBadJSON(t *testing.T) {
	assert := assert.New(t)
	body := `[{"foo": "bar", "baz": 1`
	payload := bytes.NewBufferString(body)

	decoded, err := decodePayloadSlice(payload.Bytes())

	assert.Nil(decoded)
	assert.NotNil(err)
}

// Ensures that decodePayloadSlice returns a decoded map for JSON payloads.
func TestDecodePayloadSliceHappyPath(t *testing.T) {
	assert := assert.New(t)
	body := `[{"foo": "bar", "baz": 1}]`
	payload := bytes.NewBufferString(body)

	decoded, err := decodePayloadSlice(payload.Bytes())

	assert.Equal([]Payload{Payload{"foo": "bar", "baz": float64(1)}}, decoded)
	assert.Nil(err)
}
