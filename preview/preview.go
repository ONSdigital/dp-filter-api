package preview

import (
	"bytes"
	"context"
	"encoding/csv"
	"io"

	"github.com/ONSdigital/dp-filter-api/models"
	"github.com/ONSdigital/dp-graph/observation"
	"github.com/ONSdigital/go-ns/log"
)

//go:generate moq -out previewtest/observationstore.go -pkg observationstoretest . ObservationStore

// ObservationStore provides filtered observation data in CSV rows.
type ObservationStore interface {
	StreamCSVRows(ctx context.Context, filter *observation.Filter, limit *int) (observation.StreamRowReader, error)
}

// DatasetStore used to query observations for previews
type DatasetStore struct {
	Store ObservationStore
}

// FilterPreview contains the results of a requested preview
type FilterPreview struct {
	Headers         []string   `json:"headers"`
	NumberOfRows    int        `json:"number_of_rows"`
	NumberOfColumns int        `json:"number_of_columns"`
	Rows            [][]string `json:"rows"`
}

// GetPreview generates a preview using the data stored in the graph database
func (preview *DatasetStore) GetPreview(ctx context.Context, bluePrint *models.Filter, limit int) (*FilterPreview, error) {
	var filter = observation.Filter{}
	filter.InstanceID = bluePrint.InstanceID

	for _, dimension := range bluePrint.Dimensions {
		d := observation.DimensionFilter{Name: dimension.Name, Options: dimension.Options}
		filter.DimensionFilters = append(filter.DimensionFilters, &d)
	}

	previewLimit := limit
	rows, err := preview.Store.StreamCSVRows(ctx, &filter, &previewLimit)
	if err != nil {
		return nil, err
	}

	log.InfoCtx(ctx, "reading rows into csv reader", nil)
	csvReader, err := convertRowReaderToCSVReader(rows)
	if err != nil {
		return nil, err
	}

	results, err := buildResults(csvReader)
	rows.Close(ctx)
	return results, err
}

func convertRowReaderToCSVReader(rows observation.StreamRowReader) (*csv.Reader, error) {
	var buffer bytes.Buffer
	for {
		row, err := rows.Read()
		if err != nil {
			if err == io.EOF {
				break // We read the end of the stream
			}
			return nil, err
		}
		buffer.WriteString(row)
	}
	csvReader := csv.NewReader(bytes.NewReader(buffer.Bytes()))
	return csvReader, nil
}

func buildResults(csvReader *csv.Reader) (*FilterPreview, error) {
	var results FilterPreview
	row, err := csvReader.Read()
	if err != nil {
		return nil, err
	}
	headers := row
	results.Headers = headers
	results.NumberOfColumns = len(headers)

	log.Info("building preview results", log.Data{"headers": headers, "number_of_columns": results.NumberOfColumns})

	for {
		row, err = csvReader.Read()
		if err != nil {
			if err == io.EOF {
				break // We read the end of the stream
			}
			return nil, err
		}
		results.Rows = append(results.Rows, row)
		results.NumberOfRows++
	}

	log.Info("built preview results", log.Data{"number_of_rows": results.NumberOfRows})

	return &results, nil
}
