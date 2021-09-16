package response

import (
	"fmt"
	"sync"
)

// BasicResponse
//
// When embedded into a Response object, this wil provide error handling functionality
type BasicResponse struct {
	errString string
	code      int
}

// Failed
//
// Implements endpoint.Failer
func (b BasicResponse) Failed() error {
	if b.errString != "" {
		return b
	}
	return nil
}

// NewError
//
// Use this function when it is necessary to indicate an error result for business logic
func (b *BasicResponse) NewError(code int, format string, vars ...interface{}) {
	b.code = code
	b.errString = fmt.Sprintf(format, vars...)
}

// StatusCode
//
// Returns the status code set by the NewError method
func (b BasicResponse) StatusCode() int {
	return b.code
}

// Error
//
// Implements error interface
func (b BasicResponse) Error() string {
	return b.errString
}

type ExtendedLog interface {
	GetAll() map[string]interface{}
}

// ExpandedLogging
//
// Added to a response, should enable additional request-scoped log values
type ExpandedLogging struct {
	lvalues map[string]interface{}
	lock    sync.Mutex
}

// Log
//
// create a new log entry to be traversed later
func (l *ExpandedLogging) Log(values ...interface{}) {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.lvalues[values[0].(string)] = values[1]
}

// GetAll
//
// creates defensive copy of the underlying map
func (l *ExpandedLogging) GetAll() map[string]interface{} {
	l.lock.Lock()
	defer l.lock.Unlock()
	result := make(map[string]interface{}, len(l.lvalues))
	for k, v := range l.lvalues {
		result[k] = v
	}
	return result
}
