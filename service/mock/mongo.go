// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package mock

import (
	"context"
	"github.com/ONSdigital/dp-filter-api/models"
	"github.com/ONSdigital/dp-filter-api/service"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/globalsign/mgo/bson"
	"sync"
)

// Ensure, that MongoDBMock does implement service.MongoDB.
// If this is not the case, regenerate this file with moq.
var _ service.MongoDB = &MongoDBMock{}

// MongoDBMock is a mock implementation of service.MongoDB.
//
// 	func TestSomethingThatUsesMongoDB(t *testing.T) {
//
// 		// make and configure a mocked service.MongoDB
// 		mockedMongoDB := &MongoDBMock{
// 			AddEventToFilterOutputFunc: func(filterOutputID string, event *models.Event) error {
// 				panic("mock out the AddEventToFilterOutput method")
// 			},
// 			AddFilterFunc: func(filter *models.Filter) (*models.Filter, error) {
// 				panic("mock out the AddFilter method")
// 			},
// 			AddFilterDimensionFunc: func(filterID string, name string, options []string, dimensions []models.Dimension, timestamp bson.MongoTimestamp, eTagSelector string, currentFilter *models.Filter) (string, error) {
// 				panic("mock out the AddFilterDimension method")
// 			},
// 			AddFilterDimensionOptionFunc: func(filterID string, name string, option string, timestamp bson.MongoTimestamp, eTagSelector string, currentFilter *models.Filter) (string, error) {
// 				panic("mock out the AddFilterDimensionOption method")
// 			},
// 			AddFilterDimensionOptionsFunc: func(filterID string, name string, options []string, timestamp bson.MongoTimestamp, eTagSelector string, currentFilter *models.Filter) (string, error) {
// 				panic("mock out the AddFilterDimensionOptions method")
// 			},
// 			CheckerFunc: func(ctx context.Context, state *healthcheck.CheckState) error {
// 				panic("mock out the Checker method")
// 			},
// 			CloseFunc: func(ctx context.Context) error {
// 				panic("mock out the Close method")
// 			},
// 			CreateFilterOutputFunc: func(filter *models.Filter) error {
// 				panic("mock out the CreateFilterOutput method")
// 			},
// 			GetFilterFunc: func(filterID string, eTagSelector string) (*models.Filter, error) {
// 				panic("mock out the GetFilter method")
// 			},
// 			GetFilterDimensionFunc: func(filterID string, name string, eTagSelector string) (*models.Dimension, error) {
// 				panic("mock out the GetFilterDimension method")
// 			},
// 			GetFilterOutputFunc: func(filterOutputID string) (*models.Filter, error) {
// 				panic("mock out the GetFilterOutput method")
// 			},
// 			RemoveFilterDimensionFunc: func(filterID string, name string, timestamp bson.MongoTimestamp, eTagSelector string, currentFilter *models.Filter) (string, error) {
// 				panic("mock out the RemoveFilterDimension method")
// 			},
// 			RemoveFilterDimensionOptionFunc: func(filterID string, name string, option string, timestamp bson.MongoTimestamp, eTagSelector string, currentFilter *models.Filter) (string, error) {
// 				panic("mock out the RemoveFilterDimensionOption method")
// 			},
// 			RemoveFilterDimensionOptionsFunc: func(filterID string, name string, options []string, timestamp bson.MongoTimestamp, eTagSelector string, currentFilter *models.Filter) (string, error) {
// 				panic("mock out the RemoveFilterDimensionOptions method")
// 			},
// 			UpdateFilterFunc: func(updatedFilter *models.Filter, timestamp bson.MongoTimestamp, eTagSelector string, currentFilter *models.Filter) (string, error) {
// 				panic("mock out the UpdateFilter method")
// 			},
// 			UpdateFilterOutputFunc: func(filter *models.Filter, timestamp bson.MongoTimestamp) error {
// 				panic("mock out the UpdateFilterOutput method")
// 			},
// 		}
//
// 		// use mockedMongoDB in code that requires service.MongoDB
// 		// and then make assertions.
//
// 	}
type MongoDBMock struct {
	// AddEventToFilterOutputFunc mocks the AddEventToFilterOutput method.
	AddEventToFilterOutputFunc func(filterOutputID string, event *models.Event) error

	// AddFilterFunc mocks the AddFilter method.
	AddFilterFunc func(filter *models.Filter) (*models.Filter, error)

	// AddFilterDimensionFunc mocks the AddFilterDimension method.
	AddFilterDimensionFunc func(filterID string, name string, options []string, dimensions []models.Dimension, timestamp bson.MongoTimestamp, eTagSelector string, currentFilter *models.Filter) (string, error)

	// AddFilterDimensionOptionFunc mocks the AddFilterDimensionOption method.
	AddFilterDimensionOptionFunc func(filterID string, name string, option string, timestamp bson.MongoTimestamp, eTagSelector string, currentFilter *models.Filter) (string, error)

	// AddFilterDimensionOptionsFunc mocks the AddFilterDimensionOptions method.
	AddFilterDimensionOptionsFunc func(filterID string, name string, options []string, timestamp bson.MongoTimestamp, eTagSelector string, currentFilter *models.Filter) (string, error)

	// CheckerFunc mocks the Checker method.
	CheckerFunc func(ctx context.Context, state *healthcheck.CheckState) error

	// CloseFunc mocks the Close method.
	CloseFunc func(ctx context.Context) error

	// CreateFilterOutputFunc mocks the CreateFilterOutput method.
	CreateFilterOutputFunc func(filter *models.Filter) error

	// GetFilterFunc mocks the GetFilter method.
	GetFilterFunc func(filterID string, eTagSelector string) (*models.Filter, error)

	// GetFilterDimensionFunc mocks the GetFilterDimension method.
	GetFilterDimensionFunc func(filterID string, name string, eTagSelector string) (*models.Dimension, error)

	// GetFilterOutputFunc mocks the GetFilterOutput method.
	GetFilterOutputFunc func(filterOutputID string) (*models.Filter, error)

	// RemoveFilterDimensionFunc mocks the RemoveFilterDimension method.
	RemoveFilterDimensionFunc func(filterID string, name string, timestamp bson.MongoTimestamp, eTagSelector string, currentFilter *models.Filter) (string, error)

	// RemoveFilterDimensionOptionFunc mocks the RemoveFilterDimensionOption method.
	RemoveFilterDimensionOptionFunc func(filterID string, name string, option string, timestamp bson.MongoTimestamp, eTagSelector string, currentFilter *models.Filter) (string, error)

	// RemoveFilterDimensionOptionsFunc mocks the RemoveFilterDimensionOptions method.
	RemoveFilterDimensionOptionsFunc func(filterID string, name string, options []string, timestamp bson.MongoTimestamp, eTagSelector string, currentFilter *models.Filter) (string, error)

	// UpdateFilterFunc mocks the UpdateFilter method.
	UpdateFilterFunc func(updatedFilter *models.Filter, timestamp bson.MongoTimestamp, eTagSelector string, currentFilter *models.Filter) (string, error)

	// UpdateFilterOutputFunc mocks the UpdateFilterOutput method.
	UpdateFilterOutputFunc func(filter *models.Filter, timestamp bson.MongoTimestamp) error

	// calls tracks calls to the methods.
	calls struct {
		// AddEventToFilterOutput holds details about calls to the AddEventToFilterOutput method.
		AddEventToFilterOutput []struct {
			// FilterOutputID is the filterOutputID argument value.
			FilterOutputID string
			// Event is the event argument value.
			Event *models.Event
		}
		// AddFilter holds details about calls to the AddFilter method.
		AddFilter []struct {
			// Filter is the filter argument value.
			Filter *models.Filter
		}
		// AddFilterDimension holds details about calls to the AddFilterDimension method.
		AddFilterDimension []struct {
			// FilterID is the filterID argument value.
			FilterID string
			// Name is the name argument value.
			Name string
			// Options is the options argument value.
			Options []string
			// Dimensions is the dimensions argument value.
			Dimensions []models.Dimension
			// Timestamp is the timestamp argument value.
			Timestamp bson.MongoTimestamp
			// ETagSelector is the eTagSelector argument value.
			ETagSelector string
			// CurrentFilter is the currentFilter argument value.
			CurrentFilter *models.Filter
		}
		// AddFilterDimensionOption holds details about calls to the AddFilterDimensionOption method.
		AddFilterDimensionOption []struct {
			// FilterID is the filterID argument value.
			FilterID string
			// Name is the name argument value.
			Name string
			// Option is the option argument value.
			Option string
			// Timestamp is the timestamp argument value.
			Timestamp bson.MongoTimestamp
			// ETagSelector is the eTagSelector argument value.
			ETagSelector string
			// CurrentFilter is the currentFilter argument value.
			CurrentFilter *models.Filter
		}
		// AddFilterDimensionOptions holds details about calls to the AddFilterDimensionOptions method.
		AddFilterDimensionOptions []struct {
			// FilterID is the filterID argument value.
			FilterID string
			// Name is the name argument value.
			Name string
			// Options is the options argument value.
			Options []string
			// Timestamp is the timestamp argument value.
			Timestamp bson.MongoTimestamp
			// ETagSelector is the eTagSelector argument value.
			ETagSelector string
			// CurrentFilter is the currentFilter argument value.
			CurrentFilter *models.Filter
		}
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
		// CreateFilterOutput holds details about calls to the CreateFilterOutput method.
		CreateFilterOutput []struct {
			// Filter is the filter argument value.
			Filter *models.Filter
		}
		// GetFilter holds details about calls to the GetFilter method.
		GetFilter []struct {
			// FilterID is the filterID argument value.
			FilterID string
			// ETagSelector is the eTagSelector argument value.
			ETagSelector string
		}
		// GetFilterDimension holds details about calls to the GetFilterDimension method.
		GetFilterDimension []struct {
			// FilterID is the filterID argument value.
			FilterID string
			// Name is the name argument value.
			Name string
			// ETagSelector is the eTagSelector argument value.
			ETagSelector string
		}
		// GetFilterOutput holds details about calls to the GetFilterOutput method.
		GetFilterOutput []struct {
			// FilterOutputID is the filterOutputID argument value.
			FilterOutputID string
		}
		// RemoveFilterDimension holds details about calls to the RemoveFilterDimension method.
		RemoveFilterDimension []struct {
			// FilterID is the filterID argument value.
			FilterID string
			// Name is the name argument value.
			Name string
			// Timestamp is the timestamp argument value.
			Timestamp bson.MongoTimestamp
			// ETagSelector is the eTagSelector argument value.
			ETagSelector string
			// CurrentFilter is the currentFilter argument value.
			CurrentFilter *models.Filter
		}
		// RemoveFilterDimensionOption holds details about calls to the RemoveFilterDimensionOption method.
		RemoveFilterDimensionOption []struct {
			// FilterID is the filterID argument value.
			FilterID string
			// Name is the name argument value.
			Name string
			// Option is the option argument value.
			Option string
			// Timestamp is the timestamp argument value.
			Timestamp bson.MongoTimestamp
			// ETagSelector is the eTagSelector argument value.
			ETagSelector string
			// CurrentFilter is the currentFilter argument value.
			CurrentFilter *models.Filter
		}
		// RemoveFilterDimensionOptions holds details about calls to the RemoveFilterDimensionOptions method.
		RemoveFilterDimensionOptions []struct {
			// FilterID is the filterID argument value.
			FilterID string
			// Name is the name argument value.
			Name string
			// Options is the options argument value.
			Options []string
			// Timestamp is the timestamp argument value.
			Timestamp bson.MongoTimestamp
			// ETagSelector is the eTagSelector argument value.
			ETagSelector string
			// CurrentFilter is the currentFilter argument value.
			CurrentFilter *models.Filter
		}
		// UpdateFilter holds details about calls to the UpdateFilter method.
		UpdateFilter []struct {
			// UpdatedFilter is the updatedFilter argument value.
			UpdatedFilter *models.Filter
			// Timestamp is the timestamp argument value.
			Timestamp bson.MongoTimestamp
			// ETagSelector is the eTagSelector argument value.
			ETagSelector string
			// CurrentFilter is the currentFilter argument value.
			CurrentFilter *models.Filter
		}
		// UpdateFilterOutput holds details about calls to the UpdateFilterOutput method.
		UpdateFilterOutput []struct {
			// Filter is the filter argument value.
			Filter *models.Filter
			// Timestamp is the timestamp argument value.
			Timestamp bson.MongoTimestamp
		}
	}
	lockAddEventToFilterOutput       sync.RWMutex
	lockAddFilter                    sync.RWMutex
	lockAddFilterDimension           sync.RWMutex
	lockAddFilterDimensionOption     sync.RWMutex
	lockAddFilterDimensionOptions    sync.RWMutex
	lockChecker                      sync.RWMutex
	lockClose                        sync.RWMutex
	lockCreateFilterOutput           sync.RWMutex
	lockGetFilter                    sync.RWMutex
	lockGetFilterDimension           sync.RWMutex
	lockGetFilterOutput              sync.RWMutex
	lockRemoveFilterDimension        sync.RWMutex
	lockRemoveFilterDimensionOption  sync.RWMutex
	lockRemoveFilterDimensionOptions sync.RWMutex
	lockUpdateFilter                 sync.RWMutex
	lockUpdateFilterOutput           sync.RWMutex
}

