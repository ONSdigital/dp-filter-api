// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package mocks

import (
	"context"
	"github.com/ONSdigital/dp-api-clients-go/dataset"
	"sync"
)

var (
	lockDatasetAPIMockGetOptionsBatchProcess sync.RWMutex
	lockDatasetAPIMockGetVersion             sync.RWMutex
	lockDatasetAPIMockGetVersionDimensions   sync.RWMutex
)

// DatasetAPIMock is a mock implementation of api.DatasetAPI.
//
//     func TestSomethingThatUsesDatasetAPI(t *testing.T) {
//
//         // make and configure a mocked api.DatasetAPI
//         mockedDatasetAPI := &DatasetAPIMock{
//             GetOptionsBatchProcessFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, id string, edition string, version string, dimension string, optionIDs *[]string, processBatch dataset.OptionsBatchProcessor, batchSize int, maxWorkers int) error {
// 	               panic("mock out the GetOptionsBatchProcess method")
//             },
//             GetVersionFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, downloadServiceAuthToken string, collectionID string, datasetID string, edition string, version string) (dataset.Version, error) {
// 	               panic("mock out the GetVersion method")
//             },
//             GetVersionDimensionsFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, id string, edition string, version string) (dataset.VersionDimensions, error) {
// 	               panic("mock out the GetVersionDimensions method")
//             },
//         }
//
//         // use mockedDatasetAPI in code that requires api.DatasetAPI
//         // and then make assertions.
//
//     }
type DatasetAPIMock struct {
	// GetOptionsBatchProcessFunc mocks the GetOptionsBatchProcess method.
	GetOptionsBatchProcessFunc func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, id string, edition string, version string, dimension string, optionIDs *[]string, processBatch dataset.OptionsBatchProcessor, batchSize int, maxWorkers int) error

	// GetVersionFunc mocks the GetVersion method.
	GetVersionFunc func(ctx context.Context, userAuthToken string, serviceAuthToken string, downloadServiceAuthToken string, collectionID string, datasetID string, edition string, version string) (dataset.Version, error)

	// GetVersionDimensionsFunc mocks the GetVersionDimensions method.
	GetVersionDimensionsFunc func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, id string, edition string, version string) (dataset.VersionDimensions, error)

	// calls tracks calls to the methods.
	calls struct {
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
			ProcessBatch dataset.OptionsBatchProcessor
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
}

// GetOptionsBatchProcess calls GetOptionsBatchProcessFunc.
func (mock *DatasetAPIMock) GetOptionsBatchProcess(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, id string, edition string, version string, dimension string, optionIDs *[]string, processBatch dataset.OptionsBatchProcessor, batchSize int, maxWorkers int) error {
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
		ProcessBatch     dataset.OptionsBatchProcessor
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
	lockDatasetAPIMockGetOptionsBatchProcess.Lock()
	mock.calls.GetOptionsBatchProcess = append(mock.calls.GetOptionsBatchProcess, callInfo)
	lockDatasetAPIMockGetOptionsBatchProcess.Unlock()
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
	ProcessBatch     dataset.OptionsBatchProcessor
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
		ProcessBatch     dataset.OptionsBatchProcessor
		BatchSize        int
		MaxWorkers       int
	}
	lockDatasetAPIMockGetOptionsBatchProcess.RLock()
	calls = mock.calls.GetOptionsBatchProcess
	lockDatasetAPIMockGetOptionsBatchProcess.RUnlock()
	return calls
}

// GetVersion calls GetVersionFunc.
func (mock *DatasetAPIMock) GetVersion(ctx context.Context, userAuthToken string, serviceAuthToken string, downloadServiceAuthToken string, collectionID string, datasetID string, edition string, version string) (dataset.Version, error) {
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
	lockDatasetAPIMockGetVersion.Lock()
	mock.calls.GetVersion = append(mock.calls.GetVersion, callInfo)
	lockDatasetAPIMockGetVersion.Unlock()
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
	lockDatasetAPIMockGetVersion.RLock()
	calls = mock.calls.GetVersion
	lockDatasetAPIMockGetVersion.RUnlock()
	return calls
}

// GetVersionDimensions calls GetVersionDimensionsFunc.
func (mock *DatasetAPIMock) GetVersionDimensions(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, id string, edition string, version string) (dataset.VersionDimensions, error) {
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
	lockDatasetAPIMockGetVersionDimensions.Lock()
	mock.calls.GetVersionDimensions = append(mock.calls.GetVersionDimensions, callInfo)
	lockDatasetAPIMockGetVersionDimensions.Unlock()
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
	lockDatasetAPIMockGetVersionDimensions.RLock()
	calls = mock.calls.GetVersionDimensions
	lockDatasetAPIMockGetVersionDimensions.RUnlock()
	return calls
}
