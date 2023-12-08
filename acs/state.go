package acs

import (
	"fmt"
	"sync"
)

var globalStateMap map[string]bool
var globalStateMapMutex sync.Mutex
var globalStateMapOnce sync.Once

func getStateMap() map[string]bool {
	globalStateMapOnce.Do(func() {
		globalStateMap = map[string]bool{}
	})
	return globalStateMap
}

func GetState(typ string, name string) bool {
	key := fmt.Sprintf("%v:%v", typ, name)
	m := getStateMap()
	globalStateMapMutex.Lock()
	defer globalStateMapMutex.Unlock()
	v, ok := m[key]
	if !ok {
		return false
	}
	return v
}

func SetState(typ string, name string, v bool) {
	key := fmt.Sprintf("%v:%v", typ, name)
	m := getStateMap()
	globalStateMapMutex.Lock()
	defer globalStateMapMutex.Unlock()
	m[key] = v
}
