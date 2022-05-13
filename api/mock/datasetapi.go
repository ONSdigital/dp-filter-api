// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package mock

import (
	"context"
	datasetAPI "github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-filter-api/api"
	"sync"
)

// Ensure, that DatasetAPIMock does implement api.DatasetAPI.
// If this is not the case, regenerate this file with moq.
var _ api.DatasetAPI = &DatasetAPIMock{}

// DatasetAPIMock is a mock implementation of api.DatasetAPI.
//
// 	func TestSomethingThatUsesDatasetAPI(t *testing.T) {
//
// 		// make and configure a mocked api.DatasetAPI
// 		mockedDatasetAPI := &DatasetAPIMock{
// 			GetFunc: func(ctx context.Context, userToken string, svcToken string, collectionID string, datasetID string) (datasetAPI.DatasetDetails, error) {
// 				panic("mock out the Get method")
// 			},
// 			GetOptionsBatchProcessFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, id string, edition string, version string, dimension string, optionIDs *[]string, processBatch datasetAPI.OptionsBatchProcessor, batchSize int, maxWorkers int) error {
// 				panic("mock out the GetOptionsBatchProcess method")
// 			},
// 			GetVersionFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, downloadServiceAuthToken string, collectionID string, datasetID string, edition string, version string) (datasetAPI.Version, error) {
// 				panic("mock out the GetVersion method")
// 			},
// 			GetVersionDimensionsFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, id string, edition string, version string) (datasetAPI.VersionDimensions, error) {
// 				panic("mock out the GetVersionDimensions method")
// 			},
// 		}
//
// 		// use mockedDatasetAPI in code that requires api.DatasetAPI
// 		// and then make assertions.
//
// 	}
type DatasetAPIMock struct {
	// GetFunc mocks the Get method.
	GetFunc func(ctx context.Context, userToken string, svcToken string, collectionID string, datasetID string) (datasetAPI.DatasetDetails, error)

	// GetOptionsBatchProcessFunc mocks the GetOptionsBatchProcess method.
	GetOptionsBatchProcessFunc func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, id string, edition string, version string, dimension string, optionIDs *[]string, processBatch datasetAPI.OptionsBatchProcessor, batchSize int, maxWorkers int) error

	// GetVersionFunc mocks the GetVersion method.
	GetVersionFunc func(ctx context.Context, userAuthToken string, serviceAuthToken string, downloadServiceAuthToken string, collectionID string, datasetID string, edition string, version string) (datasetAPI.Version, error)

	// GetVersionDimensionsFunc mocks the GetVersionDimensions method.
	GetVersionDimensionsFunc func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, id string, edition string, version string) (datasetAPI.VersionDimensions, error)

	// calls tracks calls to the methods.
	calls struct {
		// Get holds details about calls to the Get method.
		Get []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// UserToken is the userToken argument value.
			UserToken string
			// SvcToken is the svcToken argument value.
			SvcToken string
			// CollectionID is the collectionID argument value.
			CollectionID string
			// DatasetID is the datasetID argument value.
			DatasetID string
		}
		// GetOptionsBatchProcess holds details about calls to the GetOptionsBatchProcess method.
		GetOptionsBatchProcess []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// UserAuthToken is the userAuthToken argument value.
			UserAuthToken string
			// ServiceAuthToken is the serviceAuthToken argument value.
			ServiceAuthToken string
			// CollectionID is the collectionID argument value.
			CollectionID string
			// ID is the id argument value.
			ID string
			// Edition is the edition argument value.
			Edition string
			// Version is the version argument value.
			Version string
			// Dimension is the dimension argument value.
			Dimension string
			// OptionIDs is the optionIDs argument value.
			OptionIDs *[]string
			// ProcessBatch is the processBatch argument value.
			ProcessBatch datasetAPI.OptionsBatchProcessor
			// BatchSize is the batchSize argument value.
			BatchSize int
			// MaxWorkers is the maxWorkers argument value.
			MaxWorkers int
		}
		// GetVersion holds details about calls to the GetVersion method.
		GetVersion []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// UserAuthToken is the userAuthToken argument value.
			UserAuthToken string
			// ServiceAuthToken is the serviceAuthToken argument value.
			ServiceAuthToken string
			// DownloadServiceAuthToken is the downloadServiceAuthToken argument value.
			DownloadServiceAuthToken string
			// CollectionID is the collectionID argument value.
			CollectionID string
			// DatasetID is the datasetID argument value.
			DatasetID string
			// Edition is the edition argument value.
			Edition string
			// Version is the version argument value.
			Version string
		}
		// GetVersionDimensions holds details about calls to the GetVersionDimensions method.
		GetVersionDimensions []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// UserAuthToken is the userAuthToken argument value.
			UserAuthToken string
			// ServiceAuthToken is the serviceAuthToken argument value.
			ServiceAuthToken string
			// CollectionID is the collectionID argument value.
			CollectionID string
			// ID is the id argument value.
			ID string
			// Edition is the edition argument value.
			Edition string
			// Version is the version argument value.
			Version string
		}
	}
	lockGet                    sync.RWMutex
	lockGetOptionsBatchProcess sync.RWMutex
	lockGetVersion             sync.RWMutex
	lockGetVersionDimensions   sync.RWMutex
}

