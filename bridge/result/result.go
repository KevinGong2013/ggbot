package result

import (
	"encoding/json"
	"fmt"
)

// Result ...
type Result struct {
	Value []byte
	Err   string
}

// IsSuccess ...
func (result *Result) IsSuccess() bool {
	return len(result.Err) == 0
}

// IsFailure ...
func (result *Result) IsFailure() bool {
	return !result.IsSuccess()
}

// NewResultWithError create return value with error
func NewResultWithError(err string) *Result {
	return &Result{nil, err}
}

// NewResultWithValue create return value with value
func NewResultWithValue(value []byte) *Result {
	return &Result{value, ``}
}

// NewResultWithMap ...
func NewResultWithMap(m map[string]interface{}) *Result {
	by, err := json.Marshal(m)
	if err != nil {
		return NewResultWithError(`marshal m error`)
	}
	return NewResultWithValue(by)
}

// String ..
func (result *Result) String() string {
	if result.IsSuccess() {
		return fmt.Sprintf(`[SUCCESSED] %v`, string(result.Value))
	}
	return fmt.Sprintf(`[FAILED] %v`, result.Err)
}
