package dataset

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/ONSdigital/dp-filter-api/models"
	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/go-ns/rchttp"
)

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

var (
	ErrorUnexpectedStatusCode = errors.New("unexpected status code from api")
	ErrorInstanceNotFound     = errors.New("Instance not found")

	publishedState = "published"
)

// GetInstance queries the Dataset API to get an instance
func (api *DatasetAPI) GetInstance(ctx context.Context, instanceID string) (instance *models.Instance, err error) {
	path := api.url + "/instances/" + instanceID
	logData := log.Data{"func": "GetInstance", "URL": path, "instance_id": instanceID}

	jsonResult, httpCode, err := api.get(ctx, path, nil)
	logData["httpCode"] = httpCode
	logData["jsonResult"] = jsonResult
	if err != nil {
		log.ErrorC("api get", err, logData)
		return nil, handleError(httpCode, err)
	}

	instance = &models.Instance{}
	if err = json.Unmarshal(jsonResult, instance); err != nil {
		log.ErrorC("unmarshal", err, logData)
		return
	}

	// External facing customers should NOT be able to filter an unpublished instance
	// TODO If authorised (internal user or precanned results) should be able to filter
	// instances which have a state of edition-confirmed, associated and published
	if instance.State != publishedState {
		return instance, ErrorInstanceNotFound
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

	req.Header.Set("Internal-token", api.authToken)
	resp, err := api.client.Do(ctx, req)
	if err != nil {
		log.ErrorC("Failed to action dataset api", err, logData)
		return nil, 0, err
	}
	defer resp.Body.Close()

	logData["httpCode"] = resp.StatusCode
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= 300 {
		return nil, resp.StatusCode, ErrorUnexpectedStatusCode
	}

	jsonBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.ErrorC("failed to read body from dataset api", err, logData)
		return nil, resp.StatusCode, err
	}

	return jsonBody, resp.StatusCode, nil
}

func handleError(httpCode int, err error) error {
	if err == ErrorUnexpectedStatusCode {
		switch httpCode {
		case http.StatusNotFound:
			return ErrorInstanceNotFound
		}
	}

	return err
}