// AddEventToFilterOutput calls AddEventToFilterOutputFunc.
func (mock *MongoDBMock) AddEventToFilterOutput(filterOutputID string, event *models.Event) error {
	if mock.AddEventToFilterOutputFunc == nil {
		panic("MongoDBMock.AddEventToFilterOutputFunc: method is nil but MongoDB.AddEventToFilterOutput was just called")
	}
	callInfo := struct {
		FilterOutputID string
		Event          *models.Event
	}{
		FilterOutputID: filterOutputID,
		Event:          event,
	}
	mock.lockAddEventToFilterOutput.Lock()
	mock.calls.AddEventToFilterOutput = append(mock.calls.AddEventToFilterOutput, callInfo)
	mock.lockAddEventToFilterOutput.Unlock()
	return mock.AddEventToFilterOutputFunc(filterOutputID, event)
}

// AddEventToFilterOutputCalls gets all the calls that were made to AddEventToFilterOutput.
// Check the length with:
//     len(mockedMongoDB.AddEventToFilterOutputCalls())
func (mock *MongoDBMock) AddEventToFilterOutputCalls() []struct {
	FilterOutputID string
	Event          *models.Event
} {
	var calls []struct {
		FilterOutputID string
		Event          *models.Event
	}
	mock.lockAddEventToFilterOutput.RLock()
	calls = mock.calls.AddEventToFilterOutput
	mock.lockAddEventToFilterOutput.RUnlock()
	return calls
}

