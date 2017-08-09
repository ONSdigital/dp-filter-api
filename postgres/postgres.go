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
	db                  *sql.DB
	addFilter           *sql.Stmt
	addDimension        *sql.Stmt
	getFilter           *sql.Stmt
	getFilterState      *sql.Stmt
	getDimension        *sql.Stmt
	getDimensions       *sql.Stmt
	getDimensionOptions *sql.Stmt
	getDimensionOption  *sql.Stmt
	getDownloadItems    *sql.Stmt
}

// NewDatastore manages a postgres datastore used to store and find information about filters and dimensions
func NewDatastore(db *sql.DB) (Datastore, error) {
	addFilter, err := prepare("INSERT INTO Filters(filterJobId, datasetfilterID, state) VALUES($1, $2, $3)", db)
	addDimension, err := prepare("INSERT INTO Dimensions(filterJobId, name, option) VALUES($1, $2, $3)", db)
	getFilter, err := prepare("SELECT * FROM Filters WHERE filterJobId = $1", db)
	getFilterState, err := prepare("SELECT state FROM Filters WHERE filterJobId = $1", db)
	getDimension, err := prepare("SELECT name FROM Dimensions WHERE filterJobId = $1 AND name = $2", db)
	getDimensions, err := prepare("SELECT name FROM Dimensions WHERE filterJobId = $1", db)
	getDimensionOptions, err := prepare("SELECT option FROM Dimensions WHERE filterJobId = $1 AND name = $2", db)
	getDimensionOption, err := prepare("SELECT option FROM Dimensions WHERE filterJobId = $1 AND name = $2 AND option = $3", db)
	getDownloadItems, err := prepare("SELECT size, type, url FROM Downloads WHERE filterJobId = $1", db)
	if err != nil {
		return Datastore{db: db, addFilter: addFilter, addDimension: addDimension, getFilter: getFilter, getFilterState: getFilterState, getDimensions: getDimensions, getDimension: getDimension, getDimensionOptions: getDimensionOptions, getDimensionOption: getDimensionOption, getDownloadItems: getDownloadItems}, err
	}

	return Datastore{db: db, addFilter: addFilter, addDimension: addDimension, getFilter: getFilter, getFilterState: getFilterState, getDimensions: getDimensions, getDimension: getDimension, getDimensionOptions: getDimensionOptions, getDimensionOption: getDimensionOption, getDownloadItems: getDownloadItems}, nil
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
		return filterJob, convertSQLError(err, "")
	}

	filterJob.DataSetFilterID = datasetFilterID.String
	filterJob.FilterID = filterID
	filterJob.State = state.String
	filterJob.DimensionListURL = "/filters/" + filterID + "/dimensions"

	downloadRows, err := ds.getDownloadItems.Query(filterID)
	if err != nil {
		return filterJob, convertSQLError(err, "")
	}

	downloads := models.Downloads{}
	for downloadRows.Next() {
		var size, downloadType, url sql.NullString
		err := downloadRows.Scan(&size, &downloadType, &url)
		if err != nil {
			return filterJob, err
		}

		switch downloadType.String {
		case "csv":
			downloads.CSV.Size = size.String
			downloads.CSV.URL = url.String
		case "json":
			downloads.JSON.Size = size.String
			downloads.JSON.URL = url.String
		case "xls":
			downloads.XLS.Size = size.String
			downloads.XLS.URL = url.String
		}
	}

	filterJob.Downloads = downloads

	return filterJob, nil
}

// GetFilterDimensions gets all dimensions for a filter job
func (ds Datastore) GetFilterDimensions(filterID string) ([]models.Dimension, error) {
	dimensions := []models.Dimension{}

	checkFilterJobExists := ds.getFilter.QueryRow(filterID)

	var filterJobID, datasetFilterID, state sql.NullString

	if err := checkFilterJobExists.Scan(&filterJobID, &datasetFilterID, &state); err != nil {
		return dimensions, convertSQLError(err, "")
	}

	uniqueDimensions := make(map[string]string)

	dimensionRows, err := ds.getDimensions.Query(filterID)
	if err != nil {
		return dimensions, convertSQLError(err, "")
	}

	for dimensionRows.Next() {
		var name sql.NullString
		err := dimensionRows.Scan(&name)
		if err != nil {
			return dimensions, err
		}

		if _, ok := uniqueDimensions[name.String]; ok {
			continue
		}

		uniqueDimensions[name.String] = name.String

		dimensionURL := "/filters/" + filterID + "/dimensions/" + name.String

		dimensions = append(dimensions, models.Dimension{Name: name.String, DimensionURL: dimensionURL})
	}

	return dimensions, nil
}

