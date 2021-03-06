// Code generated by counterfeiter. DO NOT EDIT.
package fake

import (
	"sync"

	"github.com/timotto/ardumower-relay/internal/app_endpoint"
	"github.com/timotto/ardumower-relay/internal/model"
)

type FakeBroker struct {
	FindStub        func(model.User) (model.Tunnel, bool)
	findMutex       sync.RWMutex
	findArgsForCall []struct {
		arg1 model.User
	}
	findReturns struct {
		result1 model.Tunnel
		result2 bool
	}
	findReturnsOnCall map[int]struct {
		result1 model.Tunnel
		result2 bool
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeBroker) Find(arg1 model.User) (model.Tunnel, bool) {
	fake.findMutex.Lock()
	ret, specificReturn := fake.findReturnsOnCall[len(fake.findArgsForCall)]
	fake.findArgsForCall = append(fake.findArgsForCall, struct {
		arg1 model.User
	}{arg1})
	stub := fake.FindStub
	fakeReturns := fake.findReturns
	fake.recordInvocation("Find", []interface{}{arg1})
	fake.findMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeBroker) FindCallCount() int {
	fake.findMutex.RLock()
	defer fake.findMutex.RUnlock()
	return len(fake.findArgsForCall)
}

func (fake *FakeBroker) FindCalls(stub func(model.User) (model.Tunnel, bool)) {
	fake.findMutex.Lock()
	defer fake.findMutex.Unlock()
	fake.FindStub = stub
}

func (fake *FakeBroker) FindArgsForCall(i int) model.User {
	fake.findMutex.RLock()
	defer fake.findMutex.RUnlock()
	argsForCall := fake.findArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeBroker) FindReturns(result1 model.Tunnel, result2 bool) {
	fake.findMutex.Lock()
	defer fake.findMutex.Unlock()
	fake.FindStub = nil
	fake.findReturns = struct {
		result1 model.Tunnel
		result2 bool
	}{result1, result2}
}

func (fake *FakeBroker) FindReturnsOnCall(i int, result1 model.Tunnel, result2 bool) {
	fake.findMutex.Lock()
	defer fake.findMutex.Unlock()
	fake.FindStub = nil
	if fake.findReturnsOnCall == nil {
		fake.findReturnsOnCall = make(map[int]struct {
			result1 model.Tunnel
			result2 bool
		})
	}
	fake.findReturnsOnCall[i] = struct {
		result1 model.Tunnel
		result2 bool
	}{result1, result2}
}

func (fake *FakeBroker) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.findMutex.RLock()
	defer fake.findMutex.RUnlock()
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

var _ app_endpoint.Broker = new(FakeBroker)
