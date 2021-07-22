// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package mock

import (
	"context"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"sync"
)

var (
	lockGraphDriverMockChecker     sync.RWMutex
	lockGraphDriverMockClose       sync.RWMutex
	lockGraphDriverMockHealthcheck sync.RWMutex
)

// GraphDriverMock is a mock implementation of service.GraphDriver.
//
// 	func TestSomethingThatUsesGraphDriver(t *testing.T) {
//
// 		// make and configure a mocked service.GraphDriver
// 		mockedGraphDriver := &GraphDriverMock{
// 			CheckerFunc: func(ctx context.Context, state *healthcheck.CheckState) error {
// 				panic("mock out the Checker method")
// 			},
// 			CloseFunc: func(ctx context.Context) error {
// 				panic("mock out the Close method")
// 			},
// 			HealthcheckFunc: func() (string, error) {
// 				panic("mock out the Healthcheck method")
// 			},
// 		}
//
// 		// use mockedGraphDriver in code that requires service.GraphDriver
// 		// and then make assertions.
//
// 	}
type GraphDriverMock struct {
	// CheckerFunc mocks the Checker method.
	CheckerFunc func(ctx context.Context, state *healthcheck.CheckState) error

	// CloseFunc mocks the Close method.
	CloseFunc func(ctx context.Context) error

	// HealthcheckFunc mocks the Healthcheck method.
	HealthcheckFunc func() (string, error)

	// calls tracks calls to the methods.
	calls struct {
		// Checker holds details about calls to the Checker method.
		Checker []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// State is the state argument value.
			State *healthcheck.CheckState
		}
		// Close holds details about calls to the Close method.
		Close []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
		}
		// Healthcheck holds details about calls to the Healthcheck method.
		Healthcheck []struct {
		}
	}
	lockChecker     sync.RWMutex
	lockClose       sync.RWMutex
	lockHealthcheck sync.RWMutex
}

// Checker calls CheckerFunc.
func (mock *GraphDriverMock) Checker(ctx context.Context, state *healthcheck.CheckState) error {
	if mock.CheckerFunc == nil {
		panic("GraphDriverMock.CheckerFunc: method is nil but GraphDriver.Checker was just called")
	}
	callInfo := struct {
		Ctx   context.Context
		State *healthcheck.CheckState
	}{
		Ctx:   ctx,
		State: state,
	}
	mock.lockChecker.Lock()
	mock.calls.Checker = append(mock.calls.Checker, callInfo)
	mock.lockChecker.Unlock()
	return mock.CheckerFunc(ctx, state)
}

// CheckerCalls gets all the calls that were made to Checker.
// Check the length with:
//     len(mockedGraphDriver.CheckerCalls())
func (mock *GraphDriverMock) CheckerCalls() []struct {
	Ctx   context.Context
	State *healthcheck.CheckState
} {
	var calls []struct {
		Ctx   context.Context
		State *healthcheck.CheckState
	}
	mock.lockChecker.RLock()
	calls = mock.calls.Checker
	mock.lockChecker.RUnlock()
	return calls
}

// Close calls CloseFunc.
func (mock *GraphDriverMock) Close(ctx context.Context) error {
	if mock.CloseFunc == nil {
		panic("GraphDriverMock.CloseFunc: method is nil but GraphDriver.Close was just called")
	}
	callInfo := struct {
		Ctx context.Context
	}{
		Ctx: ctx,
	}
	mock.lockClose.Lock()
	mock.calls.Close = append(mock.calls.Close, callInfo)
	mock.lockClose.Unlock()
	return mock.CloseFunc(ctx)
}

// CloseCalls gets all the calls that were made to Close.
// Check the length with:
//     len(mockedGraphDriver.CloseCalls())
func (mock *GraphDriverMock) CloseCalls() []struct {
	Ctx context.Context
} {
	var calls []struct {
		Ctx context.Context
	}
	mock.lockClose.RLock()
	calls = mock.calls.Close
	mock.lockClose.RUnlock()
	return calls
}

// Healthcheck calls HealthcheckFunc.
func (mock *GraphDriverMock) Healthcheck() (string, error) {
	if mock.HealthcheckFunc == nil {
		panic("GraphDriverMock.HealthcheckFunc: method is nil but GraphDriver.Healthcheck was just called")
	}
	callInfo := struct {
	}{}
	mock.lockHealthcheck.Lock()
	mock.calls.Healthcheck = append(mock.calls.Healthcheck, callInfo)
	mock.lockHealthcheck.Unlock()
	return mock.HealthcheckFunc()
}

// HealthcheckCalls gets all the calls that were made to Healthcheck.
// Check the length with:
//     len(mockedGraphDriver.HealthcheckCalls())
func (mock *GraphDriverMock) HealthcheckCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockHealthcheck.RLock()
	calls = mock.calls.Healthcheck
	mock.lockHealthcheck.RUnlock()
	return calls
}
