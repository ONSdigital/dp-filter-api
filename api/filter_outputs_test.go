package api

import (
	"testing"

	"github.com/ONSdigital/dp-filter-api/models"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	filterID1 = "121"
	filterID2 = "122"
	filterID3 = "123"
)

var testOptions = []struct {
	inputPreviousFilterOutput *models.Filter
	inputFilterOutput         *models.Filter
	expectedOutput            *models.Filter
	title                     string
}{
	{
		inputPreviousFilterOutput: &models.Filter{FilterID: filterID1, Downloads: nil},
		inputFilterOutput:         &models.Filter{Downloads: &fullDownloads},
		expectedOutput:            &models.Filter{Downloads: &fullDownloadsOutputs},
		title:                     "1",
	},
	{
		inputPreviousFilterOutput: &models.Filter{FilterID: filterID1, Downloads: &fullDownloads},
		inputFilterOutput:         &models.Filter{Downloads: nil},
		expectedOutput:            &models.Filter{Downloads: &fullDownloadsOutputs},
		title:                     "2",
	},
	{
		inputPreviousFilterOutput: &models.Filter{FilterID: filterID1, Downloads: nil},
		inputFilterOutput:         &models.Filter{Downloads: &csvDownloadsOnly},
		expectedOutput:            &models.Filter{Downloads: &csvDownloadsOnlyOutputs},
		title:                     "3",
	},
	{
		inputPreviousFilterOutput: &models.Filter{FilterID: filterID1, Downloads: nil},
		inputFilterOutput:         &models.Filter{Downloads: &xlsDownloadsOnly},
		expectedOutput:            &models.Filter{Downloads: &xlsDownloadsOnlyOutputs},
		title:                     "4",
	},
	{
		inputPreviousFilterOutput: &models.Filter{FilterID: filterID1, Downloads: &models.Downloads{CSV: &csvScenario[0].csv}},
		inputFilterOutput:         &models.Filter{Downloads: &models.Downloads{XLS: &xlsScenario[0].xls}},
		expectedOutput:            &models.Filter{Downloads: &fullDownloadsOutputs},
		title:                     "5",
	},
	{
		inputPreviousFilterOutput: &models.Filter{FilterID: filterID1, Downloads: &models.Downloads{XLS: &xlsScenario[0].xls}},
		inputFilterOutput:         &models.Filter{Downloads: &models.Downloads{CSV: &csvScenario[0].csv}},
		expectedOutput:            &models.Filter{Downloads: &fullDownloadsOutputs},
		title:                     "6",
	},
	{
		inputPreviousFilterOutput: &models.Filter{FilterID: filterID2, Downloads: &models.Downloads{CSV: &csvScenario[0].csv}},
		inputFilterOutput:         &models.Filter{Downloads: &models.Downloads{CSV: &csvScenario[1].csv}},
		expectedOutput:            &models.Filter{Downloads: &models.Downloads{CSV: &expectedDownloadItems[0].csv}},
		title:                     "7",
	},
	{
		inputPreviousFilterOutput: &models.Filter{FilterID: filterID3, Downloads: &models.Downloads{CSV: &csvScenario[0].csv}},
		inputFilterOutput:         &models.Filter{Downloads: &models.Downloads{CSV: &csvScenario[2].csv}},
		expectedOutput:            &models.Filter{Downloads: &models.Downloads{CSV: &expectedDownloadItems[1].csv}},
		title:                     "8",
	},
	{
		inputPreviousFilterOutput: &models.Filter{FilterID: filterID1, Downloads: &models.Downloads{CSV: &csvScenario[0].csv}},
		inputFilterOutput:         &models.Filter{Downloads: &models.Downloads{CSV: &csvScenario[3].csv}},
		expectedOutput:            &models.Filter{Downloads: &models.Downloads{CSV: &expectedDownloadItems[2].csv}},
		title:                     "9",
	},
	{
		inputPreviousFilterOutput: &models.Filter{FilterID: filterID2, Downloads: &models.Downloads{XLS: &xlsScenario[0].xls}},
		inputFilterOutput:         &models.Filter{Downloads: &models.Downloads{XLS: &xlsScenario[1].xls}},
		expectedOutput:            &models.Filter{Downloads: &models.Downloads{XLS: &expectedDownloadItems[3].xls}},
		title:                     "10",
	},
	{
		inputPreviousFilterOutput: &models.Filter{FilterID: filterID3, Downloads: &models.Downloads{XLS: &xlsScenario[0].xls}},
		inputFilterOutput:         &models.Filter{Downloads: &models.Downloads{XLS: &xlsScenario[2].xls}},
		expectedOutput:            &models.Filter{Downloads: &models.Downloads{XLS: &expectedDownloadItems[4].xls}},
		title:                     "11",
	},
	{
		inputPreviousFilterOutput: &models.Filter{FilterID: filterID1, Downloads: &models.Downloads{XLS: &xlsScenario[0].xls}},
		inputFilterOutput:         &models.Filter{Downloads: &models.Downloads{XLS: &xlsScenario[3].xls}},
		expectedOutput:            &models.Filter{Downloads: &models.Downloads{XLS: &expectedDownloadItems[5].xls}},
		title:                     "12",
	},
	{
		inputPreviousFilterOutput: &models.Filter{FilterID: filterID1, Downloads: &models.Downloads{CSV: &csvScenario[0].csv, XLS: &xlsScenario[0].xls}},
		inputFilterOutput:         &models.Filter{Downloads: &models.Downloads{CSV: &csvScenario[3].csv, XLS: &xlsScenario[3].xls}},
		expectedOutput:            &models.Filter{Downloads: &models.Downloads{CSV: &expectedDownloadItems[2].csv, XLS: &expectedDownloadItems[5].xls}},
		title:                     "13",
	},
}