// AddFilter calls AddFilterFunc.
func (mock *MongoDBMock) AddFilter(filter *models.Filter) (*models.Filter, error) {
	if mock.AddFilterFunc == nil {
		panic("MongoDBMock.AddFilterFunc: method is nil but MongoDB.AddFilter was just called")
	}
	callInfo := struct {
		Filter *models.Filter
	}{
		Filter: filter,
	}
	mock.lockAddFilter.Lock()
	mock.calls.AddFilter = append(mock.calls.AddFilter, callInfo)
	mock.lockAddFilter.Unlock()
	return mock.AddFilterFunc(filter)
}

// AddFilterCalls gets all the calls that were made to AddFilter.
// Check the length with:
//     len(mockedMongoDB.AddFilterCalls())
func (mock *MongoDBMock) AddFilterCalls() []struct {
	Filter *models.Filter
} {
	var calls []struct {
		Filter *models.Filter
	}
	mock.lockAddFilter.RLock()
	calls = mock.calls.AddFilter
	mock.lockAddFilter.RUnlock()
	return calls
}

// AddFilterDimension calls AddFilterDimensionFunc.
func (mock *MongoDBMock) AddFilterDimension(filterID string, name string, options []string, dimensions []models.Dimension, timestamp bson.MongoTimestamp, eTagSelector string, currentFilter *models.Filter) (string, error) {
	if mock.AddFilterDimensionFunc == nil {
		panic("MongoDBMock.AddFilterDimensionFunc: method is nil but MongoDB.AddFilterDimension was just called")
	}
	callInfo := struct {
		FilterID      string
		Name          string
		Options       []string
		Dimensions    []models.Dimension
		Timestamp     bson.MongoTimestamp
		ETagSelector  string
		CurrentFilter *models.Filter
	}{
		FilterID:      filterID,
		Name:          name,
		Options:       options,
		Dimensions:    dimensions,
		Timestamp:     timestamp,
		ETagSelector:  eTagSelector,
		CurrentFilter: currentFilter,
	}
	mock.lockAddFilterDimension.Lock()
	mock.calls.AddFilterDimension = append(mock.calls.AddFilterDimension, callInfo)
	mock.lockAddFilterDimension.Unlock()
	return mock.AddFilterDimensionFunc(filterID, name, options, dimensions, timestamp, eTagSelector, currentFilter)
}

