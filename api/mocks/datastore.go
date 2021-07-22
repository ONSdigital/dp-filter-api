// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package mocks

import (
	"context"
	"github.com/ONSdigital/dp-filter-api/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"sync"
)

// DataStoreMock is a mock implementation of api.DataStore.
//
// 	func TestSomethingThatUsesDataStore(t *testing.T) {
//
// 		// make and configure a mocked api.DataStore
// 		mockedDataStore := &DataStoreMock{
// 			AddEventToFilterOutputFunc: func(ctx context.Context, filterOutputID string, event *models.Event) error {
// 				panic("mock out the AddEventToFilterOutput method")
// 			},
// 			AddFilterFunc: func(ctx context.Context, filter *models.Filter) (*models.Filter, error) {
// 				panic("mock out the AddFilter method")
// 			},
// 			AddFilterDimensionFunc: func(ctx context.Context, filterID string, name string, options []string, dimensions []models.Dimension, timestamp primitive.Timestamp, eTagSelector string, currentFilter *models.Filter) (string, error) {
// 				panic("mock out the AddFilterDimension method")
// 			},
// 			AddFilterDimensionOptionFunc: func(ctx context.Context, filterID string, name string, option string, timestamp primitive.Timestamp, eTagSelector string, currentFilter *models.Filter) (string, error) {
// 				panic("mock out the AddFilterDimensionOption method")
// 			},
// 			AddFilterDimensionOptionsFunc: func(ctx context.Context, filterID string, name string, options []string, timestamp primitive.Timestamp, eTagSelector string, currentFilter *models.Filter) (string, error) {
// 				panic("mock out the AddFilterDimensionOptions method")
// 			},
// 			CreateFilterOutputFunc: func(ctx context.Context, filter *models.Filter) error {
// 				panic("mock out the CreateFilterOutput method")
// 			},
// 			GetFilterFunc: func(ctx context.Context, filterID string, eTagSelector string) (*models.Filter, error) {
// 				panic("mock out the GetFilter method")
// 			},
// 			GetFilterDimensionFunc: func(ctx context.Context, filterID string, name string, eTagSelector string) (*models.Dimension, error) {
// 				panic("mock out the GetFilterDimension method")
// 			},
// 			GetFilterOutputFunc: func(ctx context.Context, filterOutputID string) (*models.Filter, error) {
// 				panic("mock out the GetFilterOutput method")
// 			},
// 			RemoveFilterDimensionFunc: func(ctx context.Context, filterID string, name string, timestamp primitive.Timestamp, eTagSelector string, currentFilter *models.Filter) (string, error) {
// 				panic("mock out the RemoveFilterDimension method")
// 			},
// 			RemoveFilterDimensionOptionFunc: func(ctx context.Context, filterID string, name string, option string, timestamp primitive.Timestamp, eTagSelector string, currentFilter *models.Filter) (string, error) {
// 				panic("mock out the RemoveFilterDimensionOption method")
// 			},
// 			RemoveFilterDimensionOptionsFunc: func(ctx context.Context, filterID string, name string, options []string, timestamp primitive.Timestamp, eTagSelector string, currentFilter *models.Filter) (string, error) {
// 				panic("mock out the RemoveFilterDimensionOptions method")
// 			},
// 			UpdateFilterFunc: func(ctx context.Context, updatedFilter *models.Filter, timestamp primitive.Timestamp, eTagSelector string, currentFilter *models.Filter) (string, error) {
// 				panic("mock out the UpdateFilter method")
// 			},
// 			UpdateFilterOutputFunc: func(ctx context.Context, filter *models.Filter, timestamp primitive.Timestamp) error {
// 				panic("mock out the UpdateFilterOutput method")
// 			},
// 		}
//
// 		// use mockedDataStore in code that requires api.DataStore
// 		// and then make assertions.
//
// 	}
type DataStoreMock struct {
	// AddEventToFilterOutputFunc mocks the AddEventToFilterOutput method.
	AddEventToFilterOutputFunc func(ctx context.Context, filterOutputID string, event *models.Event) error

	// AddFilterFunc mocks the AddFilter method.
	AddFilterFunc func(ctx context.Context, filter *models.Filter) (*models.Filter, error)

	// AddFilterDimensionFunc mocks the AddFilterDimension method.
	AddFilterDimensionFunc func(ctx context.Context, filterID string, name string, options []string, dimensions []models.Dimension, timestamp primitive.Timestamp, eTagSelector string, currentFilter *models.Filter) (string, error)

	// AddFilterDimensionOptionFunc mocks the AddFilterDimensionOption method.
	AddFilterDimensionOptionFunc func(ctx context.Context, filterID string, name string, option string, timestamp primitive.Timestamp, eTagSelector string, currentFilter *models.Filter) (string, error)

	// AddFilterDimensionOptionsFunc mocks the AddFilterDimensionOptions method.
	AddFilterDimensionOptionsFunc func(ctx context.Context, filterID string, name string, options []string, timestamp primitive.Timestamp, eTagSelector string, currentFilter *models.Filter) (string, error)

	// CreateFilterOutputFunc mocks the CreateFilterOutput method.
	CreateFilterOutputFunc func(ctx context.Context, filter *models.Filter) error

	// GetFilterFunc mocks the GetFilter method.
	GetFilterFunc func(ctx context.Context, filterID string, eTagSelector string) (*models.Filter, error)

	// GetFilterDimensionFunc mocks the GetFilterDimension method.
	GetFilterDimensionFunc func(ctx context.Context, filterID string, name string, eTagSelector string) (*models.Dimension, error)

	// GetFilterOutputFunc mocks the GetFilterOutput method.
	GetFilterOutputFunc func(ctx context.Context, filterOutputID string) (*models.Filter, error)

	// RemoveFilterDimensionFunc mocks the RemoveFilterDimension method.
	RemoveFilterDimensionFunc func(ctx context.Context, filterID string, name string, timestamp primitive.Timestamp, eTagSelector string, currentFilter *models.Filter) (string, error)

	// RemoveFilterDimensionOptionFunc mocks the RemoveFilterDimensionOption method.
	RemoveFilterDimensionOptionFunc func(ctx context.Context, filterID string, name string, option string, timestamp primitive.Timestamp, eTagSelector string, currentFilter *models.Filter) (string, error)

	// RemoveFilterDimensionOptionsFunc mocks the RemoveFilterDimensionOptions method.
	RemoveFilterDimensionOptionsFunc func(ctx context.Context, filterID string, name string, options []string, timestamp primitive.Timestamp, eTagSelector string, currentFilter *models.Filter) (string, error)

	// UpdateFilterFunc mocks the UpdateFilter method.
	UpdateFilterFunc func(ctx context.Context, updatedFilter *models.Filter, timestamp primitive.Timestamp, eTagSelector string, currentFilter *models.Filter) (string, error)

	// UpdateFilterOutputFunc mocks the UpdateFilterOutput method.
	UpdateFilterOutputFunc func(ctx context.Context, filter *models.Filter, timestamp primitive.Timestamp) error

	// calls tracks calls to the methods.
	calls struct {
		// AddEventToFilterOutput holds details about calls to the AddEventToFilterOutput method.
		AddEventToFilterOutput []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// FilterOutputID is the filterOutputID argument value.
			FilterOutputID string
			// Event is the event argument value.
			Event *models.Event
		}
		// AddFilter holds details about calls to the AddFilter method.
		AddFilter []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Filter is the filter argument value.
			Filter *models.Filter
		}
		// AddFilterDimension holds details about calls to the AddFilterDimension method.
		AddFilterDimension []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// FilterID is the filterID argument value.
			FilterID string
			// Name is the name argument value.
			Name string
			// Options is the options argument value.
			Options []string
			// Dimensions is the dimensions argument value.
			Dimensions []models.Dimension
			// Timestamp is the timestamp argument value.
			Timestamp primitive.Timestamp
			// ETagSelector is the eTagSelector argument value.
			ETagSelector string
			// CurrentFilter is the currentFilter argument value.
			CurrentFilter *models.Filter
		}
		// AddFilterDimensionOption holds details about calls to the AddFilterDimensionOption method.
		AddFilterDimensionOption []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// FilterID is the filterID argument value.
			FilterID string
			// Name is the name argument value.
			Name string
			// Option is the option argument value.
			Option string
			// Timestamp is the timestamp argument value.
			Timestamp primitive.Timestamp
			// ETagSelector is the eTagSelector argument value.
			ETagSelector string
			// CurrentFilter is the currentFilter argument value.
			CurrentFilter *models.Filter
		}
		// AddFilterDimensionOptions holds details about calls to the AddFilterDimensionOptions method.
		AddFilterDimensionOptions []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// FilterID is the filterID argument value.
			FilterID string
			// Name is the name argument value.
			Name string
			// Options is the options argument value.
			Options []string
			// Timestamp is the timestamp argument value.
			Timestamp primitive.Timestamp
			// ETagSelector is the eTagSelector argument value.
			ETagSelector string
			// CurrentFilter is the currentFilter argument value.
			CurrentFilter *models.Filter
		}
		// CreateFilterOutput holds details about calls to the CreateFilterOutput method.
		CreateFilterOutput []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Filter is the filter argument value.
			Filter *models.Filter
		}
		// GetFilter holds details about calls to the GetFilter method.
		GetFilter []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// FilterID is the filterID argument value.
			FilterID string
			// ETagSelector is the eTagSelector argument value.
			ETagSelector string
		}
		// GetFilterDimension holds details about calls to the GetFilterDimension method.
		GetFilterDimension []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// FilterID is the filterID argument value.
			FilterID string
			// Name is the name argument value.
			Name string
			// ETagSelector is the eTagSelector argument value.
			ETagSelector string
		}
		// GetFilterOutput holds details about calls to the GetFilterOutput method.
		GetFilterOutput []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// FilterOutputID is the filterOutputID argument value.
			FilterOutputID string
		}
		// RemoveFilterDimension holds details about calls to the RemoveFilterDimension method.
		RemoveFilterDimension []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// FilterID is the filterID argument value.
			FilterID string
			// Name is the name argument value.
			Name string
			// Timestamp is the timestamp argument value.
			Timestamp primitive.Timestamp
			// ETagSelector is the eTagSelector argument value.
			ETagSelector string
			// CurrentFilter is the currentFilter argument value.
			CurrentFilter *models.Filter
		}
		// RemoveFilterDimensionOption holds details about calls to the RemoveFilterDimensionOption method.
		RemoveFilterDimensionOption []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// FilterID is the filterID argument value.
			FilterID string
			// Name is the name argument value.
			Name string
			// Option is the option argument value.
			Option string
			// Timestamp is the timestamp argument value.
			Timestamp primitive.Timestamp
			// ETagSelector is the eTagSelector argument value.
			ETagSelector string
			// CurrentFilter is the currentFilter argument value.
			CurrentFilter *models.Filter
		}
		// RemoveFilterDimensionOptions holds details about calls to the RemoveFilterDimensionOptions method.
		RemoveFilterDimensionOptions []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// FilterID is the filterID argument value.
			FilterID string
			// Name is the name argument value.
			Name string
			// Options is the options argument value.
			Options []string
			// Timestamp is the timestamp argument value.
			Timestamp primitive.Timestamp
			// ETagSelector is the eTagSelector argument value.
			ETagSelector string
			// CurrentFilter is the currentFilter argument value.
			CurrentFilter *models.Filter
		}
		// UpdateFilter holds details about calls to the UpdateFilter method.
		UpdateFilter []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// UpdatedFilter is the updatedFilter argument value.
			UpdatedFilter *models.Filter
			// Timestamp is the timestamp argument value.
			Timestamp primitive.Timestamp
			// ETagSelector is the eTagSelector argument value.
			ETagSelector string
			// CurrentFilter is the currentFilter argument value.
			CurrentFilter *models.Filter
		}
		// UpdateFilterOutput holds details about calls to the UpdateFilterOutput method.
		UpdateFilterOutput []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Filter is the filter argument value.
			Filter *models.Filter
			// Timestamp is the timestamp argument value.
			Timestamp primitive.Timestamp
		}
	}
	lockAddEventToFilterOutput       sync.RWMutex
	lockAddFilter                    sync.RWMutex
	lockAddFilterDimension           sync.RWMutex
	lockAddFilterDimensionOption     sync.RWMutex
	lockAddFilterDimensionOptions    sync.RWMutex
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
func (mock *DataStoreMock) AddEventToFilterOutput(ctx context.Context, filterOutputID string, event *models.Event) error {
	if mock.AddEventToFilterOutputFunc == nil {
		panic("DataStoreMock.AddEventToFilterOutputFunc: method is nil but DataStore.AddEventToFilterOutput was just called")
	}
	callInfo := struct {
		Ctx            context.Context
		FilterOutputID string
		Event          *models.Event
	}{
		Ctx:            ctx,
		FilterOutputID: filterOutputID,
		Event:          event,
	}
	mock.lockAddEventToFilterOutput.Lock()
	mock.calls.AddEventToFilterOutput = append(mock.calls.AddEventToFilterOutput, callInfo)
	mock.lockAddEventToFilterOutput.Unlock()
	return mock.AddEventToFilterOutputFunc(ctx, filterOutputID, event)
}

