package preview

import (
	"bytes"
	"encoding/csv"
	"github.com/ONSdigital/dp-filter-api/models"
	"github.com/ONSdigital/dp-filter/observation"
	"io"
)

//go:generate moq -out previewtest/observationstore.go -pkg observationstoretest . ObservationStore

// ObservationStore used to get observations in a CSV format
type ObservationStore interface {
	GetCSVRows(*observation.Filter, *int) (observation.CSVRowReader, error)
}

// PreviewDatasetStore used to query observations for previews
type PreviewDatasetStore struct {
	Store ObservationStore
	Limit int
}

// FilterPreview contains the results of a requested preview
type FilterPreview struct {
	Headers         []string   `json:"headers"`
	NumberOfRows    int        `json:"number_of_rows"`
	NumberOfColumns int        `json:"number_of_columns"`
	Rows            [][]string `json:"rows"`
}

// GetPreview generates a preview using the data stored in the graph database
func (preview *PreviewDatasetStore) GetPreview(bluePrint models.Filter) (*FilterPreview, error) {
	var filter = observation.Filter{}
	filter.InstanceID = bluePrint.InstanceID
	for _, dimension := range bluePrint.Dimensions {
		d := observation.DimensionFilter{Name: dimension.Name, Options: dimension.Options}
		filter.DimensionFilters = append(filter.DimensionFilters, &d)
	}
	rows, err := preview.Store.GetCSVRows(&filter, &preview.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	csvReader, err := convertRowReaderToCSVReader(rows)
	if err != nil {
		return nil, err
	}

	return buildResults(csvReader)
}

func convertRowReaderToCSVReader(rows observation.CSVRowReader) (*csv.Reader, error) {
	var buffer bytes.Buffer
	for true {
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
	// Replace the V4 header withe values
	results.Headers[0] = "Values"
	results.NumberOfColumns = len(headers)
	for true {
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
	return &results, nil
}
