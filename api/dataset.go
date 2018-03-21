package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"fmt"

	"strconv"

	"github.com/ONSdigital/dp-filter-api/models"
	"github.com/ONSdigital/go-ns/clients/dataset"
	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/go-ns/rchttp"
	"github.com/ONSdigital/go-ns/identity"
)

// DatasetAPIer - An interface used to access the DatasetAPI
type DatasetAPIer interface {
	GetVersionDimensions(ctx context.Context, dataset models.Dataset) (*models.DatasetDimensionResults, error)
	GetVersionDimensionOptions(ctx context.Context, dataset models.Dataset, dimension string) (*models.DatasetDimensionOptionResults, error)
	GetVersion(ctx context.Context, dataset models.Dataset) (version *models.Version, err error)
}

// DatasetAPI aggregates a client and URL and other common data for accessing the API
type DatasetAPI struct {
	client    *rchttp.Client
	url       string
	authToken string
}

// NewDatasetAPI creates an DatasetAPI object
func NewDatasetAPI(client *rchttp.Client, datasetAPIURL, datasetAPIAuthToken string) *DatasetAPI {
	return &DatasetAPI{
		client:    client,
		url:       datasetAPIURL,
		authToken: datasetAPIAuthToken,
	}
}

// A list of errors that the dataset package could return
var (
	ErrUnexpectedStatusCode     = errors.New("unexpected status code from api")
	ErrVersionNotFound          = errors.New("Version not found")
	ErrDimensionNotFound        = errors.New("Dimension not found")
	ErrDimensionOptionsNotFound = errors.New("Dimension options not found")

	publishedState  = "published"
	versionNotFound = "Version not found"
)

// GetVersion queries the Dataset API to get an version
func (api *DatasetAPI) GetVersion(ctx context.Context, d models.Dataset) (version *models.Version, err error) {
	path := fmt.Sprintf("%s/datasets/%s/editions/%s/versions/%d", api.url, d.ID, d.Edition, d.Version)
	logData := log.Data{"func": "GetDataset", "URL": path, "dataset": d}

	jsonResult, httpCode, err := api.get(ctx, path, nil)
	logData["httpCode"] = httpCode
	logData["jsonResult"] = jsonResult
	if err != nil {
		log.ErrorC("api get", err, logData)
		return nil, handleError(httpCode, err, "dataset")
	}

	version = &models.Version{}
	if err = json.Unmarshal(jsonResult, version); err != nil {
		log.ErrorC("unmarshal", err, logData)
		return
	}

	// External facing customers should NOT be able to filter an unpublished version

	if version.State != publishedState && !identity.IsPresent(ctx) {
		log.Error(errors.New("invalid authorization, returning not found status"), log.Data{"dataset": d})
		return nil, ErrVersionNotFound
	}

	return
}

// GetVersionDimensions queries the Dataset API to get a list of dimensions
func (api *DatasetAPI) GetVersionDimensions(ctx context.Context, dataset models.Dataset) (dimensions *models.DatasetDimensionResults, err error) {
	path := api.url + "/datasets/" + dataset.ID + "/editions/" + dataset.Edition + "/versions/" + strconv.Itoa(dataset.Version) + "/dimensions"
	logData := log.Data{"func": "GetVersionDimensions", "URL": path, "dataset": dataset}

	bytesResult, httpCode, err := api.get(ctx, path, nil)
	logData["httpCode"] = httpCode
	jsonResult := strings.TrimSpace(string(bytesResult))
	logData["jsonResult"] = jsonResult
	if err != nil {
		log.ErrorC("GetVersionDimensions api get", err, logData)
		typ := "dimension"
		if jsonResult == versionNotFound {
			typ = "version"
		}
		return nil, handleError(httpCode, err, typ)
	}

	dimensions = &models.DatasetDimensionResults{}
	if err = json.Unmarshal(bytesResult, dimensions); err != nil {
		log.ErrorC("GetVersionDimensions unmarshal", err, logData)
		return
	}

	return
}

