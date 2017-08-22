package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/ONSdigital/dp-filter-api/models"
	sqlmock "github.com/go-sqlmock"
	"github.com/lib/pq"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	addFilterSQL             = "INSERT INTO Filters(.+) VALUES(.+)"
	addDimensionSQL          = "INSERT INTO Dimensions"
	addDimensionOptionSQL    = "INSERT INTO Dimensions(.+) VALUES"
	deleteDimensionSQL       = "DELETE FROM Dimensions WHERE filterJobId = (.+) AND name = (.+)"
	deleteDimensionOptionSQL = "DELETE FROM Dimensions WHERE filterJobId = (.+) AND name = (.+) AND option = (.+)"
	getFilterSQL             = "SELECT (.*) FROM Filters WHERE"
	getFilterStateSQL        = "SELECT state FROM Filters WHERE"
	getDimensionSQL          = "SELECT name FROM Dimensions WHERE filterJobId = (.+) AND name = (.+)"
	getDimensionsSQL         = "SELECT name FROM Dimensions WHERE filterJobId = (.+)"
	getDimensionOptionsSQL   = "SELECT option FROM Dimensions WHERE filterJobId = (.+) AND name = (.+)"
	getDimensionOptionSQL    = "SELECT option FROM Dimensions WHERE filterJobId = (.+) AND name = (.+) AND option = (.+)"
	getDownloadItemsSQL      = "SELECT size, type, url FROM Downloads WHERE"
	updateFilterStateSQL     = "UPDATE Filters SET"
	upsertDimensionOptionSQL = "INSERT INTO Dimensions(.+) VALUES(.+) ON CONFLICT ON CONSTRAINT (.+)"
	upsertDownloadURLSQL     = "INSERT INTO Downloads(.+) VALUES(.+) ON CONFLICT ON CONSTRAINT (.+)"
)

func TestNewPostgresDatastore(t *testing.T) {
	t.Parallel()
	Convey("When creating a postgres datastore no errors are returned", t, func() {
		mock, db := NewSQLMockWithSQLStatements()
		_, err := NewDatastore(db)
		So(err, ShouldBeNil)
		So(mock.ExpectationsWereMet(), ShouldBeNil)
	})
}

func TestAddFilter(t *testing.T) {
	t.Parallel()
	Convey("Successfully create a new filter job", t, func() {
		mock, db := NewSQLMockWithSQLStatements()
		ds, err := NewDatastore(db)
		So(err, ShouldBeNil)
		mock.ExpectBegin()
		mock.ExpectPrepare(addFilterSQL).ExpectExec().WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectPrepare(addDimensionOptionSQL).ExpectExec().WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectPrepare(addDimensionOptionSQL).ExpectExec().WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		var dimensions []models.Dimension

		dimensions = append(dimensions, models.Dimension{
			Name:    "age",
			Options: []string{"10", "20"},
		})

		newFilter := &models.Filter{
			FilterID:   "123",
			Dimensions: dimensions,
			State:      "created",
		}

		filter, err := ds.AddFilter("80", newFilter)
		So(err, ShouldBeNil)
		So(mock.ExpectationsWereMet(), ShouldBeNil)
		So(filter, ShouldNotBeNil)
	})
}

