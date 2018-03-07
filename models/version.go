package models

// Instance represents the json returned from the dataset API for an instance document
type Version struct {
	ID      string       `json:"id,omitempty"`
	Links   VersionLinks `json:"links,omitempty"`
	State   string       `json:"state,omitempty"`
	Version int          `json:"version,omitempty"`
}

// InstanceLinks represents a list of link objects related to the instance resource
type VersionLinks struct {
	Dataset LinkObject `bson:"dataset,omitempty" json:"dataset"`
	Edition LinkObject `bson:"edition,omitempty" json:"edition"`
	Self    LinkObject `bson:"self,omitempty"    json:"self"`
	Version LinkObject `bson:"version,omitempty" json:"version"`
}
