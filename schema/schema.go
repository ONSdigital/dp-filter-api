package schema

import "github.com/ONSdigital/go-ns/avro"

var filterJobSubmitted = `{
  "type": "record",
  "name": "filter-job-submitted",
  "fields": [
    {"name": "filter_job_id", "type": "string"}
  ]
}`

// FilterJobSubmittedSchema is the Avro schema for each
// filter job submitted
var FilterJobSubmittedSchema *avro.Schema = &avro.Schema{
	Definition: filterJobSubmitted,
}
