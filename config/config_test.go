package config

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSpec(t *testing.T) {

	Convey("Given an environment with no environment variables set", t, func() {
		cfg, err := Get()

		Convey("When the config values are retrieved", func() {

			Convey("There should be no error returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("The values should be set to the expected defaults", func() {
				So(cfg.BindAddr, ShouldEqual, ":22100")
				So(cfg.Host, ShouldEqual, "http://localhost:22100")
				So(cfg.Brokers, ShouldResemble, []string{"localhost:9092"})
				So(cfg.FilterOutputSubmittedTopic, ShouldEqual, "filter-job-submitted")
				So(cfg.KafkaMaxBytes, ShouldEqual, "2000000")
				So(cfg.ShutdownTimeout, ShouldEqual, 5*time.Second)
				So(cfg.MongoConfig.BindAddr, ShouldEqual, "localhost:27017")
				So(cfg.MongoConfig.Database, ShouldEqual, "filters")
				So(cfg.MongoConfig.FiltersCollection, ShouldEqual, "filters")
				So(cfg.MongoConfig.OutputsCollection, ShouldEqual, "filterOutputs")
				So(cfg.ServiceAuthToken, ShouldEqual, "FD0108EA-825D-411C-9B1D-41EF7727F465")
				So(cfg.ZebedeeURL, ShouldEqual, "http://localhost:8082")
				So(cfg.EnablePrivateEndpoints, ShouldEqual, true)
				So(cfg.DownloadServiceURL, ShouldEqual, "http://localhost:23600")
				So(cfg.DownloadServiceSecretKey, ShouldEqual, "QB0108EZ-825D-412C-9B1D-41EF7747F462")
				So(cfg.AuditEventsTopic, ShouldEqual, "audit-events")
				So(cfg.HealthCheckInterval, ShouldEqual, 30*time.Second)
				So(cfg.HealthCheckCriticalTimeout, ShouldEqual, 90*time.Second)
			})
		})
	})
}
