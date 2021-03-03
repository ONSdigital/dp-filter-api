// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package mocks

import (
	"github.com/ONSdigital/dp-filter-api/models"
	"github.com/globalsign/mgo/bson"
	"sync"
)

var (
	lockDataStoreMockAddEventToFilterOutput       sync.RWMutex
	lockDataStoreMockAddFilter                    sync.RWMutex
	lockDataStoreMockAddFilterDimension           sync.RWMutex
	lockDataStoreMockAddFilterDimensionOption     sync.RWMutex
	lockDataStoreMockAddFilterDimensionOptions    sync.RWMutex
	lockDataStoreMockCreateFilterOutput           sync.RWMutex
	lockDataStoreMockGetFilter                    sync.RWMutex
	lockDataStoreMockGetFilterDimension           sync.RWMutex
	lockDataStoreMockGetFilterOutput              sync.RWMutex
	lockDataStoreMockRemoveFilterDimension        sync.RWMutex
	lockDataStoreMockRemoveFilterDimensionOption  sync.RWMutex
	lockDataStoreMockRemoveFilterDimensionOptions sync.RWMutex
	lockDataStoreMockUpdateFilter                 sync.RWMutex
	lockDataStoreMockUpdateFilterOutput           sync.RWMutex
)

// DataStoreMock is a mock implementation of api.DataStore.
//
//     func TestSomethingThatUsesDataStore(t *testing.T) {
//
//         // make and configure a mocked api.DataStore
//         mockedDataStore := &DataStoreMock{
//             AddEventToFilterOutputFunc: func(filterOutputID string, event *models.Event) error {
// 	               panic("mock out the AddEventToFilterOutput method")
//             },
//             AddFilterFunc: func(filter *models.Filter) (*models.Filter, error) {
// 	               panic("mock out the AddFilter method")
//             },
//             AddFilterDimensionFunc: func(filterID string, name string, options []string, dimensions []models.Dimension, timestamp bson.MongoTimestamp, eTagSelector string, currentFilter *models.Filter) (string, error) {
// 	               panic("mock out the AddFilterDimension method")
//             },
//             AddFilterDimensionOptionFunc: func(filterID string, name string, option string, timestamp bson.MongoTimestamp, eTagSelector string, currentFilter *models.Filter) (string, error) {
// 	               panic("mock out the AddFilterDimensionOption method")
//             },
//             AddFilterDimensionOptionsFunc: func(filterID string, name string, options []string, timestamp bson.MongoTimestamp, eTagSelector string, currentFilter *models.Filter) (string, error) {
// 	               panic("mock out the AddFilterDimensionOptions method")
//             },
//             CreateFilterOutputFunc: func(filter *models.Filter) error {
// 	               panic("mock out the CreateFilterOutput method")
//             },
//             GetFilterFunc: func(filterID string, eTagSelector string) (*models.Filter, error) {
// 	               panic("mock out the GetFilter method")
//             },
//             GetFilterDimensionFunc: func(filterID string, name string, eTagSelector string) (*models.Dimension, error) {
// 	               panic("mock out the GetFilterDimension method")
//             },
//             GetFilterOutputFunc: func(filterOutputID string) (*models.Filter, error) {
// 	               panic("mock out the GetFilterOutput method")
//             },
//             RemoveFilterDimensionFunc: func(filterID string, name string, timestamp bson.MongoTimestamp, eTagSelector string, currentFilter *models.Filter) (string, error) {
// 	               panic("mock out the RemoveFilterDimension method")
//             },
//             RemoveFilterDimensionOptionFunc: func(filterID string, name string, option string, timestamp bson.MongoTimestamp, eTagSelector string, currentFilter *models.Filter) (string, error) {
// 	               panic("mock out the RemoveFilterDimensionOption method")
//             },
//             RemoveFilterDimensionOptionsFunc: func(filterID string, name string, options []string, timestamp bson.MongoTimestamp, eTagSelector string, currentFilter *models.Filter) (string, error) {
// 	               panic("mock out the RemoveFilterDimensionOptions method")
//             },
//             UpdateFilterFunc: func(updatedFilter *models.Filter, timestamp bson.MongoTimestamp, eTagSelector string, currentFilter *models.Filter) (string, error) {
// 	               panic("mock out the UpdateFilter method")
//             },
//             UpdateFilterOutputFunc: func(filter *models.Filter, timestamp bson.MongoTimestamp) error {
// 	               panic("mock out the UpdateFilterOutput method")
//             },
//         }
//
//         // use mockedDataStore in code that requires api.DataStore
//         // and then make assertions.
//
//     }
type DataStoreMock struct {
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
}

