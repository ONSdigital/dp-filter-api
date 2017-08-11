package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/ONSdigital/dp-filter-api/models"
	sqlmock "github.com/go-sqlmock"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	addFilterSQL             = "INSERT INTO Filters"
	addDimensionSQL          = "INSERT INTO Dimensions"
	deleteDimensionSQL       = "DELETE FROM Dimensions WHERE"
	getFilterSQL             = "SELECT * FROM Filters WHERE"
	getFilterStateSQL        = "SELECT state FROM Filters WHERE"
	getDimensionValuesSQL    = "SELECT name, value FROM Dimensions WHERE"
	getDownloadsSQL          = "SELECT size, type, url FROM Downloads WHERE"
	upsertDimensionOptionSQL = "INSERT INTO Dimensions(.+) VALUES(.+) ON CONSTRAINT"
)

func NewSQLMockWithSQLStatements() (sqlmock.Sqlmock, *sql.DB) {
	db, mock, err := sqlmock.New()
	So(err, ShouldBeNil)
	mock.ExpectBegin()
	mock.MatchExpectationsInOrder(false)
	mock.ExpectPrepare(addFilterSQL)
	mock.ExpectPrepare(addDimensionSQL)
	mock.ExpectPrepare(deleteDimensionSQL)
	mock.ExpectPrepare(getFilterSQL)
	mock.ExpectPrepare(getFilterStateSQL)
	mock.ExpectPrepare(getDimensionValuesSQL)
	mock.ExpectPrepare(getDownloadsSQL)
	mock.ExpectPrepare(upsertDimensionOptionSQL)
	_, dbError := db.Begin()
	So(dbError, ShouldBeNil)
	return mock, db
}

func TestNewPostgresDatastore(t *testing.T) {
	t.Parallel()
	Convey("When creating a postgres datastore no errors are returned", t, func() {
		_, db := NewSQLMockWithSQLStatements()
		_, err := NewDatastore(db)
		So(err, ShouldBeNil)
	})
}

func TestUpdateStatement(t *testing.T) {
	t.Parallel()
	Convey("when update filter job has a state in json body successfully return statement", t, func() {
		filter := &models.Filter{
			State: "submitted",
		}

		statement, err := updateStatement(filter)
		So(err, ShouldBeNil)
		So(statement, ShouldEqual, "UPDATE Filters SET state = 'submitted' WHERE filterJobId = $1 RETURNING filterJobId")
	})

	Convey("when update filter job has a state and dataset in json body successfully return statement", t, func() {
		filter := &models.Filter{
			DatasetFilterID: "12345678",
			State:           "submitted",
		}

		statement, err := updateStatement(filter)
		So(err, ShouldBeNil)
		So(statement, ShouldEqual, "UPDATE Filters SET state = 'submitted' WHERE filterJobId = $1 RETURNING filterJobId")
	})

	Convey("when update filter job has only dataset_filter_id in json body return error", t, func() {
		filter := &models.Filter{
			DatasetFilterID: "12345678",
		}

		statement, err := updateStatement(filter)
		So(err, ShouldNotBeNil)
		So(err, ShouldResemble, fmt.Errorf("Bad request"))
		So(statement, ShouldEqual, "")
	})
}

func TestConvertSQLError(t *testing.T) {
	t.Parallel()
	Convey("Testing convert sql error function", t, func() {

		Convey("when receiving an SQL number of rows error, successfully return \"Not found\" error", func() {
			err := convertSQLError(sql.ErrNoRows, "")
			So(err, ShouldNotBeNil)
			So(err, ShouldResemble, errors.New("Not found"))
		})

		Convey("when receiving an SQL number of rows error and a string type of \"bad request\", successfuly return \"Bad request\" error", func() {
			err := convertSQLError(sql.ErrNoRows, "bad request")
			So(err, ShouldNotBeNil)
			So(err, ShouldResemble, errors.New("Bad request"))
		})

		Convey("when receiving an SQL number of rows error and a string type of \"dimension not found\", successfuly return \"Dimension not found\" error", func() {
			err := convertSQLError(sql.ErrNoRows, "dimension not found")
			So(err, ShouldNotBeNil)
			So(err, ShouldResemble, errors.New("Dimension not found"))
		})

		Convey("when receiving an SQL number of rows error and a string type of \"option not found\", successfuly return \"Option not found\" error", func() {
			err := convertSQLError(sql.ErrNoRows, "option not found")
			So(err, ShouldNotBeNil)
			So(err, ShouldResemble, errors.New("Option not found"))
		})

		Convey("when receiving a generic error (not an sql error), successfuly return same error", func() {
			err := convertSQLError(fmt.Errorf("not an sql error"), "")
			So(err, ShouldNotBeNil)
			So(err, ShouldResemble, errors.New("not an sql error"))
		})

		Convey("when no error is passed in, return nil", func() {
			err := convertSQLError(nil, "")
			So(err, ShouldBeNil)
		})
	})
}