func TestAddFilterDimensions(t *testing.T) {
	t.Parallel()
	Convey("Successfully add single dimension to filter", t, func() {
		mock, db := NewSQLMockWithSQLStatements()
		ds, err := NewDatastore(db)
		So(err, ShouldBeNil)
		mock.ExpectBegin()
		mock.ExpectPrepare(getFilterSQL).ExpectQuery().WithArgs(sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"filterId", "datasetFilterId", "state"}).
				AddRow("123", "12345678", "created"))
		mock.ExpectPrepare(deleteDimensionSQL).ExpectExec().WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 0))
		mock.ExpectPrepare(addDimensionSQL).ExpectExec().WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		dimensionObject := &models.AddDimension{
			FilterID: "123",
			Name:     "age",
			Options:  []string{},
		}

		err = ds.AddFilterDimension(dimensionObject)
		So(err, ShouldBeNil)
		So(mock.ExpectationsWereMet(), ShouldBeNil)
	})

	Convey("Successfully add populated list of options for a dimension for filter", t, func() {
		mock, db := NewSQLMockWithSQLStatements()
		ds, err := NewDatastore(db)
		So(err, ShouldBeNil)
		mock.ExpectBegin()
		mock.ExpectPrepare(getFilterSQL).ExpectQuery().WithArgs(sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"filterId", "datasetFilterId", "state"}).
				AddRow("123", "12345678", "created"))
		mock.ExpectPrepare(deleteDimensionSQL).ExpectExec().WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectPrepare(addDimensionOptionSQL).ExpectExec().WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectPrepare(addDimensionOptionSQL).ExpectExec().WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		dimensionObject := &models.AddDimension{
			FilterID: "123",
			Name:     "age",
			Options:  []string{"10", "20"},
		}

		err = ds.AddFilterDimension(dimensionObject)
		So(err, ShouldBeNil)
		So(mock.ExpectationsWereMet(), ShouldBeNil)
	})
}

func TestAddFilterDimensionOption(t *testing.T) {
	t.Parallel()
	Convey("Successfully add an option to a dimension on a filter job", t, func() {
		mock, db := NewSQLMockWithSQLStatements()
		ds, err := NewDatastore(db)
		So(err, ShouldBeNil)
		mock.ExpectQuery(getFilterSQL).WithArgs(sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"filterId", "datasetFilterId", "state"}).
				AddRow("123", "12345678", "created"))
		mock.ExpectQuery(getDimensionSQL).WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"name"}).
				AddRow("age"))
		mock.ExpectPrepare(upsertDimensionOptionSQL).ExpectExec().WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))

		dimensionOptionObject := &models.AddDimensionOption{
			FilterID: "123",
			Name:     "age",
			Option:   "12",
		}

		err = ds.AddFilterDimensionOption(dimensionOptionObject)
		So(err, ShouldBeNil)
		So(mock.ExpectationsWereMet(), ShouldBeNil)
	})
}

func TestGetFilter(t *testing.T) {
	t.Parallel()
	Convey("When a filterJobId is provided, the filter job is returned", t, func() {
		mock, db := NewSQLMockWithSQLStatements()
		ds, err := NewDatastore(db)
		So(err, ShouldBeNil)
		mock.ExpectBegin()
		mock.ExpectPrepare(getFilterSQL).ExpectQuery().WithArgs(sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"filterJobId", "datasetFilterId", "state"}).
				AddRow("123", "12345678", "completed"))
		mock.ExpectPrepare(getDownloadItemsSQL).ExpectQuery().WithArgs(sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"size", "type", "url"}).
				AddRow("24mb", "csv", "csv/s3url").
				AddRow("2mb", "json", "json/s3url").
				AddRow("30mb", "xls", "xls/s3url"))

		expectedFilter := models.Filter{
			FilterID:         "123",
			DatasetFilterID:  "12345678",
			State:            "completed",
			DimensionListURL: "/filters/123/dimensions",
			Downloads: models.Downloads{
				CSV: models.DownloadItem{
					Size: "24mb",
					URL:  "csv/s3url",
				},
				JSON: models.DownloadItem{
					Size: "2mb",
					URL:  "json/s3url",
				},
				XLS: models.DownloadItem{
					Size: "30mb",
					URL:  "xls/s3url",
				},
			},
		}

		filter, err := ds.GetFilter("123")
		So(err, ShouldBeNil)
		So(filter, ShouldResemble, expectedFilter)
		So(mock.ExpectationsWereMet(), ShouldBeNil)
	})
}

