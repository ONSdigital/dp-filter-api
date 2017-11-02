// Code generated by moq; DO NOT EDIT
// github.com/matryer/moq

package datastoretest

import (
	"github.com/ONSdigital/dp-filter-api/models"
	"sync"
)

var (
	lockDataStoreMockAddFilter                   sync.RWMutex
	lockDataStoreMockAddFilterDimension          sync.RWMutex
	lockDataStoreMockAddFilterDimensionOption    sync.RWMutex
	lockDataStoreMockCreateFilterOutput          sync.RWMutex
	lockDataStoreMockGetFilter                   sync.RWMutex
	lockDataStoreMockGetFilterDimension          sync.RWMutex
	lockDataStoreMockGetFilterDimensionOption    sync.RWMutex
	lockDataStoreMockGetFilterDimensionOptions   sync.RWMutex
	lockDataStoreMockGetFilterDimensions         sync.RWMutex
	lockDataStoreMockGetFilterOutput             sync.RWMutex
	lockDataStoreMockRemoveFilterDimension       sync.RWMutex
	lockDataStoreMockRemoveFilterDimensionOption sync.RWMutex
	lockDataStoreMockUpdateFilter                sync.RWMutex
	lockDataStoreMockUpdateFilterOutput          sync.RWMutex
)

