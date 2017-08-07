package postgres

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/ONSdigital/dp-filter-api/models"
	"github.com/ONSdigital/go-ns/log"
)

// Datastore represents a structure to hold SQL statements to be used to gather information or insert about filters and dimensions
type Datastore struct {
	db                 *sql.DB
	addFilter          *sql.Stmt
	addDimension       *sql.Stmt
	getFilter          *sql.Stmt
	getFilterState     *sql.Stmt
	getDimensionValues *sql.Stmt
	getDownloadItems   *sql.Stmt
}

// NewDatastore manages a postgres datastore used to store and find information about filters and dimensions
func NewDatastore(db *sql.DB) (Datastore, error) {
	addFilter, err := prepare("INSERT INTO Filters(filterJobId, datasetfilterID, state) VALUES($1, $2, $3)", db)
	addDimension, err := prepare("INSERT INTO Dimensions(filterJobId, name, value) VALUES($1, $2, $3)", db)
	getFilter, err := prepare("SELECT * FROM Filters WHERE filterJobId = $1", db)
	getFilterState, err := prepare("SELECT state FROM Filters WHERE filterJobId = $1", db)
	getDimensionValues, err := prepare("SELECT name, value FROM Dimensions WHERE filterJobId = $1", db)
	getDownloadItems, err := prepare("SELECT size, type, url FROM Downloads WHERE filterJobId = $1", db)
	if err != nil {
		return Datastore{db: db, addFilter: addFilter, addDimension: addDimension, getFilter: getFilter, getFilterState: getFilterState, getDimensionValues: getDimensionValues, getDownloadItems: getDownloadItems}, err
	}

	return Datastore{db: db, addFilter: addFilter, addDimension: addDimension, getFilter: getFilter, getFilterState: getFilterState, getDimensionValues: getDimensionValues, getDownloadItems: getDownloadItems}, nil
}

// AddFilter adds a filter for a given dataset to be stored in postgres
func (ds Datastore) AddFilter(host string, newFilter *models.Filter) (models.Filter, error) {
	tx, err := ds.db.Begin()
	if err != nil {
		return models.Filter{}, err
	}

	_, err = tx.Stmt(ds.addFilter).Exec(newFilter.FilterID, newFilter.DataSetFilterID, newFilter.State)
	if err != nil {
		return models.Filter{}, err
	}

	if err := ds.addDimensions(tx, newFilter.FilterID, newFilter.Dimensions); err != nil {
		if err = tx.Rollback(); err != nil {
			return models.Filter{}, err
		}
		return models.Filter{}, err
	}

	if err := tx.Commit(); err != nil {
		return models.Filter{}, err
	}

	return *newFilter, nil
}

// GetFilter gets a filter for a given dataset that is stored in postgres
func (ds Datastore) GetFilter(filterID string) (models.Filter, error) {
	var filterJob models.Filter

	tx, err := ds.db.Begin()
	if err != nil {
		return filterJob, err
	}

	row := tx.Stmt(ds.getFilter).QueryRow(filterID)

	var filterJobID, datasetFilterID, state sql.NullString

	if err := row.Scan(&filterJobID, &datasetFilterID, &state); err != nil {
		return filterJob, convertSQLError(err)
	}

	filterJob.DataSetFilterID = datasetFilterID.String
	filterJob.FilterID = filterID
	filterJob.State = state.String
	filterJob.DimensionListURL = "/filters/" + filterID + "/dimensions"

	downloadRows, err := ds.getDownloadItems.Query(filterID)
	if err != nil {
		return filterJob, convertSQLError(err)
	}

	downloads := models.Downloads{}
	for downloadRows.Next() {
		var size, downloadType, url sql.NullString
		err := downloadRows.Scan(&size, &downloadType, &url)
		if err != nil {
			return filterJob, err
		}

		if downloadType.String == "csv" {
			downloads.CSV.Size = size.String
			downloads.CSV.URL = url.String
		}

		if downloadType.String == "json" {
			downloads.JSON.Size = size.String
			downloads.JSON.URL = url.String
		}

		if downloadType.String == "xls" {
			downloads.XLS.Size = size.String
			downloads.XLS.URL = url.String
		}
	}

	filterJob.Downloads = downloads

	return filterJob, nil
}

// UpdateFilter updates a given filter against a given dataset in the postgres databaseilter job
func (ds Datastore) UpdateFilter(host string, filter *models.Filter) error {
	tx, err := ds.db.Begin()
	if err != nil {
		return err
	}

	currentFilterJob, err := getFilterJobState(tx, ds, filter.FilterID)
	if err != nil {
		log.Error(err, log.Data{"filter_job_id": filter.FilterID, "current_filter_job": currentFilterJob})
		return err
	}

	updateStatement, err := updateStatement(filter)
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(updateStatement)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(filter.FilterID)
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	log.Debug("successfully committed filter job update", log.Data{"filter_job_id": filter.FilterID, "statement": updateStatement})

	return nil
}

// addDimensions - Add dimensions and relate it to filter.
func (ds Datastore) addDimensions(tx *sql.Tx, filterID string, dimensions []models.Dimension) error {
	for _, dimension := range dimensions {
		if err := ds.addDimensionValues(tx, filterID, dimension); err != nil {
			return err
		}
	}

	return nil
}

// addDimensionValues method adds a single dimension to be stored in postgres and relates it to filter job
func (ds Datastore) addDimensionValues(tx *sql.Tx, filterID string, dimension models.Dimension) error {
	for _, value := range dimension.Values {
		_, err := tx.Stmt(ds.addDimension).Exec(filterID, dimension.Name, value)
		if err != nil {
			return err
		}
	}

	return nil
}

func getFilterJobState(tx *sql.Tx, ds Datastore, filterID string) (models.Filter, error) {
	var filterJob models.Filter
	row := tx.Stmt(ds.getFilterState).QueryRow(filterID)

	var filter sql.NullString

	if err := row.Scan(&filter); err != nil {
		return filterJob, convertSQLError(err)
	}

	state := filter.String

	filterJob.FilterID = filterID
	filterJob.State = state

	if state == "submitted" {
		err := errors.New("Forbidden")
		log.ErrorC("update to filter job forbidden, job already submitted", err, log.Data{"filter_job_id": filterID})
		return filterJob, err
	}

	return filterJob, nil
}

// This could be simplified by removing this and adding the statement to the
// NewDatastore method but there will be other fields that will be available to
// update in the near future and hence this function will stay here
func updateStatement(filter *models.Filter) (string, error) {
	statement := "UPDATE Filters SET"
	var hasUpdate bool

	if filter.State != "" {
		hasUpdate = true
		statement = statement + " state = '" + filter.State + "',"
	}

	if !hasUpdate {
		err := errors.New("Bad request")
		log.ErrorC("missing a filter field that can be updated", err, log.Data{"filter": filter})
		return "", err
	}

	statement = strings.TrimSuffix(statement, ",")
	statement = statement + " WHERE filterJobId = $1 RETURNING filterJobId"

	return statement, nil
}

func prepare(sql string, db *sql.DB) (*sql.Stmt, error) {
	statement, err := db.Prepare(sql)
	if err != nil {
		return statement, err
	}
	return statement, nil
}

func convertSQLError(err error) error {
	switch {
	case err == sql.ErrNoRows:
		return errors.New("Not found")
	case err != nil:
		return err
	}
	return nil
}