// AddFilterDimensionCalls gets all the calls that were made to AddFilterDimension.
// Check the length with:
//     len(mockedMongoDB.AddFilterDimensionCalls())
func (mock *MongoDBMock) AddFilterDimensionCalls() []struct {
	FilterID      string
	Name          string
	Options       []string
	Dimensions    []models.Dimension
	Timestamp     bson.MongoTimestamp
	ETagSelector  string
	CurrentFilter *models.Filter
} {
	var calls []struct {
		FilterID      string
		Name          string
		Options       []string
		Dimensions    []models.Dimension
		Timestamp     bson.MongoTimestamp
		ETagSelector  string
		CurrentFilter *models.Filter
	}
	mock.lockAddFilterDimension.RLock()
	calls = mock.calls.AddFilterDimension
	mock.lockAddFilterDimension.RUnlock()
	return calls
}

// AddFilterDimensionOption calls AddFilterDimensionOptionFunc.
func (mock *MongoDBMock) AddFilterDimensionOption(filterID string, name string, option string, timestamp bson.MongoTimestamp, eTagSelector string, currentFilter *models.Filter) (string, error) {
	if mock.AddFilterDimensionOptionFunc == nil {
		panic("MongoDBMock.AddFilterDimensionOptionFunc: method is nil but MongoDB.AddFilterDimensionOption was just called")
	}
	callInfo := struct {
		FilterID      string
		Name          string
		Option        string
		Timestamp     bson.MongoTimestamp
		ETagSelector  string
		CurrentFilter *models.Filter
	}{
		FilterID:      filterID,
		Name:          name,
		Option:        option,
		Timestamp:     timestamp,
		ETagSelector:  eTagSelector,
		CurrentFilter: currentFilter,
	}
	mock.lockAddFilterDimensionOption.Lock()
	mock.calls.AddFilterDimensionOption = append(mock.calls.AddFilterDimensionOption, callInfo)
	mock.lockAddFilterDimensionOption.Unlock()
	return mock.AddFilterDimensionOptionFunc(filterID, name, option, timestamp, eTagSelector, currentFilter)
}