// GetVersionDimensionOptions queries the Dataset API to get a list of dimension options
func (api *DatasetAPI) GetVersionDimensionOptions(ctx context.Context, dataset models.Dataset, dimension string) (options *models.DatasetDimensionOptionResults, err error) {
	path := api.url + "/datasets/" + dataset.ID + "/editions/" + dataset.Edition + "/versions/" + strconv.Itoa(dataset.Version) + "/dimensions/" + dimension + "/options"
	logData := log.Data{"func": "GetVersionDimensions", "URL": path, "dataset": dataset, "dimension": dimension}

	bytesResult, httpCode, err := api.get(ctx, path, nil)
	logData["httpCode"] = httpCode
	jsonResult := strings.TrimSpace(string(bytesResult))
	logData["jsonResult"] = jsonResult
	if err != nil {
		log.ErrorC("GetVersionDimensionOptions api get", err, logData)
		typ := "dimension"
		if jsonResult == versionNotFound {
			typ = "version"
		}
		return nil, handleError(httpCode, err, typ)
	}

	options = &models.DatasetDimensionOptionResults{}
	if err = json.Unmarshal(bytesResult, options); err != nil {
		log.ErrorC("GetVersionDimensionOptions unmarshal", err, logData)
		return
	}

	return
}

func (api *DatasetAPI) get(ctx context.Context, path string, vars url.Values) ([]byte, int, error) {
	return api.callDatasetAPI(ctx, "GET", path, vars)
}

// callDatasetAPI contacts the Dataset API returns the json body (action = PUT, GET, POST, ...)
func (api *DatasetAPI) callDatasetAPI(ctx context.Context, method, path string, payload interface{}) ([]byte, int, error) {
	logData := log.Data{"URL": path, "method": method}

	URL, err := url.Parse(path)
	if err != nil {
		log.ErrorC("failed to create url for dataset api call", err, logData)
		return nil, 0, err
	}
	path = URL.String()
	logData["URL"] = path

	var req *http.Request

	if payload != nil && method != "GET" {
		req, err = http.NewRequest(method, path, bytes.NewReader(payload.([]byte)))
		req.Header.Add("Content-type", "application/json")
		logData["payload"] = string(payload.([]byte))
	} else {
		req, err = http.NewRequest(method, path, nil)

		if payload != nil && method == "GET" {
			req.URL.RawQuery = payload.(url.Values).Encode()
			logData["payload"] = payload.(url.Values)
		}
	}

	// check req, above, didn't error
	if err != nil {
		log.ErrorC("failed to create request for dataset api", err, logData)
		return nil, 0, err
	}

	req.Header.Set(string(internalTokenKey), api.authToken)
	resp, err := api.client.Do(ctx, req)
	if err != nil {
		log.ErrorC("Failed to action dataset api", err, logData)
		return nil, 0, err
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			log.ErrorC("error cleaning up request body", err, logData)
		}
	}()
	logData["httpCode"] = resp.StatusCode

	jsonBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.ErrorC("failed to read body from dataset api", err, logData)
		return nil, resp.StatusCode, err
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= 300 {
		return jsonBody, resp.StatusCode, ErrUnexpectedStatusCode
	}

	return jsonBody, resp.StatusCode, nil
}

// GetHealthCheckClient returns a healthcheck-compatible client
func (api *DatasetAPI) GetHealthCheckClient() *dataset.Client {
	return dataset.New(api.url)
}

func handleError(httpCode int, err error, typ string) error {
	if err == ErrUnexpectedStatusCode {
		switch httpCode {
		case http.StatusNotFound:
			if typ == "dimension" {
				return ErrDimensionNotFound
			}
			if typ == "dimension option" {
				return ErrDimensionOptionsNotFound
			}
			return ErrVersionNotFound
		}
	}

	return err
}
