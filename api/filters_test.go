package api

import (
	"testing"

	"github.com/ONSdigital/dp-filter-api/models"
	. "github.com/smartystreets/goconvey/convey"
)

var testOptions = []struct {
	inputPreviousFilterOutput *models.Filter
	inputFilterOutput         *models.Filter
	expectedOutput            *models.Filter
}{
	{
		inputPreviousFilterOutput: &models.Filter{Downloads: nil},
		inputFilterOutput:         &models.Filter{Downloads: &fullDownloads},
		expectedOutput:            &models.Filter{Downloads: &fullDownloads},
	},
	{
		inputPreviousFilterOutput: &models.Filter{Downloads: &fullDownloads},
		inputFilterOutput:         &models.Filter{Downloads: nil},
		expectedOutput:            &models.Filter{Downloads: &fullDownloads},
	},
	{
		inputPreviousFilterOutput: &models.Filter{Downloads: nil},
		inputFilterOutput:         &models.Filter{Downloads: &csvDownloadsOnly},
		expectedOutput:            &models.Filter{Downloads: &csvDownloadsOnly},
	},
	{
		inputPreviousFilterOutput: &models.Filter{Downloads: nil},
		inputFilterOutput:         &models.Filter{Downloads: &xlsDownloadsOnly},
		expectedOutput:            &models.Filter{Downloads: &xlsDownloadsOnly},
	},
	{
		inputPreviousFilterOutput: &models.Filter{Downloads: &models.Downloads{CSV: &csvScenario[0].csv}},
		inputFilterOutput:         &models.Filter{Downloads: &models.Downloads{XLS: &xlsScenario[0].xls}},
		expectedOutput:            &models.Filter{Downloads: &fullDownloads},
	},
	{
		inputPreviousFilterOutput: &models.Filter{Downloads: &models.Downloads{XLS: &xlsScenario[0].xls}},
		inputFilterOutput:         &models.Filter{Downloads: &models.Downloads{CSV: &csvScenario[0].csv}},
		expectedOutput:            &models.Filter{Downloads: &fullDownloads},
	},
	{
		inputPreviousFilterOutput: &models.Filter{Downloads: &models.Downloads{CSV: &csvScenario[0].csv}},
		inputFilterOutput:         &models.Filter{Downloads: &models.Downloads{CSV: &csvScenario[1].csv}},
		expectedOutput:            &models.Filter{Downloads: &models.Downloads{CSV: &expectedDownloadItems[0].csv}},
	},
	{
		inputPreviousFilterOutput: &models.Filter{Downloads: &models.Downloads{CSV: &csvScenario[0].csv}},
		inputFilterOutput:         &models.Filter{Downloads: &models.Downloads{CSV: &csvScenario[2].csv}},
		expectedOutput:            &models.Filter{Downloads: &models.Downloads{CSV: &expectedDownloadItems[1].csv}},
	},
	{
		inputPreviousFilterOutput: &models.Filter{Downloads: &models.Downloads{CSV: &csvScenario[0].csv}},
		inputFilterOutput:         &models.Filter{Downloads: &models.Downloads{CSV: &csvScenario[3].csv}},
		expectedOutput:            &models.Filter{Downloads: &models.Downloads{CSV: &expectedDownloadItems[2].csv}},
	},
	{
		inputPreviousFilterOutput: &models.Filter{Downloads: &models.Downloads{XLS: &xlsScenario[0].xls}},
		inputFilterOutput:         &models.Filter{Downloads: &models.Downloads{XLS: &xlsScenario[1].xls}},
		expectedOutput:            &models.Filter{Downloads: &models.Downloads{XLS: &expectedDownloadItems[3].xls}},
	},
	{
		inputPreviousFilterOutput: &models.Filter{Downloads: &models.Downloads{XLS: &xlsScenario[0].xls}},
		inputFilterOutput:         &models.Filter{Downloads: &models.Downloads{XLS: &xlsScenario[2].xls}},
		expectedOutput:            &models.Filter{Downloads: &models.Downloads{XLS: &expectedDownloadItems[4].xls}},
	},
	{
		inputPreviousFilterOutput: &models.Filter{Downloads: &models.Downloads{XLS: &xlsScenario[0].xls}},
		inputFilterOutput:         &models.Filter{Downloads: &models.Downloads{XLS: &xlsScenario[3].xls}},
		expectedOutput:            &models.Filter{Downloads: &models.Downloads{XLS: &expectedDownloadItems[5].xls}},
	},
	{
		inputPreviousFilterOutput: &models.Filter{Downloads: &models.Downloads{CSV: &csvScenario[0].csv, XLS: &xlsScenario[0].xls}},
		inputFilterOutput:         &models.Filter{Downloads: &models.Downloads{CSV: &csvScenario[3].csv, XLS: &xlsScenario[3].xls}},
		expectedOutput:            &models.Filter{Downloads: &models.Downloads{CSV: &expectedDownloadItems[2].csv, XLS: &expectedDownloadItems[5].xls}},
	},
}

