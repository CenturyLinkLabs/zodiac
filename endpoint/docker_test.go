package endpoint

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFromStringOrStringSlice_UnmarshalFromString(t *testing.T) {
	input := "foo bar"
	bytes, _ := json.Marshal(input)

	var result FromStringOrStringSlice
	err := json.Unmarshal(bytes, &result)

	assert.NoError(t, err)
	assert.Equal(t, []string{"foo", "bar"}, result.Value)
}

func TestFromStringOrStringSlice_UnmarshalFromStringSlice(t *testing.T) {
	input := []string{"foo", "bar"}
	bytes, _ := json.Marshal(input)

	var result FromStringOrStringSlice
	err := json.Unmarshal(bytes, &result)

	assert.NoError(t, err)
	assert.Equal(t, []string{"foo", "bar"}, result.Value)
}

func TestFromStringOrStringSlice_Marshal(t *testing.T) {
	input := FromStringOrStringSlice{
		Value: []string{"foo", "bar"},
	}

	output, err := json.Marshal(input)

	assert.NoError(t, err)
	assert.Equal(t, `["foo","bar"]`, string(output))
}

func TestFromStringOrBool_UnmarshalFromStringTrue(t *testing.T) {
	input := "TRUE"
	bytes, _ := json.Marshal(input)

	var result FromStringOrBool
	err := json.Unmarshal(bytes, &result)

	assert.NoError(t, err)
	assert.True(t, result.Value)
}

func TestFromStringOrBool_UnmarshalFromStringFalse(t *testing.T) {
	input := "false"
	bytes, _ := json.Marshal(input)

	var result FromStringOrBool
	err := json.Unmarshal(bytes, &result)

	assert.NoError(t, err)
	assert.False(t, result.Value)
}

func TestFromStringOrBool_UnmarshalFromStringNonsense(t *testing.T) {
	input := "foobar"
	bytes, _ := json.Marshal(input)

	var result FromStringOrBool
	err := json.Unmarshal(bytes, &result)

	assert.NoError(t, err)
	assert.False(t, result.Value)
}

func TestFromStringOrBool_UnmarshalFromBoolTrue(t *testing.T) {
	input := true
	bytes, _ := json.Marshal(input)

	var result FromStringOrBool
	err := json.Unmarshal(bytes, &result)

	assert.NoError(t, err)
	assert.True(t, result.Value)
}

func TestFromStringOrBool_UnmarshalFromBoolFalse(t *testing.T) {
	input := false
	bytes, _ := json.Marshal(input)

	var result FromStringOrBool
	err := json.Unmarshal(bytes, &result)

	assert.NoError(t, err)
	assert.False(t, result.Value)
}

func TestFromStringOrBool_Marshal(t *testing.T) {
	input := FromStringOrBool{Value: true}

	output, err := json.Marshal(input)

	assert.NoError(t, err)
	assert.Equal(t, "true", string(output))
}
