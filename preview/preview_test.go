package preview

import (
	"errors"
	"github.com/ONSdigital/dp-filter-api/models"
	"github.com/ONSdigital/dp-filter-api/preview/previewtest"
	"github.com/ONSdigital/dp-filter/observation"
	"github.com/ONSdigital/dp-filter/observation/observationtest"
	. "github.com/smartystreets/goconvey/convey"
	"io"
	"testing"
)

func TestPreviewDatasetStore_GetPreview(t *testing.T) {
	t.Parallel()
	Convey("Successfully returns 3 results", t, func() {
		rowCount := 0
		mockRowReader := &observationtest.CSVRowReaderMock{
			ReadFunc: func() (string, error) {
				if rowCount == 4 {
					return "", io.EOF
				}
				rowCount++
				return "V4_0,Time_codelist,Time,Geography_codelist,Geography,Aggregate_codelist,Aggregate", nil
			},
			CloseFunc: func() error {
				return nil
			},
		}

		mockedObservationStore := &observationstoretest.ObservationStoreMock{
			GetCSVRowsFunc: func(in1 *observation.Filter, in2 *int) (observation.CSVRowReader, error) {
				return mockRowReader, nil
			},
		}
		previewDataset := PreviewDatasetStore{Store: mockedObservationStore, Limit: 3}

		results, err := previewDataset.GetPreview(models.Filter{})
		So(len(results.Headers), ShouldEqual, 7)
		So(err, ShouldBeNil)
		So(len(results.Rows), ShouldEqual, 3)

	})

	Convey("Successfully returns 0 results due to data sparsity", t, func() {
		rowCount := 0
		mockRowReader := &observationtest.CSVRowReaderMock{
			ReadFunc: func() (string, error) {
				if rowCount == 1 {
					return "", io.EOF
				}
				rowCount++
				return "V4_0,Time_codelist,Time,Geography_codelist,Geography,Aggregate_codelist,Aggregate", nil
			},
			CloseFunc: func() error {
				return nil
			},
		}

		mockedObservationStore := &observationstoretest.ObservationStoreMock{
			GetCSVRowsFunc: func(in1 *observation.Filter, in2 *int) (observation.CSVRowReader, error) {
				return mockRowReader, nil
			},
		}
		limit := 3
		previewDataset := PreviewDatasetStore{Store: mockedObservationStore, Limit: limit}

		results, err := previewDataset.GetPreview(models.Filter{})
		So(len(results.Headers), ShouldEqual, 7)
		So(err, ShouldBeNil)
		So(len(results.Rows), ShouldEqual, 0)
	})
}

func TestPreviewDatasetStore_GetPreview_ErrorStates(t *testing.T) {
	Convey("When a query error happens, return the error", t, func() {
		expectedError := errors.New("query error")
		mockedObservationStore := &observationstoretest.ObservationStoreMock{
			GetCSVRowsFunc: func(in1 *observation.Filter, in2 *int) (observation.CSVRowReader, error) {
				return nil, expectedError
			},
		}
		previewDataset := PreviewDatasetStore{Store: mockedObservationStore, Limit: 3}
		_, err := previewDataset.GetPreview(models.Filter{})
		So(err, ShouldEqual, expectedError)
	})

	Convey("When a reader stream breaks", t, func() {
		expectedError := errors.New("broken stream")
		mockRowReader := &observationtest.CSVRowReaderMock{
			ReadFunc: func() (string, error) {

				return "", expectedError
			},
			CloseFunc: func() error {
				return nil
			},
		}

		mockedObservationStore := &observationstoretest.ObservationStoreMock{
			GetCSVRowsFunc: func(in1 *observation.Filter, in2 *int) (observation.CSVRowReader, error) {
				return mockRowReader, nil
			},
		}
		previewDataset := PreviewDatasetStore{Store: mockedObservationStore, Limit: 3}
		_, err := previewDataset.GetPreview(models.Filter{})
		So(err, ShouldEqual, expectedError)
	})

}