// GetFilterDimensionOptions gets all options for a dimension of a filter job
func (ds Datastore) GetFilterDimensionOptions(filterID string, name string) (models.GetDimensionOptions, error) {
	var options models.GetDimensionOptions

	checkFilterJobExists := ds.getFilter.QueryRow(filterID)

	var filterJobID, datasetFilterID, state sql.NullString

	if err := checkFilterJobExists.Scan(&filterJobID, &datasetFilterID, &state); err != nil {
		return options, convertSQLError(err, "bad request")
	}

	checkDimensionExists := ds.getDimension.QueryRow(filterID, name)

	var dimensionName sql.NullString

	if err := checkDimensionExists.Scan(&dimensionName); err != nil {
		return options, convertSQLError(err, "dimension not found")
	}

	optionRows, err := ds.getDimensionOptions.Query(filterID, name)
	if err != nil {
		return options, convertSQLError(err, "")
	}

	var option sql.NullString
	var optionURLs []string

	for optionRows.Next() {
		err := optionRows.Scan(&option)
		if err != nil {
			return options, err
		}

		// If dimension exists but no options have been assigned continue
		if option.String == "" {
			continue
		}

		dimensionOptionURL := "/filters/" + filterID + "/dimensions/" + name + "/options/" + option.String

		optionURLs = append(optionURLs, dimensionOptionURL)
	}

	options.DimensionOptionURLs = optionURLs

	return options, nil
}

// GetFilterDimension checks if dimension exists for a filter job
func (ds Datastore) GetFilterDimension(filterID string, name string) error {
	checkFilterJobExists := ds.getFilter.QueryRow(filterID)

	var filterJobID, datasetFilterID, state sql.NullString

	if err := checkFilterJobExists.Scan(&filterJobID, &datasetFilterID, &state); err != nil {
		return convertSQLError(err, "bad request")
	}

	checkDimensionExists := ds.getDimension.QueryRow(filterID, name)

	var dimensionName sql.NullString

	if err := checkDimensionExists.Scan(&dimensionName); err != nil {
		return convertSQLError(err, "")
	}

	return nil
}

// GetFilterDimensionOption checks if option exists for a dimension on a filter job
func (ds Datastore) GetFilterDimensionOption(filterID string, name string, option string) error {
	checkFilterJobExists := ds.getFilter.QueryRow(filterID)

	var filterJobID, datasetFilterID, state sql.NullString

	if err := checkFilterJobExists.Scan(&filterJobID, &datasetFilterID, &state); err != nil {
		return convertSQLError(err, "bad request")
	}

	checkDimensionExists := ds.getDimension.QueryRow(filterID, name)

	var dimensionName sql.NullString

	if err := checkDimensionExists.Scan(&dimensionName); err != nil {
		return convertSQLError(err, "bad request")
	}

	checkDimensionOptionExists := ds.getDimensionOption.QueryRow(filterID, name, option)
	var dimensionOption sql.NullString

	if err := checkDimensionOptionExists.Scan(&dimensionOption); err != nil {
		return convertSQLError(err, "option not found")
	}

	return nil
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
		return filterJob, convertSQLError(err, "")
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

func convertSQLError(err error, typ string) error {
	switch {
	case err == sql.ErrNoRows:
		if typ == "bad request" {
			return errors.New("Bad request")
		}
		if typ == "dimension not found" {
			return errors.New("Dimension not found")
		}
		if typ == "option not found" {
			return errors.New("Option not found")
		}
		return errors.New("Not found")
	case err != nil:
		return err
	}
	return nil
}