func TestGetFilterDimensions(t *testing.T) {
	t.Parallel()
	Convey("When a filterJobId is provided, all dimensions for the filter job are returned", t, func() {
		mock, db := NewSQLMockWithSQLStatements()
		ds, err := NewDatastore(db)
		So(err, ShouldBeNil)
		mock.ExpectQuery(getFilterSQL).WithArgs(sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"filterJobId", "datasetFilterId", "state"}).
				AddRow("123", "12345678", "completed"))
		mock.ExpectQuery(getDimensionsSQL).WithArgs(sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"name"}).
				AddRow("age").AddRow("time"))

		var expectedDimensions []models.Dimension

		dimension1 := models.Dimension{
			DimensionURL: "/filters/123/dimensions/age",
			Name:         "age",
		}

		dimension2 := models.Dimension{
			DimensionURL: "/filters/123/dimensions/time",
			Name:         "time",
		}

		expectedDimensions = append(expectedDimensions, dimension1, dimension2)

		dimensions, err := ds.GetFilterDimensions("123")
		So(err, ShouldBeNil)
		So(dimensions, ShouldResemble, expectedDimensions)
		So(mock.ExpectationsWereMet(), ShouldBeNil)
	})
}

func TestGetFilterDimensionOptions(t *testing.T) {
	t.Parallel()
	Convey("When a filterJobId and dimension name is provided, all dimensions options for the filter job are returned", t, func() {
		mock, db := NewSQLMockWithSQLStatements()
		ds, err := NewDatastore(db)
		So(err, ShouldBeNil)
		mock.ExpectQuery(getFilterSQL).WithArgs(sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"filterJobId", "datasetFilterId", "state"}).
				AddRow("123", "12345678", "created"))
		mock.ExpectQuery(getDimensionSQL).WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"name"}).
				AddRow("age"))
		mock.ExpectQuery(getDimensionOptionsSQL).WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"option"}).
				AddRow("29").AddRow("7"))

		var expectedDimensionOptions []models.DimensionOption

		option1 := models.DimensionOption{
			DimensionOptionURL: "/filters/123/dimensions/age/options/29",
			Option:             "29",
		}

		option2 := models.DimensionOption{
			DimensionOptionURL: "/filters/123/dimensions/age/options/7",
			Option:             "7",
		}

		expectedDimensionOptions = append(expectedDimensionOptions, option1, option2)

		dimensionOptions, err := ds.GetFilterDimensionOptions("123", "age")
		So(err, ShouldBeNil)
		So(dimensionOptions, ShouldResemble, expectedDimensionOptions)
		So(mock.ExpectationsWereMet(), ShouldBeNil)
	})
}

func TestGetFilterDimension(t *testing.T) {
	t.Parallel()
	Convey("When a filterJobId and dimension name is provided, successfully return without erroring", t, func() {
		mock, db := NewSQLMockWithSQLStatements()
		ds, err := NewDatastore(db)
		So(err, ShouldBeNil)
		mock.ExpectQuery(getFilterSQL).WithArgs(sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"filterJobId", "datasetFilterId", "state"}).
				AddRow("123", "12345678", "created"))
		mock.ExpectQuery(getDimensionSQL).WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"name"}).
				AddRow("age"))

		err = ds.GetFilterDimension("123", "age")
		So(err, ShouldBeNil)
		So(mock.ExpectationsWereMet(), ShouldBeNil)
	})
}

func TestGetFilterDimensionOption(t *testing.T) {
	t.Parallel()
	Convey("When a filterJobId, dimension name and option are provided, successfully return without erroring", t, func() {
		mock, db := NewSQLMockWithSQLStatements()
		ds, err := NewDatastore(db)
		So(err, ShouldBeNil)
		mock.ExpectQuery(getFilterSQL).WithArgs(sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"filterJobId", "datasetFilterId", "state"}).
				AddRow("123", "12345678", "created"))
		mock.ExpectQuery(getDimensionSQL).WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"name"}).
				AddRow("age"))
		mock.ExpectQuery(getDimensionOptionSQL).WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"name"}).
				AddRow("26"))

		err = ds.GetFilterDimensionOption("123", "age", "26")
		So(err, ShouldBeNil)
		So(mock.ExpectationsWereMet(), ShouldBeNil)
	})
}

func TestRemoveFilterDimension(t *testing.T) {
	t.Parallel()
	Convey("When a filterJobId and dimension name are provided, successfuly remove dimension from filter job", t, func() {
		mock, db := NewSQLMockWithSQLStatements()
		ds, err := NewDatastore(db)
		So(err, ShouldBeNil)
		mock.ExpectBegin()
		mock.ExpectPrepare(getFilterSQL).ExpectQuery().WithArgs(sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"filterJobId", "datasetFilterId", "state"}).
				AddRow("123", "12345678", "created"))
		mock.ExpectPrepare(getDimensionSQL).ExpectQuery().WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"name"}).
				AddRow("age"))
		mock.ExpectPrepare(deleteDimensionSQL).ExpectExec().WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err = ds.RemoveFilterDimension("123", "age")
		So(err, ShouldBeNil)
		So(mock.ExpectationsWereMet(), ShouldBeNil)
	})
}

