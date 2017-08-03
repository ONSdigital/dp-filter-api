package postgres

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/ONSdigital/dp-filter-api/models"
	"github.com/ONSdigital/go-ns/log"
)

// Datastore represents a structure to hold SQL statements to be used to gather information or insert about filters and dimensions
type Datastore struct {
	db           *sql.DB
	addFilter    *sql.Stmt
	addDimension *sql.Stmt
	getFilter    *sql.Stmt
}

// NewDatastore manages a postgres datastore used to store and find information about filters and dimensions
func NewDatastore(db *sql.DB) (Datastore, error) {
	addFilter, err := prepare("INSERT INTO Filters(filterId, dataset, edition, version, state, filter) VALUES($1, $2, $3, $4, $5, $6)", db)
	addDimension, err := prepare("INSERT INTO Dimensions(filterId, name, value) VALUES($1, $2, $3)", db)
	getFilter, err := prepare("SELECT state FROM Filters WHERE filterId = $1", db)
	if err != nil {
		return Datastore{db: db, addFilter: addFilter, addDimension: addDimension, getFilter: getFilter}, err
	}

	return Datastore{db: db, addFilter: addFilter, addDimension: addDimension, getFilter: getFilter}, nil
}

// AddFilter adds a filter for a given dataset to be stored in postgres
func (ds Datastore) AddFilter(host string, newFilter *models.Filter) (models.Filter, error) {
	bytes, err := json.Marshal(newFilter)
	if err != nil {
		return models.Filter{}, err
	}

	tx, err := ds.db.Begin()
	if err != nil {
		return models.Filter{}, err
	}

	_, err = tx.Stmt(ds.addFilter).Exec(newFilter.FilterID, newFilter.DataSet, newFilter.Edition, newFilter.Version, newFilter.State, bytes)
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

// UpdateFilter updates a given filter against a given dataset in the postgres databaseilter job
func (ds Datastore) UpdateFilter(host string, filter *models.Filter) error {
	tx, err := ds.db.Begin()
	if err != nil {
		return err
	}

	currentFilterJob, err := getFilterJob(tx, ds, filter.FilterID)
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

func getFilterJob(tx *sql.Tx, ds Datastore, filterID string) (models.Filter, error) {
	var filterJob models.Filter
	row := tx.Stmt(ds.getFilter).QueryRow(filterID)

	var filter sql.NullString

	if err := row.Scan(&filter); err != nil {
		return filterJob, convertSQLError(err)
	}

	state := filter.String

	filterJob.FilterID = filterID
	filterJob.State = state

	if state == "submitted" {
		err := fmt.Errorf("Forbidden")
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
		err := fmt.Errorf("Bad request")
		log.ErrorC("missing a filter field that can be updated", err, log.Data{"filter": filter})
		return "", err
	}

	statement = strings.TrimSuffix(statement, ",")
	statement = statement + " WHERE filterId = $1 RETURNING filterId"

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
