package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/ONSdigital/log.go/v2/log"

	"github.com/pkg/errors"
)

const (
	cantabularFlexibleTable = "cantabular_flexible_table"
)

type Assert struct{
	DatasetAPI    datasetAPIClient
	FilterFlexAPI filterFlexAPIClient
	svcAuthToken  string
	enabled       bool
}

func NewAssert(d datasetAPIClient, f filterFlexAPIClient, t string, e bool) *Assert{
	return &Assert{
		svcAuthToken:  t,
		DatasetAPI:    d,
		FilterFlexAPI: f,
		enabled:       e,
	}
}

func (a *Assert) errorResponse(ctx context.Context, w http.ResponseWriter, err error){
	log.Info(ctx, "error responding to HTTP request", log.ERROR, &log.EventErrors{{
			Message:    err.Error(),
			StackTrace: stackTrace(err),
			Data:       unwrapLogData(err),
		}},
	)

	status := unwrapStatusCode(err)
	msg    := errorMessage(err)
	resp   := erResponse{
		Errors: []string{msg},
	}

	logData := log.Data{
		"error":       err.Error(),
		"response":    msg,
		"status_code": status,
	}

	b, err := json.Marshal(resp)
	if err != nil {
		log.Error(ctx, "badly formed error response", err, logData)
		http.Error(w, msg, status)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	if _, err := w.Write(b); err != nil {
		log.Error(ctx, "failed to write error response", err, logData)
		return
	}
}

func (a *Assert) DatasetType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
		if !a.enabled{
			next.ServeHTTP(w, r)
			return
		}

		ctx := r.Context()

		buf := bytes.NewBuffer(make([]byte, 0))
		rdr := io.TeeReader(r.Body, buf)

		// Accept any request with 'dataset.id' field
		var req struct{
			Dataset struct{
				ID string `json: "id"`
			} `json: "dataset"`
		}

		if err := json.NewDecoder(rdr).Decode(&req); err != nil{
			a.errorResponse(ctx, w, er{
				err:    errors.Wrap(err, "failed to decode json"),
				msg:    fmt.Sprintf("badly formed request: %s", err),
				status: http.StatusBadRequest,
			})
			return
		}

		// TODO: Probably better to create a new GetDatasetType function in dataset api-client
		d, err := a.DatasetAPI.Get(ctx, "", a.svcAuthToken, "", req.Dataset.ID)
		if err != nil{
			a.errorResponse(ctx, w, er{
				err:    errors.Wrap(err, "failed to get dataset"),
				msg:    fmt.Sprintf("failed to get dataset"),
				status: statusCode(err),
			})
			return
		}

		r.Body = io.NopCloser(buf)

		if d.Type == cantabularFlexibleTable {
			a.errorResponse(ctx, w, er{
				err:    errors.Wrap(err, "failed to do proxy request"),
				msg:    fmt.Sprintf("failed to get dataset"),
				status: statusCode(err),
			})
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (a *Assert) doProxyRequest(w http.ResponseWriter, req *http.Request) error {
	resp, err := a.FilterFlexAPI.ForwardRequest(req)
	if err != nil{
		return errors.Wrap(err, "failed to forward request")
	}

	defer func(){
		if resp.Body != nil{
			resp.Body.Close()
		}
	}()

	b, err := io.ReadAll(resp.Body)
	if err != nil{
		return errors.Wrap(err, "failed to read response body")
	}

	for k, v := range resp.Header{
		for _, h := range v{
			w.Header().Add(k, h)
		}
	}

	w.WriteHeader(resp.StatusCode)
	if _, err := w.Write(b); err != nil{
		return er{
			err:     errors.Wrap(err, "failed to write response"),
			msg:     "unexpected error handling request",
			data:    log.Data{
				"response_body": string(b),
			},
		}
	}

	return nil
}
