package models

type Instance struct {
	Links InstanceLinks `json:"links,omitempty"`
	State string        `json:"state,omitempty"`
}
type InstanceLinks struct {
	Version LinkObject `bson:"version,omitempty"   json:"version"`
}
