package config

import (
	"os"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSpec(t *testing.T) {

	Convey("Given an environment with no environment variables set", t, func() {
		os.Clearenv()
		cfg, err := Get()

		Convey("When the config values are retrieved", func() {

			Convey("There should be no error returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("The values should be set to the expected defaults", func() {
				So(cfg.BindAddr, ShouldEqual, ":22100")
				So(cfg.Host, ShouldEqual, "http://localhost:22100")
				So(cfg.MaxRequestOptions, ShouldEqual, 1000)
				So(cfg.Brokers, ShouldResemble, []string{"localhost:9092", "localhost:9093", "localhost:9094"})
				So(cfg.KafkaVersion, ShouldEqual, "1.0.2")
				So(cfg.KafkaSecProtocol, ShouldEqual, "")
				So(cfg.FilterOutputSubmittedTopic, ShouldEqual, "filter-job-submitted")
				So(cfg.KafkaMaxBytes, ShouldEqual, 2000000)
				So(cfg.ShutdownTimeout, ShouldEqual, 5*time.Second)
				So(cfg.MongoConfig.ClusterEndpoint, ShouldEqual, "localhost:27017")
				So(cfg.MongoConfig.Database, ShouldEqual, "filters")
				So(cfg.MongoConfig.Collections, ShouldResemble, map[string]string{FiltersCollection: "filters", OutputsCollection: "filterOutputs"})
				So(cfg.MongoConfig.IsStrongReadConcernEnabled, ShouldEqual, false)
				So(cfg.MongoConfig.IsWriteConcernMajorityEnabled, ShouldEqual, true)
				So(cfg.MongoConfig.ConnectTimeout, ShouldEqual, 5*time.Second)
				So(cfg.MongoConfig.QueryTimeout, ShouldEqual, 15*time.Second)
				So(cfg.MongoConfig.IsSSL, ShouldEqual, false)
				So(cfg.MongoConfig.Limit, ShouldEqual, 20)
				So(cfg.MongoConfig.Offset, ShouldEqual, 0)
				So(cfg.ServiceAuthToken, ShouldEqual, "FD0108EA-825D-411C-9B1D-41EF7727F465")
				So(cfg.ZebedeeURL, ShouldEqual, "http://localhost:8082")
				So(cfg.EnablePrivateEndpoints, ShouldEqual, false)
				So(cfg.DownloadServiceURL, ShouldEqual, "http://localhost:23600")
				So(cfg.DownloadServiceSecretKey, ShouldEqual, "QB0108EZ-825D-412C-9B1D-41EF7747F462")
				So(cfg.HealthCheckInterval, ShouldEqual, 30*time.Second)
				So(cfg.HealthCheckCriticalTimeout, ShouldEqual, 90*time.Second)
				So(cfg.MaxDatasetOptions, ShouldEqual, 200)
				So(cfg.DefaultMaxLimit, ShouldEqual, 1000)
				So(cfg.FilterFlexAPIURL, ShouldEqual, "http://localhost:27100")
			})
		})
	})
}