func TestRemoveFilterDimensionOption(t *testing.T) {
	t.Parallel()
	Convey("When a filterJobId, dimension name and option are provided, successfuly remove dimension option from filter job", t, func() {
		mock, db := NewSQLMockWithSQLStatements()
		ds, err := NewDatastore(db)
		So(err, ShouldBeNil)
		mock.ExpectBegin()
		mock.ExpectPrepare(getFilterSQL).ExpectQuery().WithArgs(sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"filterJobId", "datasetFilterId", "state"}).
				AddRow("123", "12345678", "created"))
		mock.ExpectPrepare(getDimensionSQL).ExpectQuery().WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"name"}).
				AddRow("age"))
		mock.ExpectPrepare(deleteDimensionSQL).ExpectExec().WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err = ds.RemoveFilterDimensionOption("123", "age", "26")
		So(err, ShouldBeNil)
		So(mock.ExpectationsWereMet(), ShouldBeNil)
	})
}

func TestUpdateFilter(t *testing.T) {
	t.Parallel()
	Convey("successfuly update filter job", t, func() {
		Convey("when an unauthenticated user updates the state from 'created' to 'submitted'", func() {
			mock, db := NewSQLMockWithSQLStatements()
			ds, err := NewDatastore(db)
			So(err, ShouldBeNil)
			mock.ExpectBegin()
			mock.ExpectPrepare(getFilterStateSQL).ExpectQuery().WithArgs(sqlmock.AnyArg()).
				WillReturnRows(sqlmock.NewRows([]string{"state"}).
					AddRow("created"))
			mock.ExpectPrepare(updateFilterStateSQL).ExpectExec().WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
				WillReturnResult(sqlmock.NewResult(1, 1))
			mock.ExpectCommit()

			filter := &models.Filter{
				FilterID:        "123",
				State:           "submitted",
				DatasetFilterID: "12345678",
			}

			err = ds.UpdateFilter(false, filter)
			So(err, ShouldBeNil)
			So(mock.ExpectationsWereMet(), ShouldBeNil)
		})

		Convey("when an authenticated user updates downloads object", func() {
			mock, db := NewSQLMockWithSQLStatements()
			ds, err := NewDatastore(db)
			So(err, ShouldBeNil)
			mock.ExpectBegin()
			mock.ExpectPrepare(getFilterStateSQL).ExpectQuery().WithArgs(sqlmock.AnyArg()).
				WillReturnRows(sqlmock.NewRows([]string{"state"}).
					AddRow("submitted"))
			mock.ExpectPrepare(updateFilterStateSQL).ExpectExec().WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
				WillReturnResult(sqlmock.NewResult(1, 1))
			mock.ExpectPrepare(upsertDownloadURLSQL).ExpectExec().WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
				WillReturnResult(sqlmock.NewResult(1, 1))
			mock.ExpectPrepare(upsertDownloadURLSQL).ExpectExec().WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
				WillReturnResult(sqlmock.NewResult(1, 1))
			mock.ExpectPrepare(upsertDownloadURLSQL).ExpectExec().WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
				WillReturnResult(sqlmock.NewResult(1, 1))
			mock.ExpectCommit()

			filter := &models.Filter{
				FilterID:        "123",
				State:           "completed",
				DatasetFilterID: "12345678",
				Downloads: models.Downloads{
					CSV: models.DownloadItem{
						Size: "12mb",
						URL:  "s3/csvURL",
					},
					JSON: models.DownloadItem{
						Size: "4mb",
						URL:  "s3/jsonURL",
					},
					XLS: models.DownloadItem{
						Size: "40mb",
						URL:  "s3/xlsURL",
					},
				},
			}

			err = ds.UpdateFilter(true, filter)
			So(err, ShouldBeNil)
			So(mock.ExpectationsWereMet(), ShouldBeNil)
		})
	})
}

