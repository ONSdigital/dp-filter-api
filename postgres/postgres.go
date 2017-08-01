package postgres

import (
	"database/sql"
	"encoding/json"

	"github.com/ONSdigital/dp-filter-api/models"
)

// Datastore represents a structure to hold SQL statements to be used to gather information or insert about filters and dimensions
type Datastore struct {
	db           *sql.DB
	addFilter    *sql.Stmt
	addDimension *sql.Stmt
}

// NewDatastore manages a postgres datastore used to store and find information about filters and dimensions
func NewDatastore(db *sql.DB) (Datastore, error) {
	addFilter, err := prepare("INSERT INTO Filters(filterId, dataset, edition, version, state, filter) VALUES($1, $2, $3, $4, $5, $6)", db)
	addDimension, err := prepare("INSERT INTO Dimensions(filterId, name, value) VALUES($1, $2, $3)", db)
	if err != nil {
		return Datastore{db: db, addFilter: addFilter, addDimension: addDimension}, err
	}

	return Datastore{db: db, addFilter: addFilter, addDimension: addDimension}, nil
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

// AddDimensions method adds dimensions to be stored in postgres and relates them to a filter job
func (ds Datastore) addDimensions(tx *sql.Tx, filterID string, dimensions []models.Dimension) error {
	for _, dimension := range dimensions {
		if err := ds.addDimensionValues(tx, filterID, dimension); err != nil {
			return err
		}
	}

	return nil
}

// AddDimension method adds a single dimension to be stored in postgres and relates it to filter job
func (ds Datastore) addDimensionValues(tx *sql.Tx, filterID string, dimension models.Dimension) error {
	for _, value := range dimension.Values {
		_, err := tx.Stmt(ds.addDimension).Exec(filterID, dimension.Name, value)
		if err != nil {
			return err
		}
	}

	return nil
}

func prepare(sql string, db *sql.DB) (*sql.Stmt, error) {
	statement, err := db.Prepare(sql)
	if err != nil {
		return statement, err
	}
	return statement, nil
}
