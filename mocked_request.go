package goraphql_mock_server

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// MockedRequest manages validating whether a mocked request should be returned
// and generating its response.
type MockedRequest interface {
	// CompareVariables validates if the variables provided by the GraphQL client
	// matches this mocked request.
	CompareVariables(v map[string]any) bool

	// Response returns the object that should be sent as the response to this request.
	Response() any
}

// VariableDecoder converts a generic map of variables
// to the decoder's type,
// so the expected variables for a mocked request may be more easily compared.
type VariableDecoder interface {
	// Variable tries to convert v, the generic map of variables,
	// into the decoder's type.
	Variable(v map[string]any) (any, bool)
}

// SimpleMockedRequest implements MockedRequest using a hard-coded payload
// and comparing the key of the provided variables.
type SimpleMockedRequest struct {
	StringResponse
	KeyOnlyVariables
}

// === Partial implementations for MockedRequest ===============================

// RawResponse implements a Response() that returns exactly whatever was provided to it.
type RawResponse struct {
	Payload any
}

// Response partially implements MockedRequest for RawResponse.
func (r RawResponse) Response() any {
	return r.Payload
}

// StringResponse implements a Response() that returns this string encoded as an object.
type StringResponse string

// Response partially implements MockedRequest for StringResponse.
func (s StringResponse) Response() any {
	var data map[string]any

	err := json.Unmarshal([]byte(s), &data)
	if err != nil {
		panic(fmt.Sprintf("goraphql_mock_server: failed to encode StringResponse: %v", err))
	}

	return data
}

// KeyOnlyVariables implements a CompareVariables()
// that checks that there are no variables in the payload.
type NoVariable struct{}

// CompareVariables partially implements MockedRequest for NoVariable.
func (nv NoVariable) CompareVariables(got map[string]any) bool {
	return len(got) == 0
}

// KeyOnlyVariables implements a CompareVariables()
// that checks if the variable's keys exactly matches this object.
type KeyOnlyVariables []string

// CompareVariables partially implements MockedRequest for KeyOnlyVariables.
func (kv KeyOnlyVariables) CompareVariables(got map[string]any) bool {
	want := make(map[string]bool)
	for _, k := range kv {
		want[k] = true
	}

	if len(want) != len(got) {
		return false
	}

	for k := range got {
		if !want[k] {
			return false
		}
	}

	return true
}

// ExactVariables implements a CompareVariables()
// that checks if the variable's matches exactly whatever was provided,
// including the type of each variable.
type ExactVariables struct {
	// An object that should match the request's variables.
	// If this object implements VariableDecoder,
	// then Variable() is called on the request's variables
	// to convert it to the same type as this object before comparing them.
	Variables any
}

// CompareVariables partially implements MockedRequest for ExactVariables.
func (ev ExactVariables) CompareVariables(reqVar map[string]any) bool {
	var got any

	decoder, ok := ev.Variables.(VariableDecoder)
	if ok {
		got, ok = decoder.Variable(reqVar)
		if !ok {
			return false
		}
	} else {
		got = reqVar
	}

	return reflect.DeepEqual(ev.Variables, got)
}