// AddEventToFilterOutputCalls gets all the calls that were made to AddEventToFilterOutput.
// Check the length with:
//     len(mockedDataStore.AddEventToFilterOutputCalls())
func (mock *DataStoreMock) AddEventToFilterOutputCalls() []struct {
	Ctx            context.Context
	FilterOutputID string
	Event          *models.Event
} {
	var calls []struct {
		Ctx            context.Context
		FilterOutputID string
		Event          *models.Event
	}
	mock.lockAddEventToFilterOutput.RLock()
	calls = mock.calls.AddEventToFilterOutput
	mock.lockAddEventToFilterOutput.RUnlock()
	return calls
}

// AddFilter calls AddFilterFunc.
func (mock *DataStoreMock) AddFilter(ctx context.Context, filter *models.Filter) (*models.Filter, error) {
	if mock.AddFilterFunc == nil {
		panic("DataStoreMock.AddFilterFunc: method is nil but DataStore.AddFilter was just called")
	}
	callInfo := struct {
		Ctx    context.Context
		Filter *models.Filter
	}{
		Ctx:    ctx,
		Filter: filter,
	}
	mock.lockAddFilter.Lock()
	mock.calls.AddFilter = append(mock.calls.AddFilter, callInfo)
	mock.lockAddFilter.Unlock()
	return mock.AddFilterFunc(ctx, filter)
}

