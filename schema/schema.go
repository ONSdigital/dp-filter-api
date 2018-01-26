package schema

import "github.com/ONSdigital/go-ns/avro"

var filterOutputSubmitted = `{
  "type": "record",
  "name": "filter-output-submitted",
  "fields": [
    {"name": "filter_output_id", "type": "string", "default": ""},
    {"name": "instance_id", "type": "string", "default": ""},
    {"name": "dataset_id", "type": "string", "default": ""},
    {"name": "edition", "type": "string", "default": ""},
    {"name": "version", "type": "string", "default": ""}
  ]
}`

// FilterSubmittedSchema is the Avro schema for each
// filter output submitted
var FilterSubmittedSchema *avro.Schema = &avro.Schema{
	Definition: filterOutputSubmitted,
}
