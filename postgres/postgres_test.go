package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"testing"

	sqlmock "github.com/go-sqlmock"
	"github.com/lib/pq"
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
	upsertDownloadURLSQL     = "INSERT INTO Downloads(.+) VALUES(.+) ON CONSTRAINT"
	upsertFilterStateSQL     = "UPDATE Filters SET"
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
	mock.ExpectPrepare(upsertDownloadURLSQL)
	mock.ExpectPrepare(upsertFilterStateSQL)
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

func TestConvertError(t *testing.T) {
	t.Parallel()
	Convey("Testing convert sql error function", t, func() {

		Convey("when receiving an SQL number of rows error, successfully return \"Not found\" error", func() {
			err := convertError(sql.ErrNoRows, "")
			So(err, ShouldNotBeNil)
			So(err, ShouldResemble, errors.New("Not found"))
		})

		Convey("when receiving an SQL number of rows error and a string type of \"bad request\", successfuly return \"Bad request\" error", func() {
			err := convertError(sql.ErrNoRows, "bad request")
			So(err, ShouldNotBeNil)
			So(err, ShouldResemble, errors.New("Bad request"))
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
