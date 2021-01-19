// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package mocks

import (
	"context"
	"github.com/ONSdigital/dp-filter-api/models"
	"sync"
)

var (
	lockPreviewDatasetMockGetPreview sync.RWMutex
)

// PreviewDatasetMock is a mock implementation of api.PreviewDataset.
//
//     func TestSomethingThatUsesPreviewDataset(t *testing.T) {
//
//         // make and configure a mocked api.PreviewDataset
//         mockedPreviewDataset := &PreviewDatasetMock{
//             GetPreviewFunc: func(ctx context.Context, filter *models.Filter, limit int) (*models.FilterPreview, error) {
// 	               panic("mock out the GetPreview method")
//             },
//         }
//
//         // use mockedPreviewDataset in code that requires api.PreviewDataset
//         // and then make assertions.
//
//     }
type PreviewDatasetMock struct {
	// GetPreviewFunc mocks the GetPreview method.
	GetPreviewFunc func(ctx context.Context, filter *models.Filter, limit int) (*models.FilterPreview, error)

	// calls tracks calls to the methods.
	calls struct {
		// GetPreview holds details about calls to the GetPreview method.
		GetPreview []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Filter is the filter argument value.
			Filter *models.Filter
			// Limit is the limit argument value.
			Limit int
		}
	}
}

// GetPreview calls GetPreviewFunc.
func (mock *PreviewDatasetMock) GetPreview(ctx context.Context, filter *models.Filter, limit int) (*models.FilterPreview, error) {
	if mock.GetPreviewFunc == nil {
		panic("PreviewDatasetMock.GetPreviewFunc: method is nil but PreviewDataset.GetPreview was just called")
	}
	callInfo := struct {
		Ctx    context.Context
		Filter *models.Filter
		Limit  int
	}{
		Ctx:    ctx,
		Filter: filter,
		Limit:  limit,
	}
	lockPreviewDatasetMockGetPreview.Lock()
	mock.calls.GetPreview = append(mock.calls.GetPreview, callInfo)
	lockPreviewDatasetMockGetPreview.Unlock()
	return mock.GetPreviewFunc(ctx, filter, limit)
}

// GetPreviewCalls gets all the calls that were made to GetPreview.
// Check the length with:
//     len(mockedPreviewDataset.GetPreviewCalls())
func (mock *PreviewDatasetMock) GetPreviewCalls() []struct {
	Ctx    context.Context
	Filter *models.Filter
	Limit  int
} {
	var calls []struct {
		Ctx    context.Context
		Filter *models.Filter
		Limit  int
	}
	lockPreviewDatasetMockGetPreview.RLock()
	calls = mock.calls.GetPreview
	lockPreviewDatasetMockGetPreview.RUnlock()
	return calls
}