// DataStoreMock is a mock implementation of DataStore.
//
//     func TestSomethingThatUsesDataStore(t *testing.T) {
//
//         // make and configure a mocked DataStore
//         mockedDataStore := &DataStoreMock{
//             AddFilterFunc: func(host string, filter *models.Filter) (*models.Filter, error) {
// 	               panic("TODO: mock out the AddFilter method")
//             },
//             AddFilterDimensionFunc: func(in1 *models.AddDimension) error {
// 	               panic("TODO: mock out the AddFilterDimension method")
//             },
//             AddFilterDimensionOptionFunc: func(in1 *models.AddDimensionOption) error {
// 	               panic("TODO: mock out the AddFilterDimensionOption method")
//             },
//             CreateFilterOutputFunc: func(filter *models.Filter) error {
// 	               panic("TODO: mock out the CreateFilterOutput method")
//             },
//             GetFilterFunc: func(filterID string) (*models.Filter, error) {
// 	               panic("TODO: mock out the GetFilter method")
//             },
//             GetFilterDimensionFunc: func(filterID string, name string) error {
// 	               panic("TODO: mock out the GetFilterDimension method")
//             },
//             GetFilterDimensionOptionFunc: func(filterID string, name string, option string) error {
// 	               panic("TODO: mock out the GetFilterDimensionOption method")
//             },
//             GetFilterDimensionOptionsFunc: func(filterID string, name string) ([]models.DimensionOption, error) {
// 	               panic("TODO: mock out the GetFilterDimensionOptions method")
//             },
//             GetFilterDimensionsFunc: func(filterID string) ([]models.Dimension, error) {
// 	               panic("TODO: mock out the GetFilterDimensions method")
//             },
//             GetFilterOutputFunc: func(filterOutputID string) (*models.Filter, error) {
// 	               panic("TODO: mock out the GetFilterOutput method")
//             },
//             RemoveFilterDimensionFunc: func(filterID string, name string) error {
// 	               panic("TODO: mock out the RemoveFilterDimension method")
//             },
//             RemoveFilterDimensionOptionFunc: func(filterID string, name string, option string) error {
// 	               panic("TODO: mock out the RemoveFilterDimensionOption method")
//             },
//             UpdateFilterFunc: func(filter *models.Filter) error {
// 	               panic("TODO: mock out the UpdateFilter method")
//             },
//             UpdateFilterOutputFunc: func(filterOutput *models.Filter) error {
// 	               panic("TODO: mock out the UpdateFilterOutput method")
//             },
//         }
//
//         // TODO: use mockedDataStore in code that requires DataStore
//         //       and then make assertions.
//
//     }
type DataStoreMock struct {
	// AddFilterFunc mocks the AddFilter method.
	AddFilterFunc func(host string, filter *models.Filter) (*models.Filter, error)

	// AddFilterDimensionFunc mocks the AddFilterDimension method.
	AddFilterDimensionFunc func(in1 *models.AddDimension) error

	// AddFilterDimensionOptionFunc mocks the AddFilterDimensionOption method.
	AddFilterDimensionOptionFunc func(in1 *models.AddDimensionOption) error

	// CreateFilterOutputFunc mocks the CreateFilterOutput method.
	CreateFilterOutputFunc func(filter *models.Filter) error

	// GetFilterFunc mocks the GetFilter method.
	GetFilterFunc func(filterID string) (*models.Filter, error)

	// GetFilterDimensionFunc mocks the GetFilterDimension method.
	GetFilterDimensionFunc func(filterID string, name string) error

	// GetFilterDimensionOptionFunc mocks the GetFilterDimensionOption method.
	GetFilterDimensionOptionFunc func(filterID string, name string, option string) error

	// GetFilterDimensionOptionsFunc mocks the GetFilterDimensionOptions method.
	GetFilterDimensionOptionsFunc func(filterID string, name string) ([]models.DimensionOption, error)

	// GetFilterDimensionsFunc mocks the GetFilterDimensions method.
	GetFilterDimensionsFunc func(filterID string) ([]models.Dimension, error)

	// GetFilterOutputFunc mocks the GetFilterOutput method.
	GetFilterOutputFunc func(filterOutputID string) (*models.Filter, error)

	// RemoveFilterDimensionFunc mocks the RemoveFilterDimension method.
	RemoveFilterDimensionFunc func(filterID string, name string) error

	// RemoveFilterDimensionOptionFunc mocks the RemoveFilterDimensionOption method.
	RemoveFilterDimensionOptionFunc func(filterID string, name string, option string) error

	// UpdateFilterFunc mocks the UpdateFilter method.
	UpdateFilterFunc func(filter *models.Filter) error

	// UpdateFilterOutputFunc mocks the UpdateFilterOutput method.
	UpdateFilterOutputFunc func(filterOutput *models.Filter) error

	// calls tracks calls to the methods.
	calls struct {
		// AddFilter holds details about calls to the AddFilter method.
		AddFilter []struct {
			// Host is the host argument value.
			Host string
			// Filter is the filter argument value.
			Filter *models.Filter
		}
		// AddFilterDimension holds details about calls to the AddFilterDimension method.
		AddFilterDimension []struct {
			// In1 is the in1 argument value.
			In1 *models.AddDimension
		}
		// AddFilterDimensionOption holds details about calls to the AddFilterDimensionOption method.
		AddFilterDimensionOption []struct {
			// In1 is the in1 argument value.
			In1 *models.AddDimensionOption
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
		}
		// GetFilterDimension holds details about calls to the GetFilterDimension method.
		GetFilterDimension []struct {
			// FilterID is the filterID argument value.
			FilterID string
			// Name is the name argument value.
			Name string
		}
		// GetFilterDimensionOption holds details about calls to the GetFilterDimensionOption method.
		GetFilterDimensionOption []struct {
			// FilterID is the filterID argument value.
			FilterID string
			// Name is the name argument value.
			Name string
			// Option is the option argument value.
			Option string
		}
		// GetFilterDimensionOptions holds details about calls to the GetFilterDimensionOptions method.
		GetFilterDimensionOptions []struct {
			// FilterID is the filterID argument value.
			FilterID string
			// Name is the name argument value.
			Name string
		}
		// GetFilterDimensions holds details about calls to the GetFilterDimensions method.
		GetFilterDimensions []struct {
			// FilterID is the filterID argument value.
			FilterID string
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
		}
		// RemoveFilterDimensionOption holds details about calls to the RemoveFilterDimensionOption method.
		RemoveFilterDimensionOption []struct {
			// FilterID is the filterID argument value.
			FilterID string
			// Name is the name argument value.
			Name string
			// Option is the option argument value.
			Option string
		}
		// UpdateFilter holds details about calls to the UpdateFilter method.
		UpdateFilter []struct {
			// Filter is the filter argument value.
			Filter *models.Filter
		}
		// UpdateFilterOutput holds details about calls to the UpdateFilterOutput method.
		UpdateFilterOutput []struct {
			// FilterOutput is the filterOutput argument value.
			FilterOutput *models.Filter
		}
	}
}

