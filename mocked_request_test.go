package goraphql_mock_server

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestKeyOnlyVariables checks that KeyOnlyVariables properly matches dictionaries.
func TestKeyOnlyVariables(t *testing.T) {
	type testCase struct {
		// The desired result.
		want bool
		// The structure that could have be registered to an operation.
		registeredVar KeyOnlyVariables
		// The variables that could have been received in a request.
		requestVar map[string]any
	}

	testCases := []testCase{{
		want:          false,
		registeredVar: KeyOnlyVariables{"foo", "test"},
		requestVar: map[string]any{
			"foo": struct{}{},
		},
	}, {
		want:          true,
		registeredVar: KeyOnlyVariables{"foo", "test"},
		requestVar: map[string]any{
			"foo":  struct{}{},
			"test": true,
		},
	}, {
		want:          false,
		registeredVar: KeyOnlyVariables{"foo", "test"},
		requestVar: map[string]any{
			"foo":  struct{}{},
			"test": true,
			"fail": 1,
		},
	}}

	for _, tc := range testCases {
		got := tc.registeredVar.CompareVariables(tc.requestVar)
		assert.Equal(t, tc.want, got, "var: %v", tc.registeredVar)
	}
}

// EncodedVariable implements VariableDecoder so it may match variables exactly.
type EncodedVariable struct {
	Foo string `json:"foo"`
	Num int    `json:"num"`
}

// Variable implements VariableDecoder for EncodedVariable.
func (EncodedVariable) Variable(req map[string]any) (any, bool) {
	// Check that the request's variable could potentially match this object.
	if len(req) == 0 {
		return nil, false
	} else if len(req) > 2 {
		return nil, false
	}

	// Encode the request's variables back to JSON,
	// so they be decoded into a new EncodedVariable.
	var ret EncodedVariable

	data, err := json.Marshal(req)
	if err != nil {
		return nil, false
	}

	buf := bytes.NewBuffer(data)
	dec := json.NewDecoder(buf)
	dec.DisallowUnknownFields()

	err = dec.Decode(&ret)
	return ret, err == nil
}

// TestExactVariable checks that ExactVariables properly matches
// both the key and the value in the dictionary,
// converting it as necessary.
func TestExactVariable(t *testing.T) {
	type testCase struct {
		// The desired result.
		want bool
		// The structure that could have be registered to an operation.
		registeredVar ExactVariables
		// The variables that could have been received in a request.
		requestVar map[string]any
	}

	testCases := []testCase{{
		want: false,
		registeredVar: ExactVariables{
			Variables: map[string]any{
				"foo": "1",
			},
		},
		requestVar: map[string]any{
			"foo": 1,
		},
	}, {
		want: false,
		registeredVar: ExactVariables{
			Variables: map[string]any{
				"foo": 0,
			},
		},
		requestVar: map[string]any{
			"foo": 1,
		},
	}, {
		want: true,
		registeredVar: ExactVariables{
			Variables: map[string]any{
				"foo":  1,
				"test": "str",
			},
		},
		requestVar: map[string]any{
			"foo":  1,
			"test": "str",
		},
	}, {
		want: false,
		registeredVar: ExactVariables{
			Variables: map[string]any{
				"foo":  1,
				"test": "str",
			},
		},
		requestVar: map[string]any{
			"foo":  1,
			"test": "str",
			"fail": true,
		},
	}, {
		want: false,
		registeredVar: ExactVariables{
			Variables: EncodedVariable{
				Foo: "str",
				Num: 3,
			},
		},
		requestVar: map[string]any{
			"foo": "no",
			"num": 2,
		},
	}, {
		want: true,
		registeredVar: ExactVariables{
			Variables: EncodedVariable{
				Foo: "str",
			},
		},
		requestVar: map[string]any{
			"foo": "str",
		},
	}, {
		want: true,
		registeredVar: ExactVariables{
			Variables: EncodedVariable{
				Foo: "str",
				Num: 3,
			},
		},
		requestVar: map[string]any{
			"foo": "str",
			"num": 3,
		},
	}, {
		want: false,
		registeredVar: ExactVariables{
			Variables: EncodedVariable{
				Foo: "str",
				Num: 3,
			},
		},
		requestVar: map[string]any{
			"foo":  "str",
			"num":  3,
			"fail": true,
		},
	}}

	for _, tc := range testCases {
		got := tc.registeredVar.CompareVariables(tc.requestVar)
		assert.Equal(t, tc.want, got, "var: %v", tc.registeredVar)
	}
}