// AddFilterDimensionOptionCalls gets all the calls that were made to AddFilterDimensionOption.
// Check the length with:
//     len(mockedMongoDB.AddFilterDimensionOptionCalls())
func (mock *MongoDBMock) AddFilterDimensionOptionCalls() []struct {
	FilterID      string
	Name          string
	Option        string
	Timestamp     bson.MongoTimestamp
	ETagSelector  string
	CurrentFilter *models.Filter
} {
	var calls []struct {
		FilterID      string
		Name          string
		Option        string
		Timestamp     bson.MongoTimestamp
		ETagSelector  string
		CurrentFilter *models.Filter
	}
	mock.lockAddFilterDimensionOption.RLock()
	calls = mock.calls.AddFilterDimensionOption
	mock.lockAddFilterDimensionOption.RUnlock()
	return calls
}

// AddFilterDimensionOptions calls AddFilterDimensionOptionsFunc.
func (mock *MongoDBMock) AddFilterDimensionOptions(filterID string, name string, options []string, timestamp bson.MongoTimestamp, eTagSelector string, currentFilter *models.Filter) (string, error) {
	if mock.AddFilterDimensionOptionsFunc == nil {
		panic("MongoDBMock.AddFilterDimensionOptionsFunc: method is nil but MongoDB.AddFilterDimensionOptions was just called")
	}
	callInfo := struct {
		FilterID      string
		Name          string
		Options       []string
		Timestamp     bson.MongoTimestamp
		ETagSelector  string
		CurrentFilter *models.Filter
	}{
		FilterID:      filterID,
		Name:          name,
		Options:       options,
		Timestamp:     timestamp,
		ETagSelector:  eTagSelector,
		CurrentFilter: currentFilter,
	}
	mock.lockAddFilterDimensionOptions.Lock()
	mock.calls.AddFilterDimensionOptions = append(mock.calls.AddFilterDimensionOptions, callInfo)
	mock.lockAddFilterDimensionOptions.Unlock()
	return mock.AddFilterDimensionOptionsFunc(filterID, name, options, timestamp, eTagSelector, currentFilter)
}

// AddFilterDimensionOptionsCalls gets all the calls that were made to AddFilterDimensionOptions.
// Check the length with:
//     len(mockedMongoDB.AddFilterDimensionOptionsCalls())
func (mock *MongoDBMock) AddFilterDimensionOptionsCalls() []struct {
	FilterID      string
	Name          string
	Options       []string
	Timestamp     bson.MongoTimestamp
	ETagSelector  string
	CurrentFilter *models.Filter
} {
	var calls []struct {
		FilterID      string
		Name          string
		Options       []string
		Timestamp     bson.MongoTimestamp
		ETagSelector  string
		CurrentFilter *models.Filter
	}
	mock.lockAddFilterDimensionOptions.RLock()
	calls = mock.calls.AddFilterDimensionOptions
	mock.lockAddFilterDimensionOptions.RUnlock()
	return calls
}

