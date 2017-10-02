package models

type Instance struct {
	Links InstanceLinks `json:"links,omitempty"`
}
type InstanceLinks struct {
	Version LinkObject `bson:"version,omitempty"   json:"version"`
}