// Get calls GetFunc.
func (mock *DatasetAPIMock) Get(ctx context.Context, userToken string, svcToken string, collectionID string, datasetID string) (datasetAPI.DatasetDetails, error) {
	if mock.GetFunc == nil {
		panic("DatasetAPIMock.GetFunc: method is nil but DatasetAPI.Get was just called")
	}
	callInfo := struct {
		Ctx          context.Context
		UserToken    string
		SvcToken     string
		CollectionID string
		DatasetID    string
	}{
		Ctx:          ctx,
		UserToken:    userToken,
		SvcToken:     svcToken,
		CollectionID: collectionID,
		DatasetID:    datasetID,
	}
	mock.lockGet.Lock()
	mock.calls.Get = append(mock.calls.Get, callInfo)
	mock.lockGet.Unlock()
	return mock.GetFunc(ctx, userToken, svcToken, collectionID, datasetID)
}

// GetCalls gets all the calls that were made to Get.
// Check the length with:
//     len(mockedDatasetAPI.GetCalls())
func (mock *DatasetAPIMock) GetCalls() []struct {
	Ctx          context.Context
	UserToken    string
	SvcToken     string
	CollectionID string
	DatasetID    string
} {
	var calls []struct {
		Ctx          context.Context
		UserToken    string
		SvcToken     string
		CollectionID string
		DatasetID    string
	}
	mock.lockGet.RLock()
	calls = mock.calls.Get
	mock.lockGet.RUnlock()
	return calls
}

// GetOptionsBatchProcess calls GetOptionsBatchProcessFunc.
func (mock *DatasetAPIMock) GetOptionsBatchProcess(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, id string, edition string, version string, dimension string, optionIDs *[]string, processBatch datasetAPI.OptionsBatchProcessor, batchSize int, maxWorkers int) error {
	if mock.GetOptionsBatchProcessFunc == nil {
		panic("DatasetAPIMock.GetOptionsBatchProcessFunc: method is nil but DatasetAPI.GetOptionsBatchProcess was just called")
	}
	callInfo := struct {
		Ctx              context.Context
		UserAuthToken    string
		ServiceAuthToken string
		CollectionID     string
		ID               string
		Edition          string
		Version          string
		Dimension        string
		OptionIDs        *[]string
		ProcessBatch     datasetAPI.OptionsBatchProcessor
		BatchSize        int
		MaxWorkers       int
	}{
		Ctx:              ctx,
		UserAuthToken:    userAuthToken,
		ServiceAuthToken: serviceAuthToken,
		CollectionID:     collectionID,
		ID:               id,
		Edition:          edition,
		Version:          version,
		Dimension:        dimension,
		OptionIDs:        optionIDs,
		ProcessBatch:     processBatch,
		BatchSize:        batchSize,
		MaxWorkers:       maxWorkers,
	}
	mock.lockGetOptionsBatchProcess.Lock()
	mock.calls.GetOptionsBatchProcess = append(mock.calls.GetOptionsBatchProcess, callInfo)
	mock.lockGetOptionsBatchProcess.Unlock()
	return mock.GetOptionsBatchProcessFunc(ctx, userAuthToken, serviceAuthToken, collectionID, id, edition, version, dimension, optionIDs, processBatch, batchSize, maxWorkers)
}

// GetOptionsBatchProcessCalls gets all the calls that were made to GetOptionsBatchProcess.
// Check the length with:
//     len(mockedDatasetAPI.GetOptionsBatchProcessCalls())
func (mock *DatasetAPIMock) GetOptionsBatchProcessCalls() []struct {
	Ctx              context.Context
	UserAuthToken    string
	ServiceAuthToken string
	CollectionID     string
	ID               string
	Edition          string
	Version          string
	Dimension        string
	OptionIDs        *[]string
	ProcessBatch     datasetAPI.OptionsBatchProcessor
	BatchSize        int
	MaxWorkers       int
} {
	var calls []struct {
		Ctx              context.Context
		UserAuthToken    string
		ServiceAuthToken string
		CollectionID     string
		ID               string
		Edition          string
		Version          string
		Dimension        string
		OptionIDs        *[]string
		ProcessBatch     datasetAPI.OptionsBatchProcessor
		BatchSize        int
		MaxWorkers       int
	}
	mock.lockGetOptionsBatchProcess.RLock()
	calls = mock.calls.GetOptionsBatchProcess
	mock.lockGetOptionsBatchProcess.RUnlock()
	return calls
}