// AddFilterCalls gets all the calls that were made to AddFilter.
// Check the length with:
//     len(mockedDataStore.AddFilterCalls())
func (mock *DataStoreMock) AddFilterCalls() []struct {
	Ctx    context.Context
	Filter *models.Filter
} {
	var calls []struct {
		Ctx    context.Context
		Filter *models.Filter
	}
	mock.lockAddFilter.RLock()
	calls = mock.calls.AddFilter
	mock.lockAddFilter.RUnlock()
	return calls
}

// AddFilterDimension calls AddFilterDimensionFunc.
func (mock *DataStoreMock) AddFilterDimension(ctx context.Context, filterID string, name string, options []string, dimensions []models.Dimension, timestamp primitive.Timestamp, eTagSelector string, currentFilter *models.Filter) (string, error) {
	if mock.AddFilterDimensionFunc == nil {
		panic("DataStoreMock.AddFilterDimensionFunc: method is nil but DataStore.AddFilterDimension was just called")
	}
	callInfo := struct {
		Ctx           context.Context
		FilterID      string
		Name          string
		Options       []string
		Dimensions    []models.Dimension
		Timestamp     primitive.Timestamp
		ETagSelector  string
		CurrentFilter *models.Filter
	}{
		Ctx:           ctx,
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
	return mock.AddFilterDimensionFunc(ctx, filterID, name, options, dimensions, timestamp, eTagSelector, currentFilter)
}

// AddFilterDimensionCalls gets all the calls that were made to AddFilterDimension.
// Check the length with:
//     len(mockedDataStore.AddFilterDimensionCalls())
func (mock *DataStoreMock) AddFilterDimensionCalls() []struct {
	Ctx           context.Context
	FilterID      string
	Name          string
	Options       []string
	Dimensions    []models.Dimension
	Timestamp     primitive.Timestamp
	ETagSelector  string
	CurrentFilter *models.Filter
} {
	var calls []struct {
		Ctx           context.Context
		FilterID      string
		Name          string
		Options       []string
		Dimensions    []models.Dimension
		Timestamp     primitive.Timestamp
		ETagSelector  string
		CurrentFilter *models.Filter
	}
	mock.lockAddFilterDimension.RLock()
	calls = mock.calls.AddFilterDimension
	mock.lockAddFilterDimension.RUnlock()
	return calls
}

// AddFilterDimensionOption calls AddFilterDimensionOptionFunc.
func (mock *DataStoreMock) AddFilterDimensionOption(ctx context.Context, filterID string, name string, option string, timestamp primitive.Timestamp, eTagSelector string, currentFilter *models.Filter) (string, error) {
	if mock.AddFilterDimensionOptionFunc == nil {
		panic("DataStoreMock.AddFilterDimensionOptionFunc: method is nil but DataStore.AddFilterDimensionOption was just called")
	}
	callInfo := struct {
		Ctx           context.Context
		FilterID      string
		Name          string
		Option        string
		Timestamp     primitive.Timestamp
		ETagSelector  string
		CurrentFilter *models.Filter
	}{
		Ctx:           ctx,
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
	return mock.AddFilterDimensionOptionFunc(ctx, filterID, name, option, timestamp, eTagSelector, currentFilter)
}

// AddFilterDimensionOptionCalls gets all the calls that were made to AddFilterDimensionOption.
// Check the length with:
//     len(mockedDataStore.AddFilterDimensionOptionCalls())
func (mock *DataStoreMock) AddFilterDimensionOptionCalls() []struct {
	Ctx           context.Context
	FilterID      string
	Name          string
	Option        string
	Timestamp     primitive.Timestamp
	ETagSelector  string
	CurrentFilter *models.Filter
} {
	var calls []struct {
		Ctx           context.Context
		FilterID      string
		Name          string
		Option        string
		Timestamp     primitive.Timestamp
		ETagSelector  string
		CurrentFilter *models.Filter
	}
	mock.lockAddFilterDimensionOption.RLock()
	calls = mock.calls.AddFilterDimensionOption
	mock.lockAddFilterDimensionOption.RUnlock()
	return calls
}

// AddFilterDimensionOptions calls AddFilterDimensionOptionsFunc.
func (mock *DataStoreMock) AddFilterDimensionOptions(ctx context.Context, filterID string, name string, options []string, timestamp primitive.Timestamp, eTagSelector string, currentFilter *models.Filter) (string, error) {
	if mock.AddFilterDimensionOptionsFunc == nil {
		panic("DataStoreMock.AddFilterDimensionOptionsFunc: method is nil but DataStore.AddFilterDimensionOptions was just called")
	}
	callInfo := struct {
		Ctx           context.Context
		FilterID      string
		Name          string
		Options       []string
		Timestamp     primitive.Timestamp
		ETagSelector  string
		CurrentFilter *models.Filter
	}{
		Ctx:           ctx,
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
	return mock.AddFilterDimensionOptionsFunc(ctx, filterID, name, options, timestamp, eTagSelector, currentFilter)
}

// AddFilterDimensionOptionsCalls gets all the calls that were made to AddFilterDimensionOptions.
// Check the length with:
//     len(mockedDataStore.AddFilterDimensionOptionsCalls())
func (mock *DataStoreMock) AddFilterDimensionOptionsCalls() []struct {
	Ctx           context.Context
	FilterID      string
	Name          string
	Options       []string
	Timestamp     primitive.Timestamp
	ETagSelector  string
	CurrentFilter *models.Filter
} {
	var calls []struct {
		Ctx           context.Context
		FilterID      string
		Name          string
		Options       []string
		Timestamp     primitive.Timestamp
		ETagSelector  string
		CurrentFilter *models.Filter
	}
	mock.lockAddFilterDimensionOptions.RLock()
	calls = mock.calls.AddFilterDimensionOptions
	mock.lockAddFilterDimensionOptions.RUnlock()
	return calls
}

// CreateFilterOutput calls CreateFilterOutputFunc.
func (mock *DataStoreMock) CreateFilterOutput(ctx context.Context, filter *models.Filter) error {
	if mock.CreateFilterOutputFunc == nil {
		panic("DataStoreMock.CreateFilterOutputFunc: method is nil but DataStore.CreateFilterOutput was just called")
	}
	callInfo := struct {
		Ctx    context.Context
		Filter *models.Filter
	}{
		Ctx:    ctx,
		Filter: filter,
	}
	mock.lockCreateFilterOutput.Lock()
	mock.calls.CreateFilterOutput = append(mock.calls.CreateFilterOutput, callInfo)
	mock.lockCreateFilterOutput.Unlock()
	return mock.CreateFilterOutputFunc(ctx, filter)
}

// CreateFilterOutputCalls gets all the calls that were made to CreateFilterOutput.
// Check the length with:
//     len(mockedDataStore.CreateFilterOutputCalls())
func (mock *DataStoreMock) CreateFilterOutputCalls() []struct {
	Ctx    context.Context
	Filter *models.Filter
} {
	var calls []struct {
		Ctx    context.Context
		Filter *models.Filter
	}
	mock.lockCreateFilterOutput.RLock()
	calls = mock.calls.CreateFilterOutput
	mock.lockCreateFilterOutput.RUnlock()
	return calls
}

// GetFilter calls GetFilterFunc.
func (mock *DataStoreMock) GetFilter(ctx context.Context, filterID string, eTagSelector string) (*models.Filter, error) {
	if mock.GetFilterFunc == nil {
		panic("DataStoreMock.GetFilterFunc: method is nil but DataStore.GetFilter was just called")
	}
	callInfo := struct {
		Ctx          context.Context
		FilterID     string
		ETagSelector string
	}{
		Ctx:          ctx,
		FilterID:     filterID,
		ETagSelector: eTagSelector,
	}
	mock.lockGetFilter.Lock()
	mock.calls.GetFilter = append(mock.calls.GetFilter, callInfo)
	mock.lockGetFilter.Unlock()
	return mock.GetFilterFunc(ctx, filterID, eTagSelector)
}

// GetFilterCalls gets all the calls that were made to GetFilter.
// Check the length with:
//     len(mockedDataStore.GetFilterCalls())
func (mock *DataStoreMock) GetFilterCalls() []struct {
	Ctx          context.Context
	FilterID     string
	ETagSelector string
} {
	var calls []struct {
		Ctx          context.Context
		FilterID     string
		ETagSelector string
	}
	mock.lockGetFilter.RLock()
	calls = mock.calls.GetFilter
	mock.lockGetFilter.RUnlock()
	return calls
}

// GetFilterDimension calls GetFilterDimensionFunc.
func (mock *DataStoreMock) GetFilterDimension(ctx context.Context, filterID string, name string, eTagSelector string) (*models.Dimension, error) {
	if mock.GetFilterDimensionFunc == nil {
		panic("DataStoreMock.GetFilterDimensionFunc: method is nil but DataStore.GetFilterDimension was just called")
	}
	callInfo := struct {
		Ctx          context.Context
		FilterID     string
		Name         string
		ETagSelector string
	}{
		Ctx:          ctx,
		FilterID:     filterID,
		Name:         name,
		ETagSelector: eTagSelector,
	}
	mock.lockGetFilterDimension.Lock()
	mock.calls.GetFilterDimension = append(mock.calls.GetFilterDimension, callInfo)
	mock.lockGetFilterDimension.Unlock()
	return mock.GetFilterDimensionFunc(ctx, filterID, name, eTagSelector)
}

// GetFilterDimensionCalls gets all the calls that were made to GetFilterDimension.
// Check the length with:
//     len(mockedDataStore.GetFilterDimensionCalls())
func (mock *DataStoreMock) GetFilterDimensionCalls() []struct {
	Ctx          context.Context
	FilterID     string
	Name         string
	ETagSelector string
} {
	var calls []struct {
		Ctx          context.Context
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
func (mock *DataStoreMock) GetFilterOutput(ctx context.Context, filterOutputID string) (*models.Filter, error) {
	if mock.GetFilterOutputFunc == nil {
		panic("DataStoreMock.GetFilterOutputFunc: method is nil but DataStore.GetFilterOutput was just called")
	}
	callInfo := struct {
		Ctx            context.Context
		FilterOutputID string
	}{
		Ctx:            ctx,
		FilterOutputID: filterOutputID,
	}
	mock.lockGetFilterOutput.Lock()
	mock.calls.GetFilterOutput = append(mock.calls.GetFilterOutput, callInfo)
	mock.lockGetFilterOutput.Unlock()
	return mock.GetFilterOutputFunc(ctx, filterOutputID)
}

// GetFilterOutputCalls gets all the calls that were made to GetFilterOutput.
// Check the length with:
//     len(mockedDataStore.GetFilterOutputCalls())
func (mock *DataStoreMock) GetFilterOutputCalls() []struct {
	Ctx            context.Context
	FilterOutputID string
} {
	var calls []struct {
		Ctx            context.Context
		FilterOutputID string
	}
	mock.lockGetFilterOutput.RLock()
	calls = mock.calls.GetFilterOutput
	mock.lockGetFilterOutput.RUnlock()
	return calls
}

// RemoveFilterDimension calls RemoveFilterDimensionFunc.
func (mock *DataStoreMock) RemoveFilterDimension(ctx context.Context, filterID string, name string, timestamp primitive.Timestamp, eTagSelector string, currentFilter *models.Filter) (string, error) {
	if mock.RemoveFilterDimensionFunc == nil {
		panic("DataStoreMock.RemoveFilterDimensionFunc: method is nil but DataStore.RemoveFilterDimension was just called")
	}
	callInfo := struct {
		Ctx           context.Context
		FilterID      string
		Name          string
		Timestamp     primitive.Timestamp
		ETagSelector  string
		CurrentFilter *models.Filter
	}{
		Ctx:           ctx,
		FilterID:      filterID,
		Name:          name,
		Timestamp:     timestamp,
		ETagSelector:  eTagSelector,
		CurrentFilter: currentFilter,
	}
	mock.lockRemoveFilterDimension.Lock()
	mock.calls.RemoveFilterDimension = append(mock.calls.RemoveFilterDimension, callInfo)
	mock.lockRemoveFilterDimension.Unlock()
	return mock.RemoveFilterDimensionFunc(ctx, filterID, name, timestamp, eTagSelector, currentFilter)
}

// RemoveFilterDimensionCalls gets all the calls that were made to RemoveFilterDimension.
// Check the length with:
//     len(mockedDataStore.RemoveFilterDimensionCalls())
func (mock *DataStoreMock) RemoveFilterDimensionCalls() []struct {
	Ctx           context.Context
	FilterID      string
	Name          string
	Timestamp     primitive.Timestamp
	ETagSelector  string
	CurrentFilter *models.Filter
} {
	var calls []struct {
		Ctx           context.Context
		FilterID      string
		Name          string
		Timestamp     primitive.Timestamp
		ETagSelector  string
		CurrentFilter *models.Filter
	}
	mock.lockRemoveFilterDimension.RLock()
	calls = mock.calls.RemoveFilterDimension
	mock.lockRemoveFilterDimension.RUnlock()
	return calls
}

// RemoveFilterDimensionOption calls RemoveFilterDimensionOptionFunc.
func (mock *DataStoreMock) RemoveFilterDimensionOption(ctx context.Context, filterID string, name string, option string, timestamp primitive.Timestamp, eTagSelector string, currentFilter *models.Filter) (string, error) {
	if mock.RemoveFilterDimensionOptionFunc == nil {
		panic("DataStoreMock.RemoveFilterDimensionOptionFunc: method is nil but DataStore.RemoveFilterDimensionOption was just called")
	}
	callInfo := struct {
		Ctx           context.Context
		FilterID      string
		Name          string
		Option        string
		Timestamp     primitive.Timestamp
		ETagSelector  string
		CurrentFilter *models.Filter
	}{
		Ctx:           ctx,
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
	return mock.RemoveFilterDimensionOptionFunc(ctx, filterID, name, option, timestamp, eTagSelector, currentFilter)
}

// RemoveFilterDimensionOptionCalls gets all the calls that were made to RemoveFilterDimensionOption.
// Check the length with:
//     len(mockedDataStore.RemoveFilterDimensionOptionCalls())
func (mock *DataStoreMock) RemoveFilterDimensionOptionCalls() []struct {
	Ctx           context.Context
	FilterID      string
	Name          string
	Option        string
	Timestamp     primitive.Timestamp
	ETagSelector  string
	CurrentFilter *models.Filter
} {
	var calls []struct {
		Ctx           context.Context
		FilterID      string
		Name          string
		Option        string
		Timestamp     primitive.Timestamp
		ETagSelector  string
		CurrentFilter *models.Filter
	}
	mock.lockRemoveFilterDimensionOption.RLock()
	calls = mock.calls.RemoveFilterDimensionOption
	mock.lockRemoveFilterDimensionOption.RUnlock()
	return calls
}

// RemoveFilterDimensionOptions calls RemoveFilterDimensionOptionsFunc.
func (mock *DataStoreMock) RemoveFilterDimensionOptions(ctx context.Context, filterID string, name string, options []string, timestamp primitive.Timestamp, eTagSelector string, currentFilter *models.Filter) (string, error) {
	if mock.RemoveFilterDimensionOptionsFunc == nil {
		panic("DataStoreMock.RemoveFilterDimensionOptionsFunc: method is nil but DataStore.RemoveFilterDimensionOptions was just called")
	}
	callInfo := struct {
		Ctx           context.Context
		FilterID      string
		Name          string
		Options       []string
		Timestamp     primitive.Timestamp
		ETagSelector  string
		CurrentFilter *models.Filter
	}{
		Ctx:           ctx,
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
	return mock.RemoveFilterDimensionOptionsFunc(ctx, filterID, name, options, timestamp, eTagSelector, currentFilter)
}

// RemoveFilterDimensionOptionsCalls gets all the calls that were made to RemoveFilterDimensionOptions.
// Check the length with:
//     len(mockedDataStore.RemoveFilterDimensionOptionsCalls())
func (mock *DataStoreMock) RemoveFilterDimensionOptionsCalls() []struct {
	Ctx           context.Context
	FilterID      string
	Name          string
	Options       []string
	Timestamp     primitive.Timestamp
	ETagSelector  string
	CurrentFilter *models.Filter
} {
	var calls []struct {
		Ctx           context.Context
		FilterID      string
		Name          string
		Options       []string
		Timestamp     primitive.Timestamp
		ETagSelector  string
		CurrentFilter *models.Filter
	}
	mock.lockRemoveFilterDimensionOptions.RLock()
	calls = mock.calls.RemoveFilterDimensionOptions
	mock.lockRemoveFilterDimensionOptions.RUnlock()
	return calls
}

// UpdateFilter calls UpdateFilterFunc.
func (mock *DataStoreMock) UpdateFilter(ctx context.Context, updatedFilter *models.Filter, timestamp primitive.Timestamp, eTagSelector string, currentFilter *models.Filter) (string, error) {
	if mock.UpdateFilterFunc == nil {
		panic("DataStoreMock.UpdateFilterFunc: method is nil but DataStore.UpdateFilter was just called")
	}
	callInfo := struct {
		Ctx           context.Context
		UpdatedFilter *models.Filter
		Timestamp     primitive.Timestamp
		ETagSelector  string
		CurrentFilter *models.Filter
	}{
		Ctx:           ctx,
		UpdatedFilter: updatedFilter,
		Timestamp:     timestamp,
		ETagSelector:  eTagSelector,
		CurrentFilter: currentFilter,
	}
	mock.lockUpdateFilter.Lock()
	mock.calls.UpdateFilter = append(mock.calls.UpdateFilter, callInfo)
	mock.lockUpdateFilter.Unlock()
	return mock.UpdateFilterFunc(ctx, updatedFilter, timestamp, eTagSelector, currentFilter)
}

// UpdateFilterCalls gets all the calls that were made to UpdateFilter.
// Check the length with:
//     len(mockedDataStore.UpdateFilterCalls())
func (mock *DataStoreMock) UpdateFilterCalls() []struct {
	Ctx           context.Context
	UpdatedFilter *models.Filter
	Timestamp     primitive.Timestamp
	ETagSelector  string
	CurrentFilter *models.Filter
} {
	var calls []struct {
		Ctx           context.Context
		UpdatedFilter *models.Filter
		Timestamp     primitive.Timestamp
		ETagSelector  string
		CurrentFilter *models.Filter
	}
	mock.lockUpdateFilter.RLock()
	calls = mock.calls.UpdateFilter
	mock.lockUpdateFilter.RUnlock()
	return calls
}

// UpdateFilterOutput calls UpdateFilterOutputFunc.
func (mock *DataStoreMock) UpdateFilterOutput(ctx context.Context, filter *models.Filter, timestamp primitive.Timestamp) error {
	if mock.UpdateFilterOutputFunc == nil {
		panic("DataStoreMock.UpdateFilterOutputFunc: method is nil but DataStore.UpdateFilterOutput was just called")
	}
	callInfo := struct {
		Ctx       context.Context
		Filter    *models.Filter
		Timestamp primitive.Timestamp
	}{
		Ctx:       ctx,
		Filter:    filter,
		Timestamp: timestamp,
	}
	mock.lockUpdateFilterOutput.Lock()
	mock.calls.UpdateFilterOutput = append(mock.calls.UpdateFilterOutput, callInfo)
	mock.lockUpdateFilterOutput.Unlock()
	return mock.UpdateFilterOutputFunc(ctx, filter, timestamp)
}

// UpdateFilterOutputCalls gets all the calls that were made to UpdateFilterOutput.
// Check the length with:
//     len(mockedDataStore.UpdateFilterOutputCalls())
func (mock *DataStoreMock) UpdateFilterOutputCalls() []struct {
	Ctx       context.Context
	Filter    *models.Filter
	Timestamp primitive.Timestamp
} {
	var calls []struct {
		Ctx       context.Context
		Filter    *models.Filter
		Timestamp primitive.Timestamp
	}
	mock.lockUpdateFilterOutput.RLock()
	calls = mock.calls.UpdateFilterOutput
	mock.lockUpdateFilterOutput.RUnlock()
	return calls
}
