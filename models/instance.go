package models

type Instance struct {
	Links InstanceLinks `json:"links,omitempty"`
	State string        `json:"state,omitempty"`
}
type InstanceLinks struct {
	Dataset LinkObject `bson:"dataset,omitempty" json:"dataset"`
	Edition LinkObject `bson:"edition,omitempty" json:"edition"`
	Self    LinkObject `bson:"self,omitempty"    json:"self"`
	Version LinkObject `bson:"version,omitempty" json:"version"`
}
