package api_test

import (
	"testing"

	"github.com/ONSdigital/dp-filter-api/api"
	"github.com/ONSdigital/dp-filter-api/models"
	. "github.com/smartystreets/goconvey/convey"
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
		title:                     "no previous downloads, providing full downloads",
	},
	{
		inputPreviousFilterOutput: &models.Filter{FilterID: filterID1, Downloads: &fullDownloads},
		inputFilterOutput:         &models.Filter{Downloads: nil},
		expectedOutput:            &models.Filter{Downloads: &fullDownloadsOutputs},
		title:                     "existing downloads, providing empty update",
	},
	{
		inputPreviousFilterOutput: &models.Filter{FilterID: filterID1, Downloads: nil},
		inputFilterOutput:         &models.Filter{Downloads: &csvDownloadsOnly},
		expectedOutput:            &models.Filter{Downloads: &models.Downloads{CSV: csvFullDownload}},
		title:                     "no previous downloads, providing only csv",
	},
	{
		inputPreviousFilterOutput: &models.Filter{FilterID: filterID1, Downloads: nil},
		inputFilterOutput:         &models.Filter{Downloads: &xlsDownloadsOnly},
		expectedOutput:            &models.Filter{Downloads: &xlsDownloadsOnlyOutputs},
		title:                     "no previous downloads, providing only xls",
	},
	{
		inputPreviousFilterOutput: &models.Filter{FilterID: filterID1, Downloads: &models.Downloads{CSV: &csvScenario[0].csv}},
		inputFilterOutput:         &models.Filter{Downloads: &models.Downloads{XLS: &xlsScenario[0].xls}},
		expectedOutput:            &models.Filter{Downloads: &fullDownloadsOutputs},
		title:                     "existing csv download, providing xls",
	},
	{
		inputPreviousFilterOutput: &models.Filter{FilterID: filterID1, Downloads: &models.Downloads{XLS: &xlsScenario[0].xls}},
		inputFilterOutput:         &models.Filter{Downloads: &models.Downloads{CSV: &csvScenario[0].csv}},
		expectedOutput:            &models.Filter{Downloads: &fullDownloadsOutputs},
		title:                     "existing xls download, providing csv",
	},
	{
		inputPreviousFilterOutput: &models.Filter{FilterID: filterID2, Downloads: &models.Downloads{CSV: &csvScenario[0].csv}},
		inputFilterOutput:         &models.Filter{Downloads: &models.Downloads{CSV: &csvScenario[1].csv}},
		expectedOutput:            &models.Filter{Downloads: &models.Downloads{CSV: &expectedDownloadItems[0].csv}},
		title:                     "existing csv download, providing private csv link",
	},
	{
		inputPreviousFilterOutput: &models.Filter{FilterID: filterID3, Downloads: &models.Downloads{CSV: &csvScenario[0].csv}},
		inputFilterOutput:         &models.Filter{Downloads: &models.Downloads{CSV: &csvScenario[2].csv}},
		expectedOutput:            &models.Filter{Downloads: &models.Downloads{CSV: &expectedDownloadItems[1].csv}},
		title:                     "existing csv download, provding public csv link",
	},
	{
		inputPreviousFilterOutput: &models.Filter{FilterID: filterID1, Downloads: &models.Downloads{CSV: &csvScenario[0].csv}},
		inputFilterOutput:         &models.Filter{Downloads: &models.Downloads{CSV: &csvScenario[3].csv}},
		expectedOutput:            &models.Filter{Downloads: &models.Downloads{CSV: &expectedDownloadItems[2].csv}},
		title:                     "existing csv download, providing public xls link but no size",
	},
	{
		inputPreviousFilterOutput: &models.Filter{FilterID: filterID2, Downloads: &models.Downloads{XLS: &xlsScenario[0].xls}},
		inputFilterOutput:         &models.Filter{Downloads: &models.Downloads{XLS: &xlsScenario[1].xls}},
		expectedOutput:            &models.Filter{Downloads: &models.Downloads{XLS: &expectedDownloadItems[3].xls}},
		title:                     "existing xls download, providing private xls link",
	},
	{
		inputPreviousFilterOutput: &models.Filter{FilterID: filterID3, Downloads: &models.Downloads{XLS: &xlsScenario[0].xls}},
		inputFilterOutput:         &models.Filter{Downloads: &models.Downloads{XLS: &xlsScenario[2].xls}},
		expectedOutput:            &models.Filter{Downloads: &models.Downloads{XLS: &expectedDownloadItems[4].xls}},
		title:                     "existing xls download, providing public xls link",
	},
	{
		inputPreviousFilterOutput: &models.Filter{FilterID: filterID1, Downloads: &models.Downloads{XLS: &xlsScenario[0].xls}},
		inputFilterOutput:         &models.Filter{Downloads: &models.Downloads{XLS: &xlsScenario[3].xls}},
		expectedOutput:            &models.Filter{Downloads: &models.Downloads{XLS: &expectedDownloadItems[5].xls}},
		title:                     "existing xls download, providing public xls link but no size",
	},
	{
		inputPreviousFilterOutput: &models.Filter{FilterID: filterID1, Downloads: &models.Downloads{CSV: &csvScenario[0].csv, XLS: &xlsScenario[0].xls}},
		inputFilterOutput:         &models.Filter{Downloads: &models.Downloads{CSV: &csvScenario[3].csv, XLS: &xlsScenario[3].xls}},
		expectedOutput:            &models.Filter{Downloads: &models.Downloads{CSV: &expectedDownloadItems[2].csv, XLS: &expectedDownloadItems[5].xls}},
		title:                     "existing downloads, providing update links for both but no sizes",
	},
	{
		inputPreviousFilterOutput: &models.Filter{FilterID: filterID1, Downloads: &models.Downloads{CSV: &csvScenario[0].csv}},
		inputFilterOutput:         &models.Filter{Downloads: &models.Downloads{XLS: &xlsScenario[4].xls}},
		expectedOutput:            &models.Filter{Downloads: &models.Downloads{CSV: &csvScenario[0].csv, XLS: &expectedDownloadItems[6].xls}},
		title:                     "existing csv download, skipped xls generation",
	},
}

func TestBuildDownloadsObject(t *testing.T) {
	Convey("Successfully build download object", t, func() {
		for _, option := range testOptions {
			Convey(option.title, func(c C) {
				api.BuildDownloadsObject(option.inputPreviousFilterOutput, option.inputFilterOutput, downloadServiceURL)
				So(option.inputFilterOutput, ShouldResemble, option.expectedOutput)
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

	csvFullDownload = &models.DownloadItem{
		HRef:    downloadServiceURL + "/downloads/filter-outputs/" + filterID1 + ".csv",
		Private: "csv-private-downloads-link",
		Public:  "csv-public-downloads-link",
		Size:    "12mb",
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
	{
		xls: models.DownloadItem{
			Skipped: true,
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
	{
		csv: models.DownloadItem{
			Skipped: true,
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
	{
		xls: models.DownloadItem{
			Skipped: true,
		},
	},
}
