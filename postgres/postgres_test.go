package postgres

import (
	"database/sql"
	"testing"

	sqlmock "github.com/go-sqlmock"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	addFilterSQL    = "INSERT INTO Filters"
	addDimensionSQL = "INSERT INTO Dimensions"
)

func NewSQLMockWithSQLStatements() (sqlmock.Sqlmock, *sql.DB) {
	db, mock, err := sqlmock.New()
	So(err, ShouldBeNil)
	mock.ExpectBegin()
	mock.MatchExpectationsInOrder(false)
	mock.ExpectPrepare(addFilterSQL)
	mock.ExpectPrepare(addDimensionSQL)
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
