package mongo

import (
	"context"
	"testing"

	"github.com/ONSdigital/dp-filter-api/models"
	mim "github.com/ONSdigital/dp-mongodb-in-memory"
	mongoDriver "github.com/ONSdigital/dp-mongodb/v3/mongodb"
	. "github.com/smartystreets/goconvey/convey"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestGetFilterOutput(t *testing.T) {
	ctx := context.Background()

	server, err := mim.Start(ctx, "5.0.2")
	if err != nil {
		t.Fatal(err)
	}
	defer server.Stop(ctx)

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(server.URI()))
	if err != nil {
		t.Fatal(err)
	}

	sf := createFilterStore(client)
	collection := client.Database("filters").Collection("filterOutputs")

	Convey("Given a filter output exists in the database", t, func() {
		f := createFilter()
		_, err = collection.InsertOne(ctx, f)
		if err != nil {
			t.Fatal(err)
		}

		Convey("When ONS CMD searches the filter output via 'filterID'", func() {

			output, err := sf.GetFilterOutput(ctx, "some-filter-id")
			if err != nil {
				t.Fatal(err)
			}
			Convey("Then it should return a Filter Output", func() {
				So(output.FilterID, ShouldEqual, f.FilterID)
				So(output.ID, ShouldEqual, f.ID)
			})
		})

		Convey("When ONS Cantabular searches the filter output via 'ID'", func() {

			output, err := sf.GetFilterOutput(ctx, "some-id")
			if err != nil {
				t.Fatal(err)
			}
			Convey("Then it should return a Filter Output", func() {
				So(output.FilterID, ShouldEqual, f.FilterID)
				So(output.ID, ShouldEqual, f.ID)
			})
		})

		Convey("When the service searches for a non-existing ID ", func() {

			output, err := sf.GetFilterOutput(ctx, "none-existing-id")
			Convey("Then it should not return a Filter Output", func() {
				So(err.Error(), ShouldEqual, "filter output not found")
				So(output, ShouldBeNil)
			})
		})

		Reset(func() {
			_, err = collection.DeleteMany(ctx, bson.D{{}})
			if err != nil {
				t.Fatal(err)
			}
		})
	})
}

func TestCreateUpdateFilterOutput(t *testing.T) {
	Convey("When a filter output is updated with a new CSV file", t, func() {
		filterOutput := models.Filter{
			Downloads: &models.Downloads{
				CSV: &models.DownloadItem{
					HRef: "http://dataset-bucket/123.csv",
					Size: "321",
				},
			},
		}
		data := createUpdateFilterOutput(&filterOutput)
		Convey("Then the returned bson object contains the latest changes", func() {
			downloads := data["$set"].(bson.M)["downloads"].(models.Downloads)
			So(downloads.CSV.HRef, ShouldEndWith, filterOutput.Downloads.CSV.HRef)
			So(downloads.CSV.Size, ShouldEndWith, filterOutput.Downloads.CSV.Size)
		})
	})

	Convey("When a filter output is updated with a new XLSX file", t, func() {
		filterOutput := models.Filter{
			Downloads: &models.Downloads{
				XLS: &models.DownloadItem{
					HRef: "http://dataset-bucket/123.xlsx",
					Size: "3213",
				},
			},
		}
		data := createUpdateFilterOutput(&filterOutput)
		Convey("Then the returned bson object contains the latest changes", func() {
			downloads := data["$set"].(bson.M)["downloads"].(models.Downloads)
			So(downloads.XLS.HRef, ShouldEndWith, filterOutput.Downloads.XLS.HRef)
			So(downloads.XLS.Size, ShouldEndWith, filterOutput.Downloads.XLS.Size)
		})
	})

	Convey("When a filter output is updated with both a XLSX and CSV file", t, func() {
		filterOutput := models.Filter{
			Downloads: &models.Downloads{
				XLS: &models.DownloadItem{
					HRef: "http://dataset-bucket/123.xlsx",
					Size: "3213",
				},
				CSV: &models.DownloadItem{
					HRef: "http://dataset-bucket/123.csv",
					Size: "321",
				},
			},
			State: models.CompletedState,
		}
		data := createUpdateFilterOutput(&filterOutput)
		Convey("Then the returned bson object contains the latest changes", func() {
			state := data["$set"].(bson.M)["state"].(string)
			downloads := data["$set"].(bson.M)["downloads"].(models.Downloads)
			So(downloads.XLS.HRef, ShouldEndWith, filterOutput.Downloads.XLS.HRef)
			So(downloads.XLS.Size, ShouldEndWith, filterOutput.Downloads.XLS.Size)
			So(downloads.CSV.HRef, ShouldEndWith, filterOutput.Downloads.CSV.HRef)
			So(downloads.CSV.Size, ShouldEndWith, filterOutput.Downloads.CSV.Size)
			So(state, ShouldEqual, models.CompletedState)

		})
	})
}

func TestSelector(t *testing.T) {

	Convey("Given some testing values to provide as selector parameters", t, func() {
		var testETag = "testETag"
		var testMongoTimestamp = primitive.Timestamp{1234567890, 0}

		Convey("Then, providing an empty string dimension, zero timestamp and any eTag generates a selector that only queries by filter_id", func() {
			s := selector(testFilterID, "", primitive.Timestamp{}, AnyETag)
			So(s, ShouldResemble, bson.M{"filter_id": testFilterID})
		})

		Convey("Then, providing values for dimension, timestamp, and eTag generates a selector that queries by filterID, dimensionName, timestamp and eTag", func() {
			s := selector(testFilterID, testDimensionName, testMongoTimestamp, testETag)
			So(s, ShouldResemble, bson.M{
				"filter_id":        testFilterID,
				"dimensions":       bson.M{"$elemMatch": bson.M{"name": testDimensionName}},
				"unique_timestamp": testMongoTimestamp,
				"e_tag":            testETag,
			})
		})
	})
}

func TestValidateFilter(t *testing.T) {
	Convey("validateFilter creates empty slices for dimensions and events if the values were nil", t, func() {
		filter := models.Filter{}
		validateFilter(&filter)
		So(filter.Dimensions, ShouldResemble, []models.Dimension{})
		So(filter.Events, ShouldResemble, []*models.Event{})
	})
}

func createFilter() *models.Filter {
	filter := &models.Filter{
		ID:         "some-id",
		InstanceID: "some-instance-id",
		Dimensions: nil,
		Downloads:  nil,
		Events:     nil,
		FilterID:   "some-filter-id",
		State:      "published",
		Links:      models.LinkMap{},
		Type:       "flexible",
	}
	return filter
}

func createFilterStore(client *mongo.Client) FilterStore {
	sf := FilterStore{
		MongoDriverConfig: mongoDriver.MongoDriverConfig{
			Username:                      "admin",
			Password:                      "admin",
			ClusterEndpoint:               "localhost:27017",
			Database:                      "filters",
			Collections:                   map[string]string{"FiltersCollection": "filters", "OutputsCollection": "filterOutputs"},
			ReplicaSet:                    "",
			IsStrongReadConcernEnabled:    false,
			IsWriteConcernMajorityEnabled: true,
			ConnectTimeout:                5,
			QueryTimeout:                  15,
			TLSConnectionConfig: mongoDriver.TLSConnectionConfig{
				IsSSL:              false,
				VerifyCert:         false,
				CACertChain:        "",
				RealHostnameForSSH: "",
			},
		},
		Connection:        mongoDriver.NewMongoConnection(client, "filters"),
		healthCheckClient: nil,
		URI:               "mongodb://localhost:27017",
	}
	return sf
}
