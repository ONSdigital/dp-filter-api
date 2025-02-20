package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/ONSdigital/dp-filter-api/filters"
	"github.com/ONSdigital/log.go/v2/log"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

const (
	cantabularTable             = "cantabular_table"
	cantabularFlexibleTable     = "cantabular_flexible_table"
	cantabularMultivariateTable = "cantabular_multivariate_table"
	flexible                    = "flexible"
	multivariate                = "multivariate"
	anyEtagSelector             = "*"
)

type Assert struct {
	respond       responder
	DatasetAPI    datasetAPIClient
	FilterFlexAPI filterFlexAPIClient
	store         datastore
	svcAuthToken  string
	enabled       bool
}

func NewAssert(r responder, d datasetAPIClient, f filterFlexAPIClient, ds datastore, t string, e bool) *Assert {
	return &Assert{
		svcAuthToken:  t,
		DatasetAPI:    d,
		FilterFlexAPI: f,
		store:         ds,
		enabled:       e,
		respond:       r,
	}
}

// FilterOutputType is a forwarder that checks for filter in the route, not in the body as below.
// Used for PUT filter-output/{filter-output-id}
func (a *Assert) FilterOutputType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !a.enabled {
			next.ServeHTTP(w, r)
			return
		}

		vars := mux.Vars(r)
		filterOutputID := vars["filter_output_id"]

		ctx := r.Context()
		filterOutput, err := a.store.GetFilterOutput(ctx, filterOutputID)
		if err != nil {
			a.respond.Error(ctx, w, filters.GetErrorStatusCode(err), er{
				err: errors.Wrap(err, "failed to get filter output"),
				msg: "failed to get filter output",
			})
			return
		}

		if filterOutput.Type == flexible || filterOutput.Type == multivariate {
			if err := a.doProxyRequest(w, r); err != nil {
				a.respond.Error(ctx, w, filters.GetErrorStatusCode(err), er{
					err: errors.Wrap(err, "failed to do proxy request"),
					msg: "unable to fulfil request",
				})
			}
			return
		}

		next.ServeHTTP(w, r)
	})
}
func (a *Assert) DatasetType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !a.enabled {
			next.ServeHTTP(w, r)
			return
		}

		ctx := r.Context()

		buf := bytes.NewBuffer(make([]byte, 0))
		rdr := io.TeeReader(r.Body, buf)

		// Accept any request with 'dataset.id' field
		var req struct {
			Dataset struct {
				ID string `json:"id"`
			} `json:"dataset"`
		}

		if err := json.NewDecoder(rdr).Decode(&req); err != nil {
			a.respond.Error(ctx, w, http.StatusBadRequest, er{
				err: errors.Wrap(err, "failed to decode json"),
				msg: fmt.Sprintf("badly formed request: %s", err),
			})
			return
		}

		// TODO: Probably better to create a new GetDatasetType function in dataset api-client
		d, err := a.DatasetAPI.Get(ctx, "", a.svcAuthToken, "", req.Dataset.ID)
		if err != nil {
			a.respond.Error(ctx, w, filters.GetErrorStatusCode(err), er{
				err: errors.Wrap(err, "failed to get dataset"),
				msg: "failed to get dataset",
			})
			return
		}

		r.Body = io.NopCloser(buf)

		if d.Type == cantabularFlexibleTable || d.Type == cantabularMultivariateTable {
			if err := a.doProxyRequest(w, r); err != nil {
				a.respond.Error(ctx, w, filters.GetErrorStatusCode(err), er{
					err: errors.Wrap(err, "failed to do proxy request"),
					msg: "unable to fulfil request",
				})
			}
			return
		} else if d.Type == cantabularTable {
			a.respond.Error(ctx, w, http.StatusBadRequest, errors.New("invalid dataset type"))
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (a *Assert) FilterType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !a.enabled {
			next.ServeHTTP(w, r)
			return
		}

		ctx := r.Context()

		vars := mux.Vars(r)
		filterID := vars["filter_blueprint_id"]

		// TODO: Better to add GetFilterType query to mongo?
		f, err := a.store.GetFilter(ctx, filterID, anyEtagSelector)
		if err != nil {
			a.respond.Error(ctx, w, filters.GetErrorStatusCode(err), er{
				err: errors.Wrap(err, "failed to get filter"),
				msg: "failed to get filter",
			})
			return
		}

		if f.Type == flexible || f.Type == multivariate {
			log.Info(ctx, "checking X-Forwarded-Host and path prefix in FilterType()", log.Data{"X-Forwarded-Host": r.Header.Get("X-Forwarded-Host"),
				"X-Forwarded-Path-Prefix": r.Header.Get("X-Forwarded-Path-Prefix")})
			if err := a.doProxyRequest(w, r); err != nil {
				a.respond.Error(ctx, w, filters.GetErrorStatusCode(err), er{
					err: errors.Wrap(err, "failed to do proxy request"),
					msg: "unable to fulfil request",
				})
			}
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (a *Assert) doProxyRequest(w http.ResponseWriter, req *http.Request) error {
	log.Info(context.Background(), "checking X-Forwarded-Host and path prefix in doProxyRequest()", log.Data{"X-Forwarded-Host": req.Header.Get("X-Forwarded-Host"),
		"X-Forwarded-Path-Prefix": req.Header.Get("X-Forwarded-Path-Prefix")})
	req.Header.Add("Test-Host", req.Header.Get("X-Forwarded-Host"))
	resp, err := a.FilterFlexAPI.ForwardRequest(req)
	if err != nil {
		return errors.Wrap(err, "failed to forward request")
	}

	defer func() {
		if resp.Body != nil {
			resp.Body.Close()
		}
	}()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "failed to read response body")
	}

	for k, v := range resp.Header {
		for _, h := range v {
			w.Header().Add(k, h)
		}
	}

	w.WriteHeader(resp.StatusCode)
	if _, err := w.Write(b); err != nil {
		return er{
			err: errors.Wrap(err, "failed to write response"),
			msg: "unexpected error handling request",
			data: log.Data{
				"response_body": string(b),
			},
		}
	}

	return nil
}