func TestBuildDownloadsObject(t *testing.T) {

	Convey("Successfully build download object", t, func() {

		for _, option := range testOptions {
			Convey(option.title, func() {
				filterOutput := buildDownloadsObject(option.inputPreviousFilterOutput, option.inputFilterOutput, downloadServiceURL)
				So(filterOutput, ShouldResemble, option.expectedOutput)
			})
		}
	})
}

// Test data
var (
	fullDownloads = models.Downloads{
		CSV: &models.DownloadItem{
			Private: "csv-private-downloads-link",
			Public:  "csv-public-downloads-link",
			Size:    "12mb",
		},
		XLS: &models.DownloadItem{
			Private: "xls-private-downloads-link",
			Public:  "xls-public-downloads-link",
			Size:    "24mb",
		},
	}

	csvDownloadsOnly = models.Downloads{
		CSV: &models.DownloadItem{
			Private: "csv-private-downloads-link",
			Public:  "csv-public-downloads-link",
			Size:    "12mb",
		},
		XLS: nil,
	}

	xlsDownloadsOnly = models.Downloads{
		XLS: &models.DownloadItem{
			Private: "xls-private-downloads-link",
			Public:  "xls-public-downloads-link",
			Size:    "24mb",
		},
		CSV: nil,
	}

	fullDownloadsOutputs = models.Downloads{
		CSV: &models.DownloadItem{
			HRef:    downloadServiceURL + "/downloads/filter-outputs/" + filterID1 + ".csv",
			Private: "csv-private-downloads-link",
			Public:  "csv-public-downloads-link",
			Size:    "12mb",
		},
		XLS: &models.DownloadItem{
			HRef:    downloadServiceURL + "/downloads/filter-outputs/" + filterID1 + ".xlsx",
			Private: "xls-private-downloads-link",
			Public:  "xls-public-downloads-link",
			Size:    "24mb",
		},
	}

	csvDownloadsOnlyOutputs = models.Downloads{
		CSV: &models.DownloadItem{
			HRef:    downloadServiceURL + "/downloads/filter-outputs/" + filterID1 + ".csv",
			Private: "csv-private-downloads-link",
			Public:  "csv-public-downloads-link",
			Size:    "12mb",
		},
		XLS: nil,
	}

	xlsDownloadsOnlyOutputs = models.Downloads{
		XLS: &models.DownloadItem{
			HRef:    downloadServiceURL + "/downloads/filter-outputs/" + filterID1 + ".xlsx",
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
			Private: "xls-private-downloads-link",
			Public:  "xls-public-downloads-link",
			Size:    "24mb",
		},
	},
	{
		xls: models.DownloadItem{
			Private: "xls-private-downloads-link-2",
			Size:    "34mb",
		},
	},
	{
		xls: models.DownloadItem{
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
			HRef:    downloadServiceURL + "/downloads/filter-outputs/" + filterID1 + ".csv",
			Private: "csv-private-downloads-link",
			Public:  "csv-public-downloads-link",
			Size:    "12mb",
		},
	},
	{
		csv: models.DownloadItem{
			Private: "csv-private-downloads-link-2",
			Size:    "24mb",
		},
	},
	{
		csv: models.DownloadItem{
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
			HRef:    downloadServiceURL + "/downloads/filter-outputs/" + filterID2 + ".csv",
			Private: "csv-private-downloads-link-2",
			Public:  "csv-public-downloads-link",
			Size:    "24mb",
		},
	},
	{
		csv: models.DownloadItem{
			HRef:    downloadServiceURL + "/downloads/filter-outputs/" + filterID3 + ".csv",
			Private: "csv-private-downloads-link",
			Public:  "csv-public-downloads-link-3",
			Size:    "34mb",
		},
	},
	{
		csv: models.DownloadItem{
			HRef:    downloadServiceURL + "/downloads/filter-outputs/" + filterID1 + ".csv",
			Private: "csv-private-downloads-link",
			Public:  "csv-public-downloads-link-4",
			Size:    "12mb",
		},
	},
	{
		xls: models.DownloadItem{
			HRef:    downloadServiceURL + "/downloads/filter-outputs/" + filterID2 + ".xlsx",
			Private: "xls-private-downloads-link-2",
			Public:  "xls-public-downloads-link",
			Size:    "34mb",
		},
	},
	{
		xls: models.DownloadItem{
			HRef:    downloadServiceURL + "/downloads/filter-outputs/" + filterID3 + ".xlsx",
			Private: "xls-private-downloads-link",
			Public:  "xls-public-downloads-link-3",
			Size:    "44mb",
		},
	},
	{
		xls: models.DownloadItem{
			HRef:    downloadServiceURL + "/downloads/filter-outputs/" + filterID1 + ".xlsx",
			Private: "xls-private-downloads-link",
			Public:  "xls-public-downloads-link-4",
			Size:    "24mb",
		},
	},
}