// AddFilter calls AddFilterFunc.
func (mock *DataStoreMock) AddFilter(host string, filter *models.Filter) (*models.Filter, error) {
	if mock.AddFilterFunc == nil {
		panic("moq: DataStoreMock.AddFilterFunc is nil but DataStore.AddFilter was just called")
	}
	callInfo := struct {
		Host   string
		Filter *models.Filter
	}{
		Host:   host,
		Filter: filter,
	}
	lockDataStoreMockAddFilter.Lock()
	mock.calls.AddFilter = append(mock.calls.AddFilter, callInfo)
	lockDataStoreMockAddFilter.Unlock()
	return mock.AddFilterFunc(host, filter)
}

// AddFilterCalls gets all the calls that were made to AddFilter.
// Check the length with:
//     len(mockedDataStore.AddFilterCalls())
func (mock *DataStoreMock) AddFilterCalls() []struct {
	Host   string
	Filter *models.Filter
} {
	var calls []struct {
		Host   string
		Filter *models.Filter
	}
	lockDataStoreMockAddFilter.RLock()
	calls = mock.calls.AddFilter
	lockDataStoreMockAddFilter.RUnlock()
	return calls
}

// AddFilterDimension calls AddFilterDimensionFunc.
func (mock *DataStoreMock) AddFilterDimension(in1 *models.AddDimension) error {
	if mock.AddFilterDimensionFunc == nil {
		panic("moq: DataStoreMock.AddFilterDimensionFunc is nil but DataStore.AddFilterDimension was just called")
	}
	callInfo := struct {
		In1 *models.AddDimension
	}{
		In1: in1,
	}
	lockDataStoreMockAddFilterDimension.Lock()
	mock.calls.AddFilterDimension = append(mock.calls.AddFilterDimension, callInfo)
	lockDataStoreMockAddFilterDimension.Unlock()
	return mock.AddFilterDimensionFunc(in1)
}

// AddFilterDimensionCalls gets all the calls that were made to AddFilterDimension.
// Check the length with:
//     len(mockedDataStore.AddFilterDimensionCalls())
func (mock *DataStoreMock) AddFilterDimensionCalls() []struct {
	In1 *models.AddDimension
} {
	var calls []struct {
		In1 *models.AddDimension
	}
	lockDataStoreMockAddFilterDimension.RLock()
	calls = mock.calls.AddFilterDimension
	lockDataStoreMockAddFilterDimension.RUnlock()
	return calls
}

// AddFilterDimensionOption calls AddFilterDimensionOptionFunc.
func (mock *DataStoreMock) AddFilterDimensionOption(in1 *models.AddDimensionOption) error {
	if mock.AddFilterDimensionOptionFunc == nil {
		panic("moq: DataStoreMock.AddFilterDimensionOptionFunc is nil but DataStore.AddFilterDimensionOption was just called")
	}
	callInfo := struct {
		In1 *models.AddDimensionOption
	}{
		In1: in1,
	}
	lockDataStoreMockAddFilterDimensionOption.Lock()
	mock.calls.AddFilterDimensionOption = append(mock.calls.AddFilterDimensionOption, callInfo)
	lockDataStoreMockAddFilterDimensionOption.Unlock()
	return mock.AddFilterDimensionOptionFunc(in1)
}

// AddFilterDimensionOptionCalls gets all the calls that were made to AddFilterDimensionOption.
// Check the length with:
//     len(mockedDataStore.AddFilterDimensionOptionCalls())
func (mock *DataStoreMock) AddFilterDimensionOptionCalls() []struct {
	In1 *models.AddDimensionOption
} {
	var calls []struct {
		In1 *models.AddDimensionOption
	}
	lockDataStoreMockAddFilterDimensionOption.RLock()
	calls = mock.calls.AddFilterDimensionOption
	lockDataStoreMockAddFilterDimensionOption.RUnlock()
	return calls
}

