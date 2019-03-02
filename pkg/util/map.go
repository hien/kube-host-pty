package util

import (
	"sync"
)

type StringInterfaceMap struct {
	v     map[string]interface{}
	mutex sync.RWMutex
}

func (ss *StringInterfaceMap) Set(s string, val interface{}) {
	ss.mutex.Lock()
	defer ss.mutex.Unlock()

	if ss.v == nil {
		ss.v = make(map[string]interface{})
	}

	ss.v[s] = val
}

func (ss *StringInterfaceMap) Get(s string) (interface{}, bool) {
	ss.mutex.RLock()
	defer ss.mutex.RUnlock()

	if val, ok := ss.v[s]; ok {
		return val, true
	}

	return nil, false
}

func (ss *StringInterfaceMap) Del(s string) {
	ss.mutex.Lock()
	defer ss.mutex.Unlock()

	delete(ss.v, s)
}
