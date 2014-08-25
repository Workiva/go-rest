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

	decoded, err := decodePayload(payload, 0)

	assert.Equal(Payload{}, decoded)
	assert.Nil(err)
}

// Ensures that decodePayload returns a nil and an error for invalid JSON payloads.
func TestDecodePayloadBadJSON(t *testing.T) {
	assert := assert.New(t)
	body := `{"foo": "bar", "baz": 1`
	payload := bytes.NewBufferString(body)

	decoded, err := decodePayload(payload, int64(len(body)))

	assert.Nil(decoded)
	assert.NotNil(err)
}

// Ensures that decodePayload returns a decoded map for JSON payloads.
func TestDecodePayloadHappyPath(t *testing.T) {
	assert := assert.New(t)
	body := `{"foo": "bar", "baz": 1}`
	payload := bytes.NewBufferString(body)

	decoded, err := decodePayload(payload, int64(len(body)))

	assert.Equal(Payload{"foo": "bar", "baz": float64(1)}, decoded)
	assert.Nil(err)
}

// Ensures that decodePayloadSlice returns an empty slice for empty payloads.
func TestDecodePayloadSliceEmpty(t *testing.T) {
	assert := assert.New(t)
	payload := bytes.NewBufferString("")

	decoded, err := decodePayloadSlice(payload, 0)

	assert.Equal([]Payload{}, decoded)
	assert.Nil(err)
}

// Ensures that decodePayloadSlice returns a nil and an error for invalid JSON payloads.
func TestDecodePayloadSliceBadJSON(t *testing.T) {
	assert := assert.New(t)
	body := `[{"foo": "bar", "baz": 1`
	payload := bytes.NewBufferString(body)

	decoded, err := decodePayloadSlice(payload, int64(len(body)))

	assert.Nil(decoded)
	assert.NotNil(err)
}

// Ensures that decodePayloadSlice returns a decoded map for JSON payloads.
func TestDecodePayloadSliceHappyPath(t *testing.T) {
	assert := assert.New(t)
	body := `[{"foo": "bar", "baz": 1}]`
	payload := bytes.NewBufferString(body)

	decoded, err := decodePayloadSlice(payload, int64(len(body)))

	assert.Equal([]Payload{Payload{"foo": "bar", "baz": float64(1)}}, decoded)
	assert.Nil(err)
}
