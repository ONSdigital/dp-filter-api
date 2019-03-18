package preview

import (
	"context"
	"encoding/csv"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/ONSdigital/dp-filter-api/models"
	"github.com/ONSdigital/dp-filter-api/preview/previewtest"
	"github.com/ONSdigital/dp-graph/observation"
	"github.com/ONSdigital/dp-graph/observation/observationtest"
	. "github.com/smartystreets/goconvey/convey"
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
				return "V4_0,Time_codelist,Time,Geography_codelist,Geography,Aggregate_codelist,Aggregate\n", nil
			},
			CloseFunc: func(context.Context) error {
				return nil
			},
		}

		mockedObservationStore := &observationstoretest.ObservationStoreMock{
			StreamCSVRowsFunc: func(ctx context.Context, in1 *observation.Filter, in2 *int) (observation.StreamRowReader, error) {
				return mockRowReader, nil
			},
		}
		previewDataset := DatasetStore{Store: mockedObservationStore}

		results, err := previewDataset.GetPreview(context.Background(), &models.Filter{}, 0)
		So(len(results.Headers), ShouldEqual, 7)
		So(results.NumberOfColumns, ShouldEqual, 7)
		So(results.NumberOfRows, ShouldEqual, 3)
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
			CloseFunc: func(context.Context) error {
				return nil
			},
		}

		mockedObservationStore := &observationstoretest.ObservationStoreMock{
			StreamCSVRowsFunc: func(ctx context.Context, in1 *observation.Filter, in2 *int) (observation.StreamRowReader, error) {
				return mockRowReader, nil
			},
		}
		previewDataset := DatasetStore{Store: mockedObservationStore}

		results, err := previewDataset.GetPreview(context.Background(), &models.Filter{}, 0)
		So(len(results.Headers), ShouldEqual, 7)
		So(results.NumberOfColumns, ShouldEqual, 7)
		So(results.NumberOfRows, ShouldEqual, 0)
		So(err, ShouldBeNil)
		So(len(results.Rows), ShouldEqual, 0)
	})
}

func TestPreviewDatasetStore_GetPreview_ErrorStates(t *testing.T) {
	Convey("When a query error happens, return the error", t, func() {
		expectedError := errors.New("query error")
		mockedObservationStore := &observationstoretest.ObservationStoreMock{
			StreamCSVRowsFunc: func(ctx context.Context, in1 *observation.Filter, in2 *int) (observation.StreamRowReader, error) {
				return nil, expectedError
			},
		}
		previewDataset := DatasetStore{Store: mockedObservationStore}
		_, err := previewDataset.GetPreview(context.Background(), &models.Filter{}, 0)
		So(err, ShouldEqual, expectedError)
	})

	Convey("When a reader stream breaks", t, func() {
		expectedError := errors.New("broken stream")
		mockRowReader := &observationtest.CSVRowReaderMock{
			ReadFunc: func() (string, error) {

				return "", expectedError
			},
			CloseFunc: func(context.Context) error {
				return nil
			},
		}

		mockedObservationStore := &observationstoretest.ObservationStoreMock{
			StreamCSVRowsFunc: func(ctx context.Context, in1 *observation.Filter, in2 *int) (observation.StreamRowReader, error) {
				return mockRowReader, nil
			},
		}
		previewDataset := DatasetStore{Store: mockedObservationStore}
		_, err := previewDataset.GetPreview(context.Background(), &models.Filter{}, 0)
		So(err, ShouldEqual, expectedError)
	})

}

func TestPreviewDatasetStore_buildPreview(t *testing.T) {
	Convey("When a building the preview results with as CSV cell containing a quoted commas", t, func() {
		csvReader := csv.NewReader(strings.NewReader("\",\",2,3\n1,2,3"))

		results, err := buildResults(csvReader)
		So(err, ShouldBeNil)
		So(len(results.Headers), ShouldEqual, 3)
		So(len(results.Rows[0]), ShouldEqual, 3)

	})
}