// CreateFilterOutput calls CreateFilterOutputFunc.
func (mock *DataStoreMock) CreateFilterOutput(filter *models.Filter) error {
	if mock.CreateFilterOutputFunc == nil {
		panic("moq: DataStoreMock.CreateFilterOutputFunc is nil but DataStore.CreateFilterOutput was just called")
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
func (mock *DataStoreMock) GetFilter(filterID string) (*models.Filter, error) {
	if mock.GetFilterFunc == nil {
		panic("moq: DataStoreMock.GetFilterFunc is nil but DataStore.GetFilter was just called")
	}
	callInfo := struct {
		FilterID string
	}{
		FilterID: filterID,
	}
	lockDataStoreMockGetFilter.Lock()
	mock.calls.GetFilter = append(mock.calls.GetFilter, callInfo)
	lockDataStoreMockGetFilter.Unlock()
	return mock.GetFilterFunc(filterID)
}

// GetFilterCalls gets all the calls that were made to GetFilter.
// Check the length with:
//     len(mockedDataStore.GetFilterCalls())
func (mock *DataStoreMock) GetFilterCalls() []struct {
	FilterID string
} {
	var calls []struct {
		FilterID string
	}
	lockDataStoreMockGetFilter.RLock()
	calls = mock.calls.GetFilter
	lockDataStoreMockGetFilter.RUnlock()
	return calls
}

// GetFilterDimension calls GetFilterDimensionFunc.
func (mock *DataStoreMock) GetFilterDimension(filterID string, name string) error {
	if mock.GetFilterDimensionFunc == nil {
		panic("moq: DataStoreMock.GetFilterDimensionFunc is nil but DataStore.GetFilterDimension was just called")
	}
	callInfo := struct {
		FilterID string
		Name     string
	}{
		FilterID: filterID,
		Name:     name,
	}
	lockDataStoreMockGetFilterDimension.Lock()
	mock.calls.GetFilterDimension = append(mock.calls.GetFilterDimension, callInfo)
	lockDataStoreMockGetFilterDimension.Unlock()
	return mock.GetFilterDimensionFunc(filterID, name)
}

// GetFilterDimensionCalls gets all the calls that were made to GetFilterDimension.
// Check the length with:
//     len(mockedDataStore.GetFilterDimensionCalls())
func (mock *DataStoreMock) GetFilterDimensionCalls() []struct {
	FilterID string
	Name     string
} {
	var calls []struct {
		FilterID string
		Name     string
	}
	lockDataStoreMockGetFilterDimension.RLock()
	calls = mock.calls.GetFilterDimension
	lockDataStoreMockGetFilterDimension.RUnlock()
	return calls
}

// GetFilterDimensionOption calls GetFilterDimensionOptionFunc.
func (mock *DataStoreMock) GetFilterDimensionOption(filterID string, name string, option string) error {
	if mock.GetFilterDimensionOptionFunc == nil {
		panic("moq: DataStoreMock.GetFilterDimensionOptionFunc is nil but DataStore.GetFilterDimensionOption was just called")
	}
	callInfo := struct {
		FilterID string
		Name     string
		Option   string
	}{
		FilterID: filterID,
		Name:     name,
		Option:   option,
	}
	lockDataStoreMockGetFilterDimensionOption.Lock()
	mock.calls.GetFilterDimensionOption = append(mock.calls.GetFilterDimensionOption, callInfo)
	lockDataStoreMockGetFilterDimensionOption.Unlock()
	return mock.GetFilterDimensionOptionFunc(filterID, name, option)
}

// GetFilterDimensionOptionCalls gets all the calls that were made to GetFilterDimensionOption.
// Check the length with:
//     len(mockedDataStore.GetFilterDimensionOptionCalls())
func (mock *DataStoreMock) GetFilterDimensionOptionCalls() []struct {
	FilterID string
	Name     string
	Option   string
} {
	var calls []struct {
		FilterID string
		Name     string
		Option   string
	}
	lockDataStoreMockGetFilterDimensionOption.RLock()
	calls = mock.calls.GetFilterDimensionOption
	lockDataStoreMockGetFilterDimensionOption.RUnlock()
	return calls
}

// GetFilterDimensionOptions calls GetFilterDimensionOptionsFunc.
func (mock *DataStoreMock) GetFilterDimensionOptions(filterID string, name string) ([]models.DimensionOption, error) {
	if mock.GetFilterDimensionOptionsFunc == nil {
		panic("moq: DataStoreMock.GetFilterDimensionOptionsFunc is nil but DataStore.GetFilterDimensionOptions was just called")
	}
	callInfo := struct {
		FilterID string
		Name     string
	}{
		FilterID: filterID,
		Name:     name,
	}
	lockDataStoreMockGetFilterDimensionOptions.Lock()
	mock.calls.GetFilterDimensionOptions = append(mock.calls.GetFilterDimensionOptions, callInfo)
	lockDataStoreMockGetFilterDimensionOptions.Unlock()
	return mock.GetFilterDimensionOptionsFunc(filterID, name)
}

// GetFilterDimensionOptionsCalls gets all the calls that were made to GetFilterDimensionOptions.
// Check the length with:
//     len(mockedDataStore.GetFilterDimensionOptionsCalls())
func (mock *DataStoreMock) GetFilterDimensionOptionsCalls() []struct {
	FilterID string
	Name     string
} {
	var calls []struct {
		FilterID string
		Name     string
	}
	lockDataStoreMockGetFilterDimensionOptions.RLock()
	calls = mock.calls.GetFilterDimensionOptions
	lockDataStoreMockGetFilterDimensionOptions.RUnlock()
	return calls
}

// GetFilterDimensions calls GetFilterDimensionsFunc.
func (mock *DataStoreMock) GetFilterDimensions(filterID string) ([]models.Dimension, error) {
	if mock.GetFilterDimensionsFunc == nil {
		panic("moq: DataStoreMock.GetFilterDimensionsFunc is nil but DataStore.GetFilterDimensions was just called")
	}
	callInfo := struct {
		FilterID string
	}{
		FilterID: filterID,
	}
	lockDataStoreMockGetFilterDimensions.Lock()
	mock.calls.GetFilterDimensions = append(mock.calls.GetFilterDimensions, callInfo)
	lockDataStoreMockGetFilterDimensions.Unlock()
	return mock.GetFilterDimensionsFunc(filterID)
}

// GetFilterDimensionsCalls gets all the calls that were made to GetFilterDimensions.
// Check the length with:
//     len(mockedDataStore.GetFilterDimensionsCalls())
func (mock *DataStoreMock) GetFilterDimensionsCalls() []struct {
	FilterID string
} {
	var calls []struct {
		FilterID string
	}
	lockDataStoreMockGetFilterDimensions.RLock()
	calls = mock.calls.GetFilterDimensions
	lockDataStoreMockGetFilterDimensions.RUnlock()
	return calls
}

// GetFilterOutput calls GetFilterOutputFunc.
func (mock *DataStoreMock) GetFilterOutput(filterOutputID string) (*models.Filter, error) {
	if mock.GetFilterOutputFunc == nil {
		panic("moq: DataStoreMock.GetFilterOutputFunc is nil but DataStore.GetFilterOutput was just called")
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
func (mock *DataStoreMock) RemoveFilterDimension(filterID string, name string) error {
	if mock.RemoveFilterDimensionFunc == nil {
		panic("moq: DataStoreMock.RemoveFilterDimensionFunc is nil but DataStore.RemoveFilterDimension was just called")
	}
	callInfo := struct {
		FilterID string
		Name     string
	}{
		FilterID: filterID,
		Name:     name,
	}
	lockDataStoreMockRemoveFilterDimension.Lock()
	mock.calls.RemoveFilterDimension = append(mock.calls.RemoveFilterDimension, callInfo)
	lockDataStoreMockRemoveFilterDimension.Unlock()
	return mock.RemoveFilterDimensionFunc(filterID, name)
}

// RemoveFilterDimensionCalls gets all the calls that were made to RemoveFilterDimension.
// Check the length with:
//     len(mockedDataStore.RemoveFilterDimensionCalls())
func (mock *DataStoreMock) RemoveFilterDimensionCalls() []struct {
	FilterID string
	Name     string
} {
	var calls []struct {
		FilterID string
		Name     string
	}
	lockDataStoreMockRemoveFilterDimension.RLock()
	calls = mock.calls.RemoveFilterDimension
	lockDataStoreMockRemoveFilterDimension.RUnlock()
	return calls
}

// RemoveFilterDimensionOption calls RemoveFilterDimensionOptionFunc.
func (mock *DataStoreMock) RemoveFilterDimensionOption(filterID string, name string, option string) error {
	if mock.RemoveFilterDimensionOptionFunc == nil {
		panic("moq: DataStoreMock.RemoveFilterDimensionOptionFunc is nil but DataStore.RemoveFilterDimensionOption was just called")
	}
	callInfo := struct {
		FilterID string
		Name     string
		Option   string
	}{
		FilterID: filterID,
		Name:     name,
		Option:   option,
	}
	lockDataStoreMockRemoveFilterDimensionOption.Lock()
	mock.calls.RemoveFilterDimensionOption = append(mock.calls.RemoveFilterDimensionOption, callInfo)
	lockDataStoreMockRemoveFilterDimensionOption.Unlock()
	return mock.RemoveFilterDimensionOptionFunc(filterID, name, option)
}

// RemoveFilterDimensionOptionCalls gets all the calls that were made to RemoveFilterDimensionOption.
// Check the length with:
//     len(mockedDataStore.RemoveFilterDimensionOptionCalls())
func (mock *DataStoreMock) RemoveFilterDimensionOptionCalls() []struct {
	FilterID string
	Name     string
	Option   string
} {
	var calls []struct {
		FilterID string
		Name     string
		Option   string
	}
	lockDataStoreMockRemoveFilterDimensionOption.RLock()
	calls = mock.calls.RemoveFilterDimensionOption
	lockDataStoreMockRemoveFilterDimensionOption.RUnlock()
	return calls
}

// UpdateFilter calls UpdateFilterFunc.
func (mock *DataStoreMock) UpdateFilter(filter *models.Filter) error {
	if mock.UpdateFilterFunc == nil {
		panic("moq: DataStoreMock.UpdateFilterFunc is nil but DataStore.UpdateFilter was just called")
	}
	callInfo := struct {
		Filter *models.Filter
	}{
		Filter: filter,
	}
	lockDataStoreMockUpdateFilter.Lock()
	mock.calls.UpdateFilter = append(mock.calls.UpdateFilter, callInfo)
	lockDataStoreMockUpdateFilter.Unlock()
	return mock.UpdateFilterFunc(filter)
}

// UpdateFilterCalls gets all the calls that were made to UpdateFilter.
// Check the length with:
//     len(mockedDataStore.UpdateFilterCalls())
func (mock *DataStoreMock) UpdateFilterCalls() []struct {
	Filter *models.Filter
} {
	var calls []struct {
		Filter *models.Filter
	}
	lockDataStoreMockUpdateFilter.RLock()
	calls = mock.calls.UpdateFilter
	lockDataStoreMockUpdateFilter.RUnlock()
	return calls
}

// UpdateFilterOutput calls UpdateFilterOutputFunc.
func (mock *DataStoreMock) UpdateFilterOutput(filterOutput *models.Filter) error {
	if mock.UpdateFilterOutputFunc == nil {
		panic("moq: DataStoreMock.UpdateFilterOutputFunc is nil but DataStore.UpdateFilterOutput was just called")
	}
	callInfo := struct {
		FilterOutput *models.Filter
	}{
		FilterOutput: filterOutput,
	}
	lockDataStoreMockUpdateFilterOutput.Lock()
	mock.calls.UpdateFilterOutput = append(mock.calls.UpdateFilterOutput, callInfo)
	lockDataStoreMockUpdateFilterOutput.Unlock()
	return mock.UpdateFilterOutputFunc(filterOutput)
}

// UpdateFilterOutputCalls gets all the calls that were made to UpdateFilterOutput.
// Check the length with:
//     len(mockedDataStore.UpdateFilterOutputCalls())
func (mock *DataStoreMock) UpdateFilterOutputCalls() []struct {
	FilterOutput *models.Filter
} {
	var calls []struct {
		FilterOutput *models.Filter
	}
	lockDataStoreMockUpdateFilterOutput.RLock()
	calls = mock.calls.UpdateFilterOutput
	lockDataStoreMockUpdateFilterOutput.RUnlock()
	return calls
}
