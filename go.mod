module github.com/ONSdigital/dp-filter-api

go 1.15

require (
	github.com/ONSdigital/dp-api-clients-go v1.32.10
	github.com/ONSdigital/dp-graph/v2 v2.3.0
	github.com/ONSdigital/dp-healthcheck v1.0.5
	github.com/ONSdigital/dp-kafka/v2 v2.1.2
	github.com/ONSdigital/dp-mongodb v1.5.0
	github.com/ONSdigital/dp-net v1.0.10
	github.com/ONSdigital/go-ns v0.0.0-20200902154605-290c8b5ba5eb
	github.com/ONSdigital/log.go v1.0.1
	github.com/globalsign/mgo v0.0.0-20190517090918-73267e130ca1
	github.com/gopherjs/gopherjs v0.0.0-20200217142428-fce0ec30dd00 // indirect
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/justinas/alice v1.2.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/mattn/go-colorable v0.1.7 // indirect
	github.com/pkg/errors v0.9.1
	github.com/satori/go.uuid v1.2.1-0.20181028125025-b2ce2384e17b
	github.com/smartystreets/assertions v1.2.0 // indirect
	github.com/smartystreets/goconvey v1.6.4
	go.mongodb.org/mongo-driver v1.5.2
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	gopkg.in/avro.v0 v0.0.0-20171217001914-a730b5802183 // indirect
	gopkg.in/mgo.v2 v2.0.0-20190816093944-a6b53ec6cb22
)

replace github.com/ONSdigital/dp-mongodb v1.5.0 => github.com/ONSdigital/dp-mongodb v1.5.1-0.20210526170525-d227b4ed13f5
