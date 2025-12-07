package generic

import (
	"sync"
)

type Result[T any] struct{
	Value T
	Err error
}

type JobResult struct {
    Status string      `json:"status"` // "pending", "done", "error"
    Result map[string]any `json:"result,omitempty"` // optional, usually JSON-able
    Error  string      `json:"error,omitempty"`
}

type JobStore struct {
	sync.RWMutex
	Data map[string]JobResult
}

type Effect struct {
    Type  EffectType
    ExecCommand string
    Args  []any
    Msg   string
	Fn	  func() Result[any]
}

type EffectType int

const (
    EffectDB EffectType = iota
	EffectDBQuery
    EffectLog
    EffectNotify
    EffectEmail
	EffectComplex
)
