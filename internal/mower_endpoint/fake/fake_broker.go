// Code generated by counterfeiter. DO NOT EDIT.
package fake

import (
	"sync"

	"github.com/timotto/ardumower-relay/internal/model"
	"github.com/timotto/ardumower-relay/internal/mower_endpoint"
)

type FakeBroker struct {
	OpenStub        func(model.User, model.Tunnel)
	openMutex       sync.RWMutex
	openArgsForCall []struct {
		arg1 model.User
		arg2 model.Tunnel
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeBroker) Open(arg1 model.User, arg2 model.Tunnel) {
	fake.openMutex.Lock()
	fake.openArgsForCall = append(fake.openArgsForCall, struct {
		arg1 model.User
		arg2 model.Tunnel
	}{arg1, arg2})
	stub := fake.OpenStub
	fake.recordInvocation("Open", []interface{}{arg1, arg2})
	fake.openMutex.Unlock()
	if stub != nil {
		fake.OpenStub(arg1, arg2)
	}
}

func (fake *FakeBroker) OpenCallCount() int {
	fake.openMutex.RLock()
	defer fake.openMutex.RUnlock()
	return len(fake.openArgsForCall)
}

func (fake *FakeBroker) OpenCalls(stub func(model.User, model.Tunnel)) {
	fake.openMutex.Lock()
	defer fake.openMutex.Unlock()
	fake.OpenStub = stub
}

func (fake *FakeBroker) OpenArgsForCall(i int) (model.User, model.Tunnel) {
	fake.openMutex.RLock()
	defer fake.openMutex.RUnlock()
	argsForCall := fake.openArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeBroker) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.openMutex.RLock()
	defer fake.openMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeBroker) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}

var _ mower_endpoint.Broker = new(FakeBroker)