// AddEventToFilterOutput calls AddEventToFilterOutputFunc.
func (mock *DataStoreMock) AddEventToFilterOutput(filterOutputID string, event *models.Event) error {
	if mock.AddEventToFilterOutputFunc == nil {
		panic("DataStoreMock.AddEventToFilterOutputFunc: method is nil but DataStore.AddEventToFilterOutput was just called")
	}
	callInfo := struct {
		FilterOutputID string
		Event          *models.Event
	}{
		FilterOutputID: filterOutputID,
		Event:          event,
	}
	lockDataStoreMockAddEventToFilterOutput.Lock()
	mock.calls.AddEventToFilterOutput = append(mock.calls.AddEventToFilterOutput, callInfo)
	lockDataStoreMockAddEventToFilterOutput.Unlock()
	return mock.AddEventToFilterOutputFunc(filterOutputID, event)
}

// AddEventToFilterOutputCalls gets all the calls that were made to AddEventToFilterOutput.
// Check the length with:
//     len(mockedDataStore.AddEventToFilterOutputCalls())
func (mock *DataStoreMock) AddEventToFilterOutputCalls() []struct {
	FilterOutputID string
	Event          *models.Event
} {
	var calls []struct {
		FilterOutputID string
		Event          *models.Event
	}
	lockDataStoreMockAddEventToFilterOutput.RLock()
	calls = mock.calls.AddEventToFilterOutput
	lockDataStoreMockAddEventToFilterOutput.RUnlock()
	return calls
}

// AddFilter calls AddFilterFunc.
func (mock *DataStoreMock) AddFilter(filter *models.Filter) (*models.Filter, error) {
	if mock.AddFilterFunc == nil {
		panic("DataStoreMock.AddFilterFunc: method is nil but DataStore.AddFilter was just called")
	}
	callInfo := struct {
		Filter *models.Filter
	}{
		Filter: filter,
	}
	lockDataStoreMockAddFilter.Lock()
	mock.calls.AddFilter = append(mock.calls.AddFilter, callInfo)
	lockDataStoreMockAddFilter.Unlock()
	return mock.AddFilterFunc(filter)
}

// AddFilterCalls gets all the calls that were made to AddFilter.
// Check the length with:
//     len(mockedDataStore.AddFilterCalls())
func (mock *DataStoreMock) AddFilterCalls() []struct {
	Filter *models.Filter
} {
	var calls []struct {
		Filter *models.Filter
	}
	lockDataStoreMockAddFilter.RLock()
	calls = mock.calls.AddFilter
	lockDataStoreMockAddFilter.RUnlock()
	return calls
}

