package glob

import (
	"sync"
)

var ready bool
var readyMu sync.RWMutex

var globErrs map[string]*error
var globErrsMu sync.RWMutex

func init() {
	ready = false
	globErrs = make(map[string]*error)
}

// SetReady marks the exporter as ready to serve metrics
func SetReady() {
	readyMu.Lock()
	defer readyMu.Unlock()
	ready = true
}

// SetError sets an error with the given key to display in the status page, marks the exporter as unhealthy
//
// call with nil to clear the error for that key
func SetError(key string, error *error) {
	globErrsMu.Lock()
	defer globErrsMu.Unlock()
	if error == nil {
		delete(globErrs, key)
	} else {
		globErrs[key] = error
	}
}

func GetErrorDescriptions() map[string]string {
	globErrsMu.RLock()
	defer globErrsMu.RUnlock()
	list := make(map[string]string, len(globErrs))
	for name, err := range globErrs {
		list[name] = (*err).Error()
	}
	return list
}

// IsReady returns true if the exporter is ready to serve metrics
func IsReady() bool {
	readyMu.RLock()
	defer readyMu.RUnlock()
	return ready
}

func IsOk() bool {
	globErrsMu.RLock()
	defer globErrsMu.RUnlock()
	return len(globErrs) == 0
}