func TestConvertError(t *testing.T) {
	t.Parallel()
	Convey("Testing convert sql error function", t, func() {

		Convey("when receiving an SQL number of rows error, successfully return \"Not found\" error", func() {
			err := convertError(sql.ErrNoRows, "")
			So(err, ShouldNotBeNil)
			So(err, ShouldResemble, errors.New("Not found"))
		})

		Convey("when receiving an SQL number of rows error and a string type of \"filter job not found\", successfuly return \"Bad request\" error", func() {
			err := convertError(sql.ErrNoRows, "filter job not found")
			So(err, ShouldNotBeNil)
			So(err, ShouldResemble, errors.New("Bad request - filter job not found"))
		})

		Convey("when receiving an SQL number of rows error and a string type of \"bad request - dimension not found\", successfuly return \"Bad request\" error", func() {
			err := convertError(sql.ErrNoRows, "bad request - dimension not found")
			So(err, ShouldNotBeNil)
			So(err, ShouldResemble, errors.New("Bad request - dimension not found"))
		})

		Convey("when receiving an SQL number of rows error and a string type of \"dimension not found\", successfuly return \"Dimension not found\" error", func() {
			err := convertError(sql.ErrNoRows, "dimension not found")
			So(err, ShouldNotBeNil)
			So(err, ShouldResemble, errors.New("Dimension not found"))
		})

		Convey("when receiving an SQL number of rows error and a string type of \"option not found\", successfuly return \"Option not found\" error", func() {
			err := convertError(sql.ErrNoRows, "option not found")
			So(err, ShouldNotBeNil)
			So(err, ShouldResemble, errors.New("Option not found"))
		})

		Convey("when receiving a generic error (not an sql error), successfuly return same error", func() {
			err := convertError(fmt.Errorf("not an sql error"), "")
			So(err, ShouldNotBeNil)
			So(err, ShouldResemble, errors.New("not an sql error"))
		})

		Convey("when no error is passed in, return nil", func() {
			err := convertError(nil, "")
			So(err, ShouldBeNil)
		})

		Convey("when receiving a postgres error of duplicate key error, successfully return \"Bad request\" error", func() {
			postgresError := &pq.Error{
				Constraint: "filterjobdimensionoption",
			}
			err := convertError(postgresError, "")
			So(err, ShouldNotBeNil)
			So(err, ShouldResemble, errors.New("Bad request"))
		})

		Convey("when receiving a postgres error not due to duplicate key error, successfully return error", func() {
			postgresError := &pq.Error{
				Code: pq.ErrorCode("404"),
			}
			err := convertError(postgresError, "")
			So(err, ShouldNotBeNil)
			So(err, ShouldEqual, postgresError)
		})
	})
}

func NewSQLMockWithSQLStatements() (sqlmock.Sqlmock, *sql.DB) {
	db, mock, err := sqlmock.New()
	So(err, ShouldBeNil)
	mock.ExpectBegin()
	mock.MatchExpectationsInOrder(false)
	mock.ExpectPrepare(addFilterSQL)
	mock.ExpectPrepare(addDimensionSQL)
	mock.ExpectPrepare(addDimensionOptionSQL)
	mock.ExpectPrepare(deleteDimensionSQL)
	mock.ExpectPrepare(deleteDimensionOptionSQL)
	mock.ExpectPrepare(getFilterSQL)
	mock.ExpectPrepare(getFilterStateSQL)
	mock.ExpectPrepare(getDimensionSQL)
	mock.ExpectPrepare(getDimensionsSQL)
	mock.ExpectPrepare(getDimensionOptionsSQL)
	mock.ExpectPrepare(getDimensionOptionSQL)
	mock.ExpectPrepare(getDownloadItemsSQL)
	mock.ExpectPrepare(updateFilterStateSQL)
	mock.ExpectPrepare(upsertDimensionOptionSQL)
	mock.ExpectPrepare(upsertDownloadURLSQL)
	_, dbError := db.Begin()
	So(dbError, ShouldBeNil)
	return mock, db
}