// AddFilterDimension calls AddFilterDimensionFunc.
func (mock *DataStoreMock) AddFilterDimension(filterID string, name string, options []string, dimensions []models.Dimension, timestamp bson.MongoTimestamp, eTagSelector string, currentFilter *models.Filter) (string, error) {
	if mock.AddFilterDimensionFunc == nil {
		panic("DataStoreMock.AddFilterDimensionFunc: method is nil but DataStore.AddFilterDimension was just called")
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
	lockDataStoreMockAddFilterDimension.Lock()
	mock.calls.AddFilterDimension = append(mock.calls.AddFilterDimension, callInfo)
	lockDataStoreMockAddFilterDimension.Unlock()
	return mock.AddFilterDimensionFunc(filterID, name, options, dimensions, timestamp, eTagSelector, currentFilter)
}

// AddFilterDimensionCalls gets all the calls that were made to AddFilterDimension.
// Check the length with:
//     len(mockedDataStore.AddFilterDimensionCalls())
func (mock *DataStoreMock) AddFilterDimensionCalls() []struct {
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
	lockDataStoreMockAddFilterDimension.RLock()
	calls = mock.calls.AddFilterDimension
	lockDataStoreMockAddFilterDimension.RUnlock()
	return calls
}

// AddFilterDimensionOption calls AddFilterDimensionOptionFunc.
func (mock *DataStoreMock) AddFilterDimensionOption(filterID string, name string, option string, timestamp bson.MongoTimestamp, eTagSelector string, currentFilter *models.Filter) (string, error) {
	if mock.AddFilterDimensionOptionFunc == nil {
		panic("DataStoreMock.AddFilterDimensionOptionFunc: method is nil but DataStore.AddFilterDimensionOption was just called")
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
	lockDataStoreMockAddFilterDimensionOption.Lock()
	mock.calls.AddFilterDimensionOption = append(mock.calls.AddFilterDimensionOption, callInfo)
	lockDataStoreMockAddFilterDimensionOption.Unlock()
	return mock.AddFilterDimensionOptionFunc(filterID, name, option, timestamp, eTagSelector, currentFilter)
}

// AddFilterDimensionOptionCalls gets all the calls that were made to AddFilterDimensionOption.
// Check the length with:
//     len(mockedDataStore.AddFilterDimensionOptionCalls())
func (mock *DataStoreMock) AddFilterDimensionOptionCalls() []struct {
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
	lockDataStoreMockAddFilterDimensionOption.RLock()
	calls = mock.calls.AddFilterDimensionOption
	lockDataStoreMockAddFilterDimensionOption.RUnlock()
	return calls
}

// AddFilterDimensionOptions calls AddFilterDimensionOptionsFunc.
func (mock *DataStoreMock) AddFilterDimensionOptions(filterID string, name string, options []string, timestamp bson.MongoTimestamp, eTagSelector string, currentFilter *models.Filter) (string, error) {
	if mock.AddFilterDimensionOptionsFunc == nil {
		panic("DataStoreMock.AddFilterDimensionOptionsFunc: method is nil but DataStore.AddFilterDimensionOptions was just called")
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
	lockDataStoreMockAddFilterDimensionOptions.Lock()
	mock.calls.AddFilterDimensionOptions = append(mock.calls.AddFilterDimensionOptions, callInfo)
	lockDataStoreMockAddFilterDimensionOptions.Unlock()
	return mock.AddFilterDimensionOptionsFunc(filterID, name, options, timestamp, eTagSelector, currentFilter)
}

// AddFilterDimensionOptionsCalls gets all the calls that were made to AddFilterDimensionOptions.
// Check the length with:
//     len(mockedDataStore.AddFilterDimensionOptionsCalls())
func (mock *DataStoreMock) AddFilterDimensionOptionsCalls() []struct {
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
	lockDataStoreMockAddFilterDimensionOptions.RLock()
	calls = mock.calls.AddFilterDimensionOptions
	lockDataStoreMockAddFilterDimensionOptions.RUnlock()
	return calls
}

// CreateFilterOutput calls CreateFilterOutputFunc.
func (mock *DataStoreMock) CreateFilterOutput(filter *models.Filter) error {
	if mock.CreateFilterOutputFunc == nil {
		panic("DataStoreMock.CreateFilterOutputFunc: method is nil but DataStore.CreateFilterOutput was just called")
	}
	callInfo := struct {
		Filter *models.Filter
	}{
		Filter: filter,
	}
	lockDataStoreMockCreateFilterOutput.Lock()
	mock.calls.CreateFilterOutput = append(mock.calls.CreateFilterOutput, callInfo)
	lockDataStoreMockCreateFilterOutput.Unlock()
	return mock.CreateFilterOutputFunc(filter)
}

// CreateFilterOutputCalls gets all the calls that were made to CreateFilterOutput.
// Check the length with:
//     len(mockedDataStore.CreateFilterOutputCalls())
func (mock *DataStoreMock) CreateFilterOutputCalls() []struct {
	Filter *models.Filter
} {
	var calls []struct {
		Filter *models.Filter
	}
	lockDataStoreMockCreateFilterOutput.RLock()
	calls = mock.calls.CreateFilterOutput
	lockDataStoreMockCreateFilterOutput.RUnlock()
	return calls
}

// GetFilter calls GetFilterFunc.
func (mock *DataStoreMock) GetFilter(filterID string, eTagSelector string) (*models.Filter, error) {
	if mock.GetFilterFunc == nil {
		panic("DataStoreMock.GetFilterFunc: method is nil but DataStore.GetFilter was just called")
	}
	callInfo := struct {
		FilterID     string
		ETagSelector string
	}{
		FilterID:     filterID,
		ETagSelector: eTagSelector,
	}
	lockDataStoreMockGetFilter.Lock()
	mock.calls.GetFilter = append(mock.calls.GetFilter, callInfo)
	lockDataStoreMockGetFilter.Unlock()
	return mock.GetFilterFunc(filterID, eTagSelector)
}

// GetFilterCalls gets all the calls that were made to GetFilter.
// Check the length with:
//     len(mockedDataStore.GetFilterCalls())
func (mock *DataStoreMock) GetFilterCalls() []struct {
	FilterID     string
	ETagSelector string
} {
	var calls []struct {
		FilterID     string
		ETagSelector string
	}
	lockDataStoreMockGetFilter.RLock()
	calls = mock.calls.GetFilter
	lockDataStoreMockGetFilter.RUnlock()
	return calls
}

// GetFilterDimension calls GetFilterDimensionFunc.
func (mock *DataStoreMock) GetFilterDimension(filterID string, name string, eTagSelector string) (*models.Dimension, error) {
	if mock.GetFilterDimensionFunc == nil {
		panic("DataStoreMock.GetFilterDimensionFunc: method is nil but DataStore.GetFilterDimension was just called")
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
	lockDataStoreMockGetFilterDimension.Lock()
	mock.calls.GetFilterDimension = append(mock.calls.GetFilterDimension, callInfo)
	lockDataStoreMockGetFilterDimension.Unlock()
	return mock.GetFilterDimensionFunc(filterID, name, eTagSelector)
}

// GetFilterDimensionCalls gets all the calls that were made to GetFilterDimension.
// Check the length with:
//     len(mockedDataStore.GetFilterDimensionCalls())
func (mock *DataStoreMock) GetFilterDimensionCalls() []struct {
	FilterID     string
	Name         string
	ETagSelector string
} {
	var calls []struct {
		FilterID     string
		Name         string
		ETagSelector string
	}
	lockDataStoreMockGetFilterDimension.RLock()
	calls = mock.calls.GetFilterDimension
	lockDataStoreMockGetFilterDimension.RUnlock()
	return calls
}

// GetFilterOutput calls GetFilterOutputFunc.
func (mock *DataStoreMock) GetFilterOutput(filterOutputID string) (*models.Filter, error) {
	if mock.GetFilterOutputFunc == nil {
		panic("DataStoreMock.GetFilterOutputFunc: method is nil but DataStore.GetFilterOutput was just called")
	}
	callInfo := struct {
		FilterOutputID string
	}{
		FilterOutputID: filterOutputID,
	}
	lockDataStoreMockGetFilterOutput.Lock()
	mock.calls.GetFilterOutput = append(mock.calls.GetFilterOutput, callInfo)
	lockDataStoreMockGetFilterOutput.Unlock()
	return mock.GetFilterOutputFunc(filterOutputID)
}

// GetFilterOutputCalls gets all the calls that were made to GetFilterOutput.
// Check the length with:
//     len(mockedDataStore.GetFilterOutputCalls())
func (mock *DataStoreMock) GetFilterOutputCalls() []struct {
	FilterOutputID string
} {
	var calls []struct {
		FilterOutputID string
	}
	lockDataStoreMockGetFilterOutput.RLock()
	calls = mock.calls.GetFilterOutput
	lockDataStoreMockGetFilterOutput.RUnlock()
	return calls
}

// RemoveFilterDimension calls RemoveFilterDimensionFunc.
func (mock *DataStoreMock) RemoveFilterDimension(filterID string, name string, timestamp bson.MongoTimestamp, eTagSelector string, currentFilter *models.Filter) (string, error) {
	if mock.RemoveFilterDimensionFunc == nil {
		panic("DataStoreMock.RemoveFilterDimensionFunc: method is nil but DataStore.RemoveFilterDimension was just called")
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
	lockDataStoreMockRemoveFilterDimension.Lock()
	mock.calls.RemoveFilterDimension = append(mock.calls.RemoveFilterDimension, callInfo)
	lockDataStoreMockRemoveFilterDimension.Unlock()
	return mock.RemoveFilterDimensionFunc(filterID, name, timestamp, eTagSelector, currentFilter)
}

// RemoveFilterDimensionCalls gets all the calls that were made to RemoveFilterDimension.
// Check the length with:
//     len(mockedDataStore.RemoveFilterDimensionCalls())
func (mock *DataStoreMock) RemoveFilterDimensionCalls() []struct {
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
	lockDataStoreMockRemoveFilterDimension.RLock()
	calls = mock.calls.RemoveFilterDimension
	lockDataStoreMockRemoveFilterDimension.RUnlock()
	return calls
}

// RemoveFilterDimensionOption calls RemoveFilterDimensionOptionFunc.
func (mock *DataStoreMock) RemoveFilterDimensionOption(filterID string, name string, option string, timestamp bson.MongoTimestamp, eTagSelector string, currentFilter *models.Filter) (string, error) {
	if mock.RemoveFilterDimensionOptionFunc == nil {
		panic("DataStoreMock.RemoveFilterDimensionOptionFunc: method is nil but DataStore.RemoveFilterDimensionOption was just called")
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
	lockDataStoreMockRemoveFilterDimensionOption.Lock()
	mock.calls.RemoveFilterDimensionOption = append(mock.calls.RemoveFilterDimensionOption, callInfo)
	lockDataStoreMockRemoveFilterDimensionOption.Unlock()
	return mock.RemoveFilterDimensionOptionFunc(filterID, name, option, timestamp, eTagSelector, currentFilter)
}

// RemoveFilterDimensionOptionCalls gets all the calls that were made to RemoveFilterDimensionOption.
// Check the length with:
//     len(mockedDataStore.RemoveFilterDimensionOptionCalls())
func (mock *DataStoreMock) RemoveFilterDimensionOptionCalls() []struct {
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
	lockDataStoreMockRemoveFilterDimensionOption.RLock()
	calls = mock.calls.RemoveFilterDimensionOption
	lockDataStoreMockRemoveFilterDimensionOption.RUnlock()
	return calls
}

// RemoveFilterDimensionOptions calls RemoveFilterDimensionOptionsFunc.
func (mock *DataStoreMock) RemoveFilterDimensionOptions(filterID string, name string, options []string, timestamp bson.MongoTimestamp, eTagSelector string, currentFilter *models.Filter) (string, error) {
	if mock.RemoveFilterDimensionOptionsFunc == nil {
		panic("DataStoreMock.RemoveFilterDimensionOptionsFunc: method is nil but DataStore.RemoveFilterDimensionOptions was just called")
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
	lockDataStoreMockRemoveFilterDimensionOptions.Lock()
	mock.calls.RemoveFilterDimensionOptions = append(mock.calls.RemoveFilterDimensionOptions, callInfo)
	lockDataStoreMockRemoveFilterDimensionOptions.Unlock()
	return mock.RemoveFilterDimensionOptionsFunc(filterID, name, options, timestamp, eTagSelector, currentFilter)
}

// RemoveFilterDimensionOptionsCalls gets all the calls that were made to RemoveFilterDimensionOptions.
// Check the length with:
//     len(mockedDataStore.RemoveFilterDimensionOptionsCalls())
func (mock *DataStoreMock) RemoveFilterDimensionOptionsCalls() []struct {
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
	lockDataStoreMockRemoveFilterDimensionOptions.RLock()
	calls = mock.calls.RemoveFilterDimensionOptions
	lockDataStoreMockRemoveFilterDimensionOptions.RUnlock()
	return calls
}

// UpdateFilter calls UpdateFilterFunc.
func (mock *DataStoreMock) UpdateFilter(updatedFilter *models.Filter, timestamp bson.MongoTimestamp, eTagSelector string, currentFilter *models.Filter) (string, error) {
	if mock.UpdateFilterFunc == nil {
		panic("DataStoreMock.UpdateFilterFunc: method is nil but DataStore.UpdateFilter was just called")
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
	lockDataStoreMockUpdateFilter.Lock()
	mock.calls.UpdateFilter = append(mock.calls.UpdateFilter, callInfo)
	lockDataStoreMockUpdateFilter.Unlock()
	return mock.UpdateFilterFunc(updatedFilter, timestamp, eTagSelector, currentFilter)
}

// UpdateFilterCalls gets all the calls that were made to UpdateFilter.
// Check the length with:
//     len(mockedDataStore.UpdateFilterCalls())
func (mock *DataStoreMock) UpdateFilterCalls() []struct {
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
	lockDataStoreMockUpdateFilter.RLock()
	calls = mock.calls.UpdateFilter
	lockDataStoreMockUpdateFilter.RUnlock()
	return calls
}

// UpdateFilterOutput calls UpdateFilterOutputFunc.
func (mock *DataStoreMock) UpdateFilterOutput(filter *models.Filter, timestamp bson.MongoTimestamp) error {
	if mock.UpdateFilterOutputFunc == nil {
		panic("DataStoreMock.UpdateFilterOutputFunc: method is nil but DataStore.UpdateFilterOutput was just called")
	}
	callInfo := struct {
		Filter    *models.Filter
		Timestamp bson.MongoTimestamp
	}{
		Filter:    filter,
		Timestamp: timestamp,
	}
	lockDataStoreMockUpdateFilterOutput.Lock()
	mock.calls.UpdateFilterOutput = append(mock.calls.UpdateFilterOutput, callInfo)
	lockDataStoreMockUpdateFilterOutput.Unlock()
	return mock.UpdateFilterOutputFunc(filter, timestamp)
}

// UpdateFilterOutputCalls gets all the calls that were made to UpdateFilterOutput.
// Check the length with:
//     len(mockedDataStore.UpdateFilterOutputCalls())
func (mock *DataStoreMock) UpdateFilterOutputCalls() []struct {
	Filter    *models.Filter
	Timestamp bson.MongoTimestamp
} {
	var calls []struct {
		Filter    *models.Filter
		Timestamp bson.MongoTimestamp
	}
	lockDataStoreMockUpdateFilterOutput.RLock()
	calls = mock.calls.UpdateFilterOutput
	lockDataStoreMockUpdateFilterOutput.RUnlock()
	return calls
}
