package middleware

import (
	"net/http"
	"bytes"
	"encoding/json"
	"io"
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
			http.Error(w, "failed to decode json: " + err.Error(), http.StatusBadRequest)
			return
		}

		// TODO: Probably better to create a new GetDatasetType function in dataset api-client
		d, err := a.DatasetAPI.Get(ctx, "", a.svcAuthToken, "", req.Dataset.ID)
		if err != nil{
			http.Error(w, "failed to get dataset: " + err.Error(), http.StatusInternalServerError)
			return
		}

		r.Body = io.NopCloser(buf)

		if d.Type == cantabularFlexibleTable {
			a.doProxyRequest(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (a *Assert) doProxyRequest(w http.ResponseWriter, req *http.Request){
	resp, err := a.FilterFlexAPI.ForwardRequest(req)
	if err != nil{
		http.Error(w, "failed to forward request", http.StatusInternalServerError)
		return
	}

	defer func(){
		if resp.Body != nil{
			resp.Body.Close()
		}
	}()

	b, err := io.ReadAll(resp.Body)
	if err != nil{
		http.Error(w, "failed to forward request", http.StatusInternalServerError)
		return
	}

	for k, v := range resp.Header{
		for _, h := range v{
			w.Header().Add(k, h)
		}
	}

	w.WriteHeader(resp.StatusCode)
	if _, err := w.Write(b); err != nil{
		panic(err)
	}
}