func TestBuildDownloadsObject(t *testing.T) {

	Convey("Successfully build download object", t, func() {

		for _, option := range testOptions {
			filterOutput := buildDownloadsObject(option.inputPreviousFilterOutput, option.inputFilterOutput)
			So(filterOutput, ShouldResemble, option.expectedOutput)
		}
	})
}

// Test data
var (
	fullDownloads models.Downloads = models.Downloads{
		CSV: &models.DownloadItem{
			HRef:    "csv-downloads-link",
			Private: "csv-private-downloads-link",
			Public:  "csv-public-downloads-link",
			Size:    "12mb",
		},
		XLS: &models.DownloadItem{
			HRef:    "xls-downloads-link",
			Private: "xls-private-downloads-link",
			Public:  "xls-public-downloads-link",
			Size:    "24mb",
		},
	}

	csvDownloadsOnly models.Downloads = models.Downloads{
		CSV: &models.DownloadItem{
			HRef:    "csv-downloads-link",
			Private: "csv-private-downloads-link",
			Public:  "csv-public-downloads-link",
			Size:    "12mb",
		},
		XLS: nil,
	}

	xlsDownloadsOnly models.Downloads = models.Downloads{
		XLS: &models.DownloadItem{
			HRef:    "xls-downloads-link",
			Private: "xls-private-downloads-link",
			Public:  "xls-public-downloads-link",
			Size:    "24mb",
		},
		CSV: nil,
	}
)

var xlsScenario = []struct {
	xls models.DownloadItem
}{
	{
		xls: models.DownloadItem{
			HRef:    "xls-downloads-link",
			Private: "xls-private-downloads-link",
			Public:  "xls-public-downloads-link",
			Size:    "24mb",
		},
	},
	{
		xls: models.DownloadItem{
			HRef:    "xls-downloads-link-2",
			Private: "xls-private-downloads-link-2",
			Size:    "34mb",
		},
	},
	{
		xls: models.DownloadItem{
			HRef:   "xls-downloads-link-3",
			Public: "xls-public-downloads-link-3",
			Size:   "44mb",
		},
	},
	{
		xls: models.DownloadItem{
			Public: "xls-public-downloads-link-4",
		},
	},
}

var csvScenario = []struct {
	csv models.DownloadItem
}{
	{
		csv: models.DownloadItem{
			HRef:    "csv-downloads-link",
			Private: "csv-private-downloads-link",
			Public:  "csv-public-downloads-link",
			Size:    "12mb",
		},
	},
	{
		csv: models.DownloadItem{
			HRef:    "csv-downloads-link-2",
			Private: "csv-private-downloads-link-2",
			Size:    "24mb",
		},
	},
	{
		csv: models.DownloadItem{
			HRef:   "csv-downloads-link-3",
			Public: "csv-public-downloads-link-3",
			Size:   "34mb",
		},
	},
	{
		csv: models.DownloadItem{
			Public: "csv-public-downloads-link-4",
		},
	},
}

var expectedDownloadItems = []struct {
	csv models.DownloadItem
	xls models.DownloadItem
}{
	{
		csv: models.DownloadItem{
			HRef:    "csv-downloads-link-2",
			Private: "csv-private-downloads-link-2",
			Public:  "csv-public-downloads-link",
			Size:    "24mb",
		},
	},
	{
		csv: models.DownloadItem{
			HRef:    "csv-downloads-link-3",
			Private: "csv-private-downloads-link",
			Public:  "csv-public-downloads-link-3",
			Size:    "34mb",
		},
	},
	{
		csv: models.DownloadItem{
			HRef:    "csv-downloads-link",
			Private: "csv-private-downloads-link",
			Public:  "csv-public-downloads-link-4",
			Size:    "12mb",
		},
	},
	{
		xls: models.DownloadItem{
			HRef:    "xls-downloads-link-2",
			Private: "xls-private-downloads-link-2",
			Public:  "xls-public-downloads-link",
			Size:    "34mb",
		},
	},
	{
		xls: models.DownloadItem{
			HRef:    "xls-downloads-link-3",
			Private: "xls-private-downloads-link",
			Public:  "xls-public-downloads-link-3",
			Size:    "44mb",
		},
	},
	{
		xls: models.DownloadItem{
			HRef:    "xls-downloads-link",
			Private: "xls-private-downloads-link",
			Public:  "xls-public-downloads-link-4",
			Size:    "24mb",
		},
	},
}
