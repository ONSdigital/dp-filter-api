package preview

import (
	"bytes"
	"context"
	"encoding/csv"
	"io"

	"github.com/ONSdigital/dp-filter-api/models"
	"github.com/ONSdigital/dp-graph/v2/observation"
	"github.com/ONSdigital/log.go/log"
)

//go:generate moq -out previewtest/observationstore.go -pkg observationstoretest . ObservationStore

// ObservationStore provides filtered observation data in CSV rows.
type ObservationStore interface {
	StreamCSVRows(ctx context.Context, instanceID, filterID string, filters *observation.DimensionFilters, limit *int) (observation.StreamRowReader, error)
}

// DatasetStore used to query observations for previews
type DatasetStore struct {
	Store ObservationStore
}

// GetPreview generates a preview using the data stored in the graph database
func (preview *DatasetStore) GetPreview(ctx context.Context, bluePrint *models.Filter, limit int) (*models.FilterPreview, error) {
	filters := observation.DimensionFilters{}
	dimensions := []*observation.Dimension{}

	for _, dimension := range bluePrint.Dimensions {
		d := observation.Dimension{Name: dimension.Name, Options: dimension.Options}
		dimensions = append(dimensions, &d)
	}

	filters.Dimensions = dimensions

	previewLimit := limit
	rows, err := preview.Store.StreamCSVRows(ctx, bluePrint.InstanceID, bluePrint.FilterID, &filters, &previewLimit)
	if err != nil {
		return nil, err
	}

	log.Event(ctx, "reading rows into csv reader", log.INFO)
	csvReader, err := convertRowReaderToCSVReader(rows)
	if err != nil {
		return nil, err
	}

	results, err := buildResults(ctx, csvReader)
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

func buildResults(ctx context.Context, csvReader *csv.Reader) (*models.FilterPreview, error) {
	var results models.FilterPreview
	row, err := csvReader.Read()
	if err != nil {
		return nil, err
	}
	headers := row
	results.Headers = headers
	results.NumberOfColumns = len(headers)

	log.Event(ctx, "building preview results", log.INFO, log.Data{"headers": headers, "number_of_columns": results.NumberOfColumns})

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

	log.Event(ctx, "built preview results", log.INFO, log.Data{"number_of_rows": results.NumberOfRows})

	return &results, nil
}