// GetVersion calls GetVersionFunc.
func (mock *DatasetAPIMock) GetVersion(ctx context.Context, userAuthToken string, serviceAuthToken string, downloadServiceAuthToken string, collectionID string, datasetID string, edition string, version string) (datasetAPI.Version, error) {
	if mock.GetVersionFunc == nil {
		panic("DatasetAPIMock.GetVersionFunc: method is nil but DatasetAPI.GetVersion was just called")
	}
	callInfo := struct {
		Ctx                      context.Context
		UserAuthToken            string
		ServiceAuthToken         string
		DownloadServiceAuthToken string
		CollectionID             string
		DatasetID                string
		Edition                  string
		Version                  string
	}{
		Ctx:                      ctx,
		UserAuthToken:            userAuthToken,
		ServiceAuthToken:         serviceAuthToken,
		DownloadServiceAuthToken: downloadServiceAuthToken,
		CollectionID:             collectionID,
		DatasetID:                datasetID,
		Edition:                  edition,
		Version:                  version,
	}
	mock.lockGetVersion.Lock()
	mock.calls.GetVersion = append(mock.calls.GetVersion, callInfo)
	mock.lockGetVersion.Unlock()
	return mock.GetVersionFunc(ctx, userAuthToken, serviceAuthToken, downloadServiceAuthToken, collectionID, datasetID, edition, version)
}

// GetVersionCalls gets all the calls that were made to GetVersion.
// Check the length with:
//     len(mockedDatasetAPI.GetVersionCalls())
func (mock *DatasetAPIMock) GetVersionCalls() []struct {
	Ctx                      context.Context
	UserAuthToken            string
	ServiceAuthToken         string
	DownloadServiceAuthToken string
	CollectionID             string
	DatasetID                string
	Edition                  string
	Version                  string
} {
	var calls []struct {
		Ctx                      context.Context
		UserAuthToken            string
		ServiceAuthToken         string
		DownloadServiceAuthToken string
		CollectionID             string
		DatasetID                string
		Edition                  string
		Version                  string
	}
	mock.lockGetVersion.RLock()
	calls = mock.calls.GetVersion
	mock.lockGetVersion.RUnlock()
	return calls
}

// GetVersionDimensions calls GetVersionDimensionsFunc.
func (mock *DatasetAPIMock) GetVersionDimensions(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, id string, edition string, version string) (datasetAPI.VersionDimensions, error) {
	if mock.GetVersionDimensionsFunc == nil {
		panic("DatasetAPIMock.GetVersionDimensionsFunc: method is nil but DatasetAPI.GetVersionDimensions was just called")
	}
	callInfo := struct {
		Ctx              context.Context
		UserAuthToken    string
		ServiceAuthToken string
		CollectionID     string
		ID               string
		Edition          string
		Version          string
	}{
		Ctx:              ctx,
		UserAuthToken:    userAuthToken,
		ServiceAuthToken: serviceAuthToken,
		CollectionID:     collectionID,
		ID:               id,
		Edition:          edition,
		Version:          version,
	}
	mock.lockGetVersionDimensions.Lock()
	mock.calls.GetVersionDimensions = append(mock.calls.GetVersionDimensions, callInfo)
	mock.lockGetVersionDimensions.Unlock()
	return mock.GetVersionDimensionsFunc(ctx, userAuthToken, serviceAuthToken, collectionID, id, edition, version)
}

// GetVersionDimensionsCalls gets all the calls that were made to GetVersionDimensions.
// Check the length with:
//     len(mockedDatasetAPI.GetVersionDimensionsCalls())
func (mock *DatasetAPIMock) GetVersionDimensionsCalls() []struct {
	Ctx              context.Context
	UserAuthToken    string
	ServiceAuthToken string
	CollectionID     string
	ID               string
	Edition          string
	Version          string
} {
	var calls []struct {
		Ctx              context.Context
		UserAuthToken    string
		ServiceAuthToken string
		CollectionID     string
		ID               string
		Edition          string
		Version          string
	}
	mock.lockGetVersionDimensions.RLock()
	calls = mock.calls.GetVersionDimensions
	mock.lockGetVersionDimensions.RUnlock()
	return calls
}
