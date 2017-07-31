package postgres

import (
	"database/sql"
	"encoding/json"

	"github.com/ONSdigital/dp-filter-api/models"
)

// Datastore - A structure to hold SQL statements to be used to gather information or insert about filters and dimensions
type Datastore struct {
	db           *sql.DB
	addFilter    *sql.Stmt
	addDimension *sql.Stmt
}

// NewDatastore - Create a postgres datastore. This is used to store and find information about jobs and instances.
func NewDatastore(db *sql.DB) (Datastore, error) {
	addFilter := prepare("INSERT INTO Filters(filterId, dataset, edition, version, state, filter) VALUES($1, $2, $3, $4, $5, $6)", db)
	addDimension := prepare("INSERT INTO Dimensions(filterId, name, value) VALUES($1, $2, $3)", db)
	return Datastore{db: db, addFilter: addFilter, addDimension: addDimension}, nil
}

// AddFilter - Add a filter to be stored in postgres.
func (ds Datastore) AddFilter(host string, newfilter *models.Filter) (models.Filter, error) {
	bytes, err := json.Marshal(newfilter)
	if err != nil {
		return models.Filter{}, err
	}

	tx, err := ds.db.Begin()
	if err != nil {
		return models.Filter{}, err
	}

	_, err = tx.Stmt(ds.addFilter).Exec(newfilter.FilterID, newfilter.DataSet, newfilter.Edition, newfilter.Version, newfilter.State, bytes)
	if err != nil {
		return models.Filter{}, err
	}

	if err := ds.addDimensions(tx, newfilter.FilterID, newfilter.Dimensions); err != nil {
		if err = tx.Rollback(); err != nil {
			return models.Filter{}, err
		}
		return models.Filter{}, err
	}

	if err := tx.Commit(); err != nil {
		return models.Filter{}, err
	}

	return *newfilter, nil
}

// AddDimensions - Add dimensions and relate it to filter.
func (ds Datastore) addDimensions(tx *sql.Tx, filterID string, dimensions []models.Dimension) error {
	for _, dimension := range dimensions {
		if err := ds.addDimensionValues(tx, filterID, dimension); err != nil {
			return err
		}
	}

	return nil
}

// AddDimension - Add single dimension and relate it to filter.
func (ds Datastore) addDimensionValues(tx *sql.Tx, filterID string, dimension models.Dimension) error {
	for _, value := range dimension.Values {
		_, err := tx.Stmt(ds.addDimension).Exec(filterID, dimension.Name, value)
		if err != nil {
			return err
		}
	}

	return nil
}

func prepare(sql string, db *sql.DB) *sql.Stmt {
	statement, err := db.Prepare(sql)
	if err != nil {
		panic(err)
	}
	return statement
}
