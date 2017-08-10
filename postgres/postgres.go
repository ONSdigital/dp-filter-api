package postgres

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/ONSdigital/dp-filter-api/models"
	"github.com/ONSdigital/go-ns/log"
)

const submittedState = "submitted"

// Datastore represents a structure to hold SQL statements to be used to gather information or insert about filters and dimensions
type Datastore struct {
	db                    *sql.DB
	addFilter             *sql.Stmt
	addDimension          *sql.Stmt
	addDimensionOption    *sql.Stmt
	getFilter             *sql.Stmt
	getFilterState        *sql.Stmt
	getDimension          *sql.Stmt
	getDimensions         *sql.Stmt
	getDimensionOptions   *sql.Stmt
	getDimensionOption    *sql.Stmt
	getDownloadItems      *sql.Stmt
	deleteDimension       *sql.Stmt
	upsertDimensionOption *sql.Stmt
}

// NewDatastore manages a postgres datastore used to store and find information about filters and dimensions
func NewDatastore(db *sql.DB) (Datastore, error) {
	addFilter, err := prepare("INSERT INTO Filters(filterJobId, datasetfilterID, state) VALUES($1, $2, $3)", db)
	addDimension, err := prepare("INSERT INTO Dimensions(filterJobId, name) VALUES($1, $2)", db)
	addDimensionOption, err := prepare("INSERT INTO Dimensions(filterJobId, name, option) VALUES($1, $2, $3)", db)
	getFilter, err := prepare("SELECT * FROM Filters WHERE filterJobId = $1", db)
	getFilterState, err := prepare("SELECT state FROM Filters WHERE filterJobId = $1", db)
	getDimension, err := prepare("SELECT name FROM Dimensions WHERE filterJobId = $1 AND name = $2", db)
	getDimensions, err := prepare("SELECT name FROM Dimensions WHERE filterJobId = $1", db)
	getDimensionOptions, err := prepare("SELECT option FROM Dimensions WHERE filterJobId = $1 AND name = $2", db)
	getDimensionOption, err := prepare("SELECT option FROM Dimensions WHERE filterJobId = $1 AND name = $2 AND option = $3", db)
	getDownloadItems, err := prepare("SELECT size, type, url FROM Downloads WHERE filterJobId = $1", db)
	deleteDimension, err := prepare("DELETE FROM Dimensions WHERE filterJobId = $1 AND name = $2", db)
	upsertDimensionOption, err := prepare("INSERT INTO Dimensions(filterJobId, name, option) VALUES($1, $2, $3) ON CONFLICT ON CONSTRAINT filterJobDimensionOption DO UPDATE SET option = $3", db)
	if err != nil {
		return Datastore{db: db, addFilter: addFilter, addDimension: addDimension, addDimensionOption: addDimensionOption, getFilter: getFilter, getFilterState: getFilterState, getDimensions: getDimensions, getDimension: getDimension, getDimensionOptions: getDimensionOptions, getDimensionOption: getDimensionOption, getDownloadItems: getDownloadItems, deleteDimension: deleteDimension, upsertDimensionOption: upsertDimensionOption}, err
	}

	return Datastore{db: db, addFilter: addFilter, addDimension: addDimension, addDimensionOption: addDimensionOption, getFilter: getFilter, getFilterState: getFilterState, getDimensions: getDimensions, getDimension: getDimension, getDimensionOptions: getDimensionOptions, getDimensionOption: getDimensionOption, getDownloadItems: getDownloadItems, deleteDimension: deleteDimension, upsertDimensionOption: upsertDimensionOption}, nil
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

	if err := ds.createDimensions(tx, newFilter.FilterID, newFilter.Dimensions); err != nil {
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

// AddFilterDimension adds a dimension for a given dataset to be stored in postgres
func (ds Datastore) AddFilterDimension(dimensionObject *models.AddDimension) error {
	tx, err := ds.db.Begin()
	if err != nil {
		return err
	}

	// Check filter exists
	row := ds.getFilter.QueryRow(dimensionObject.FilterID)

	var filterJobID, datasetFilterID, state sql.NullString

	if err := row.Scan(&filterJobID, &datasetFilterID, &state); err != nil {
		return convertSQLError(err, "bad request")
	}

	// Check filter is not locked (if state is equal to `submitted`)
	if state.String == submittedState {
		return errors.New("Forbidden")
	}

	// Check if dimension already exists, if it does then remove all options against dimension
	_, err = ds.getDimension.Query(dimensionObject.FilterID, dimensionObject.Name)
	if err != nil {
		if err != sql.ErrNoRows {
			return convertSQLError(err, "")
		}
	} else {
		log.Trace("dimension options already exist", log.Data{"filter_job_id": dimensionObject.FilterID, "dimension": dimensionObject.Name})
		// Remove rows continaing dimension options
		if err = ds.removeDimension(dimensionObject); err != nil {
			return err
		}
		log.Trace("dimension successfully deleted", log.Data{"filter_job_id": dimensionObject.FilterID, "dimension": dimensionObject.Name})
	}

	// Add any values against dimension
	if len(dimensionObject.Options) > 0 {
		dimension := models.Dimension{
			Name:   dimensionObject.Name,
			Values: dimensionObject.Options,
		}

		if err = ds.createDimensionOptions(tx, dimensionObject.FilterID, dimension); err != nil {
			if err = tx.Rollback(); err != nil {
				return err
			}
			return err
		}
	} else {
		if err = ds.addSingleDimension(tx, dimensionObject.FilterID, dimensionObject.Name); err != nil {
			if err = tx.Rollback(); err != nil {
				return err
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

// AddFilterDimensionOption adds an option to existing dimension for a given dataset stored in postgres
func (ds Datastore) AddFilterDimensionOption(dimensionOptionObject *models.AddDimensionOption) error { // Check filter exists
	row := ds.getFilter.QueryRow(dimensionOptionObject.FilterID)

	var filterJobID, datasetFilterID, state sql.NullString

	if err := row.Scan(&filterJobID, &datasetFilterID, &state); err != nil {
		return convertSQLError(err, "bad request")
	}

	// Check filter is not locked (if state is equal to `submitted`)
	if state.String == submittedState {
		return errors.New("Forbidden")
	}

	// Check if dimension exists
	_, err := ds.getDimension.Query(dimensionOptionObject.FilterID, dimensionOptionObject.Name)
	if err != nil {
		return convertSQLError(err, "dimension not found")
	}

	log.Debug("about to upsert", nil)
	if _, err = ds.upsertDimensionOption.Exec(dimensionOptionObject.FilterID, dimensionOptionObject.Name, dimensionOptionObject.Option); err != nil {
		return err
	}
	log.Debug("finished upsert", nil)

	return nil
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

// RemoveFilterDimension deletes a dimension if it exists for a filter job
func (ds Datastore) RemoveFilterDimension(filterID string, name string) error {
	checkFilterJobExists := ds.getFilter.QueryRow(filterID)

	var filterJobID, datasetFilterID, state sql.NullString

	if err := checkFilterJobExists.Scan(&filterJobID, &datasetFilterID, &state); err != nil {
		return convertSQLError(err, "bad request")
	}

	if state.String == submittedState {
		err := errors.New("Forbidden")
		log.ErrorC("update to filter job forbidden, job already submitted", err, log.Data{"filter_job_id": filterID})
		return err
	}

	checkDimensionExists := ds.getDimension.QueryRow(filterID, name)

	var dimensionName sql.NullString

	if err := checkDimensionExists.Scan(&dimensionName); err != nil {
		return convertSQLError(err, "dimension not found")
	}

	dimensionObject := &models.AddDimension{
		FilterID: filterID,
		Name:     name,
	}

	err := ds.removeDimension(dimensionObject)
	if err != nil {
		return err
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

// addSingleDimension creates a single dimension and relates it to filter job
func (ds Datastore) addSingleDimension(tx *sql.Tx, filterID string, name string) error {
	log.Info("about to update dimension", nil)
	help, err := tx.Stmt(ds.addDimension).Exec(filterID, name)
	if err != nil {
		log.ErrorC("got an error", err, log.Data{"what am i": help})
		return err
	}
	return nil
}

// addDimensions creates many dimensions and respective options and relates it to filter job
func (ds Datastore) createDimensions(tx *sql.Tx, filterID string, dimensions []models.Dimension) error {
	for _, dimension := range dimensions {
		if err := ds.createDimensionOptions(tx, filterID, dimension); err != nil {
			return err
		}
	}

	return nil
}

// addDimensionOptions method adds options of a single dimension to be stored in postgres and relates it to filter job
func (ds Datastore) createDimensionOptions(tx *sql.Tx, filterID string, dimension models.Dimension) error {
	for _, value := range dimension.Values {
		_, err := tx.Stmt(ds.addDimensionOption).Exec(filterID, dimension.Name, value)
		if err != nil {
			return err
		}
	}

	return nil
}

// removeDimension method deletes all dimensions that are currently stored in postgres against a filter job
func (ds Datastore) removeDimension(dimensionObject *models.AddDimension) error {
	_, err := ds.deleteDimension.Exec(dimensionObject.FilterID, dimensionObject.Name)
	if err != nil {
		return err
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

	if state == submittedState {
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