// Checker calls CheckerFunc.
func (mock *MongoDBMock) Checker(ctx context.Context, state *healthcheck.CheckState) error {
	if mock.CheckerFunc == nil {
		panic("MongoDBMock.CheckerFunc: method is nil but MongoDB.Checker was just called")
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
//     len(mockedMongoDB.CheckerCalls())
func (mock *MongoDBMock) CheckerCalls() []struct {
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
func (mock *MongoDBMock) Close(ctx context.Context) error {
	if mock.CloseFunc == nil {
		panic("MongoDBMock.CloseFunc: method is nil but MongoDB.Close was just called")
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
//     len(mockedMongoDB.CloseCalls())
func (mock *MongoDBMock) CloseCalls() []struct {
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

// CreateFilterOutput calls CreateFilterOutputFunc.
func (mock *MongoDBMock) CreateFilterOutput(filter *models.Filter) error {
	if mock.CreateFilterOutputFunc == nil {
		panic("MongoDBMock.CreateFilterOutputFunc: method is nil but MongoDB.CreateFilterOutput was just called")
	}
	callInfo := struct {
		Filter *models.Filter
	}{
		Filter: filter,
	}
	mock.lockCreateFilterOutput.Lock()
	mock.calls.CreateFilterOutput = append(mock.calls.CreateFilterOutput, callInfo)
	mock.lockCreateFilterOutput.Unlock()
	return mock.CreateFilterOutputFunc(filter)
}

// CreateFilterOutputCalls gets all the calls that were made to CreateFilterOutput.
// Check the length with:
//     len(mockedMongoDB.CreateFilterOutputCalls())
func (mock *MongoDBMock) CreateFilterOutputCalls() []struct {
	Filter *models.Filter
} {
	var calls []struct {
		Filter *models.Filter
	}
	mock.lockCreateFilterOutput.RLock()
	calls = mock.calls.CreateFilterOutput
	mock.lockCreateFilterOutput.RUnlock()
	return calls
}

// GetFilter calls GetFilterFunc.
func (mock *MongoDBMock) GetFilter(filterID string, eTagSelector string) (*models.Filter, error) {
	if mock.GetFilterFunc == nil {
		panic("MongoDBMock.GetFilterFunc: method is nil but MongoDB.GetFilter was just called")
	}
	callInfo := struct {
		FilterID     string
		ETagSelector string
	}{
		FilterID:     filterID,
		ETagSelector: eTagSelector,
	}
	mock.lockGetFilter.Lock()
	mock.calls.GetFilter = append(mock.calls.GetFilter, callInfo)
	mock.lockGetFilter.Unlock()
	return mock.GetFilterFunc(filterID, eTagSelector)
}

// GetFilterCalls gets all the calls that were made to GetFilter.
// Check the length with:
//     len(mockedMongoDB.GetFilterCalls())
func (mock *MongoDBMock) GetFilterCalls() []struct {
	FilterID     string
	ETagSelector string
} {
	var calls []struct {
		FilterID     string
		ETagSelector string
	}
	mock.lockGetFilter.RLock()
	calls = mock.calls.GetFilter
	mock.lockGetFilter.RUnlock()
	return calls
}

// GetFilterDimension calls GetFilterDimensionFunc.
func (mock *MongoDBMock) GetFilterDimension(filterID string, name string, eTagSelector string) (*models.Dimension, error) {
	if mock.GetFilterDimensionFunc == nil {
		panic("MongoDBMock.GetFilterDimensionFunc: method is nil but MongoDB.GetFilterDimension was just called")
	}
	callInfo := struct {
		FilterID     string
		Name         string
		ETagSelector string
	}{
		FilterID:     filterID,
		Name:         name,
		ETagSelector: eTagSelector,
	}
	mock.lockGetFilterDimension.Lock()
	mock.calls.GetFilterDimension = append(mock.calls.GetFilterDimension, callInfo)
	mock.lockGetFilterDimension.Unlock()
	return mock.GetFilterDimensionFunc(filterID, name, eTagSelector)
}

// GetFilterDimensionCalls gets all the calls that were made to GetFilterDimension.
// Check the length with:
//     len(mockedMongoDB.GetFilterDimensionCalls())
func (mock *MongoDBMock) GetFilterDimensionCalls() []struct {
	FilterID     string
	Name         string
	ETagSelector string
} {
	var calls []struct {
		FilterID     string
		Name         string
		ETagSelector string
	}
	mock.lockGetFilterDimension.RLock()
	calls = mock.calls.GetFilterDimension
	mock.lockGetFilterDimension.RUnlock()
	return calls
}

// GetFilterOutput calls GetFilterOutputFunc.
func (mock *MongoDBMock) GetFilterOutput(filterOutputID string) (*models.Filter, error) {
	if mock.GetFilterOutputFunc == nil {
		panic("MongoDBMock.GetFilterOutputFunc: method is nil but MongoDB.GetFilterOutput was just called")
	}
	callInfo := struct {
		FilterOutputID string
	}{
		FilterOutputID: filterOutputID,
	}
	mock.lockGetFilterOutput.Lock()
	mock.calls.GetFilterOutput = append(mock.calls.GetFilterOutput, callInfo)
	mock.lockGetFilterOutput.Unlock()
	return mock.GetFilterOutputFunc(filterOutputID)
}

// GetFilterOutputCalls gets all the calls that were made to GetFilterOutput.
// Check the length with:
//     len(mockedMongoDB.GetFilterOutputCalls())
func (mock *MongoDBMock) GetFilterOutputCalls() []struct {
	FilterOutputID string
} {
	var calls []struct {
		FilterOutputID string
	}
	mock.lockGetFilterOutput.RLock()
	calls = mock.calls.GetFilterOutput
	mock.lockGetFilterOutput.RUnlock()
	return calls
}

// RemoveFilterDimension calls RemoveFilterDimensionFunc.
func (mock *MongoDBMock) RemoveFilterDimension(filterID string, name string, timestamp bson.MongoTimestamp, eTagSelector string, currentFilter *models.Filter) (string, error) {
	if mock.RemoveFilterDimensionFunc == nil {
		panic("MongoDBMock.RemoveFilterDimensionFunc: method is nil but MongoDB.RemoveFilterDimension was just called")
	}
	callInfo := struct {
		FilterID      string
		Name          string
		Timestamp     bson.MongoTimestamp
		ETagSelector  string
		CurrentFilter *models.Filter
	}{
		FilterID:      filterID,
		Name:          name,
		Timestamp:     timestamp,
		ETagSelector:  eTagSelector,
		CurrentFilter: currentFilter,
	}
	mock.lockRemoveFilterDimension.Lock()
	mock.calls.RemoveFilterDimension = append(mock.calls.RemoveFilterDimension, callInfo)
	mock.lockRemoveFilterDimension.Unlock()
	return mock.RemoveFilterDimensionFunc(filterID, name, timestamp, eTagSelector, currentFilter)
}

// RemoveFilterDimensionCalls gets all the calls that were made to RemoveFilterDimension.
// Check the length with:
//     len(mockedMongoDB.RemoveFilterDimensionCalls())
func (mock *MongoDBMock) RemoveFilterDimensionCalls() []struct {
	FilterID      string
	Name          string
	Timestamp     bson.MongoTimestamp
	ETagSelector  string
	CurrentFilter *models.Filter
} {
	var calls []struct {
		FilterID      string
		Name          string
		Timestamp     bson.MongoTimestamp
		ETagSelector  string
		CurrentFilter *models.Filter
	}
	mock.lockRemoveFilterDimension.RLock()
	calls = mock.calls.RemoveFilterDimension
	mock.lockRemoveFilterDimension.RUnlock()
	return calls
}

// RemoveFilterDimensionOption calls RemoveFilterDimensionOptionFunc.
func (mock *MongoDBMock) RemoveFilterDimensionOption(filterID string, name string, option string, timestamp bson.MongoTimestamp, eTagSelector string, currentFilter *models.Filter) (string, error) {
	if mock.RemoveFilterDimensionOptionFunc == nil {
		panic("MongoDBMock.RemoveFilterDimensionOptionFunc: method is nil but MongoDB.RemoveFilterDimensionOption was just called")
	}
	callInfo := struct {
		FilterID      string
		Name          string
		Option        string
		Timestamp     bson.MongoTimestamp
		ETagSelector  string
		CurrentFilter *models.Filter
	}{
		FilterID:      filterID,
		Name:          name,
		Option:        option,
		Timestamp:     timestamp,
		ETagSelector:  eTagSelector,
		CurrentFilter: currentFilter,
	}
	mock.lockRemoveFilterDimensionOption.Lock()
	mock.calls.RemoveFilterDimensionOption = append(mock.calls.RemoveFilterDimensionOption, callInfo)
	mock.lockRemoveFilterDimensionOption.Unlock()
	return mock.RemoveFilterDimensionOptionFunc(filterID, name, option, timestamp, eTagSelector, currentFilter)
}

// RemoveFilterDimensionOptionCalls gets all the calls that were made to RemoveFilterDimensionOption.
// Check the length with:
//     len(mockedMongoDB.RemoveFilterDimensionOptionCalls())
func (mock *MongoDBMock) RemoveFilterDimensionOptionCalls() []struct {
	FilterID      string
	Name          string
	Option        string
	Timestamp     bson.MongoTimestamp
	ETagSelector  string
	CurrentFilter *models.Filter
} {
	var calls []struct {
		FilterID      string
		Name          string
		Option        string
		Timestamp     bson.MongoTimestamp
		ETagSelector  string
		CurrentFilter *models.Filter
	}
	mock.lockRemoveFilterDimensionOption.RLock()
	calls = mock.calls.RemoveFilterDimensionOption
	mock.lockRemoveFilterDimensionOption.RUnlock()
	return calls
}

// RemoveFilterDimensionOptions calls RemoveFilterDimensionOptionsFunc.
func (mock *MongoDBMock) RemoveFilterDimensionOptions(filterID string, name string, options []string, timestamp bson.MongoTimestamp, eTagSelector string, currentFilter *models.Filter) (string, error) {
	if mock.RemoveFilterDimensionOptionsFunc == nil {
		panic("MongoDBMock.RemoveFilterDimensionOptionsFunc: method is nil but MongoDB.RemoveFilterDimensionOptions was just called")
	}
	callInfo := struct {
		FilterID      string
		Name          string
		Options       []string
		Timestamp     bson.MongoTimestamp
		ETagSelector  string
		CurrentFilter *models.Filter
	}{
		FilterID:      filterID,
		Name:          name,
		Options:       options,
		Timestamp:     timestamp,
		ETagSelector:  eTagSelector,
		CurrentFilter: currentFilter,
	}
	mock.lockRemoveFilterDimensionOptions.Lock()
	mock.calls.RemoveFilterDimensionOptions = append(mock.calls.RemoveFilterDimensionOptions, callInfo)
	mock.lockRemoveFilterDimensionOptions.Unlock()
	return mock.RemoveFilterDimensionOptionsFunc(filterID, name, options, timestamp, eTagSelector, currentFilter)
}

// RemoveFilterDimensionOptionsCalls gets all the calls that were made to RemoveFilterDimensionOptions.
// Check the length with:
//     len(mockedMongoDB.RemoveFilterDimensionOptionsCalls())
func (mock *MongoDBMock) RemoveFilterDimensionOptionsCalls() []struct {
	FilterID      string
	Name          string
	Options       []string
	Timestamp     bson.MongoTimestamp
	ETagSelector  string
	CurrentFilter *models.Filter
} {
	var calls []struct {
		FilterID      string
		Name          string
		Options       []string
		Timestamp     bson.MongoTimestamp
		ETagSelector  string
		CurrentFilter *models.Filter
	}
	mock.lockRemoveFilterDimensionOptions.RLock()
	calls = mock.calls.RemoveFilterDimensionOptions
	mock.lockRemoveFilterDimensionOptions.RUnlock()
	return calls
}

// UpdateFilter calls UpdateFilterFunc.
func (mock *MongoDBMock) UpdateFilter(updatedFilter *models.Filter, timestamp bson.MongoTimestamp, eTagSelector string, currentFilter *models.Filter) (string, error) {
	if mock.UpdateFilterFunc == nil {
		panic("MongoDBMock.UpdateFilterFunc: method is nil but MongoDB.UpdateFilter was just called")
	}
	callInfo := struct {
		UpdatedFilter *models.Filter
		Timestamp     bson.MongoTimestamp
		ETagSelector  string
		CurrentFilter *models.Filter
	}{
		UpdatedFilter: updatedFilter,
		Timestamp:     timestamp,
		ETagSelector:  eTagSelector,
		CurrentFilter: currentFilter,
	}
	mock.lockUpdateFilter.Lock()
	mock.calls.UpdateFilter = append(mock.calls.UpdateFilter, callInfo)
	mock.lockUpdateFilter.Unlock()
	return mock.UpdateFilterFunc(updatedFilter, timestamp, eTagSelector, currentFilter)
}

// UpdateFilterCalls gets all the calls that were made to UpdateFilter.
// Check the length with:
//     len(mockedMongoDB.UpdateFilterCalls())
func (mock *MongoDBMock) UpdateFilterCalls() []struct {
	UpdatedFilter *models.Filter
	Timestamp     bson.MongoTimestamp
	ETagSelector  string
	CurrentFilter *models.Filter
} {
	var calls []struct {
		UpdatedFilter *models.Filter
		Timestamp     bson.MongoTimestamp
		ETagSelector  string
		CurrentFilter *models.Filter
	}
	mock.lockUpdateFilter.RLock()
	calls = mock.calls.UpdateFilter
	mock.lockUpdateFilter.RUnlock()
	return calls
}

// UpdateFilterOutput calls UpdateFilterOutputFunc.
func (mock *MongoDBMock) UpdateFilterOutput(filter *models.Filter, timestamp bson.MongoTimestamp) error {
	if mock.UpdateFilterOutputFunc == nil {
		panic("MongoDBMock.UpdateFilterOutputFunc: method is nil but MongoDB.UpdateFilterOutput was just called")
	}
	callInfo := struct {
		Filter    *models.Filter
		Timestamp bson.MongoTimestamp
	}{
		Filter:    filter,
		Timestamp: timestamp,
	}
	mock.lockUpdateFilterOutput.Lock()
	mock.calls.UpdateFilterOutput = append(mock.calls.UpdateFilterOutput, callInfo)
	mock.lockUpdateFilterOutput.Unlock()
	return mock.UpdateFilterOutputFunc(filter, timestamp)
}

// UpdateFilterOutputCalls gets all the calls that were made to UpdateFilterOutput.
// Check the length with:
//     len(mockedMongoDB.UpdateFilterOutputCalls())
func (mock *MongoDBMock) UpdateFilterOutputCalls() []struct {
	Filter    *models.Filter
	Timestamp bson.MongoTimestamp
} {
	var calls []struct {
		Filter    *models.Filter
		Timestamp bson.MongoTimestamp
	}
	mock.lockUpdateFilterOutput.RLock()
	calls = mock.calls.UpdateFilterOutput
	mock.lockUpdateFilterOutput.RUnlock()
	return calls
}
