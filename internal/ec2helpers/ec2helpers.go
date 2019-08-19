package ec2helpers

import (
	"github.com/aws/aws-sdk-go/service/ec2"
)

// TagMap is just a string:string map with some helpers
type TagMap map[string]string

// MapTags converts an array of ec2.Tag into a much-more-convenient Go map
func MapTags(tags []*ec2.Tag) TagMap {
	m := make(TagMap)
	for _, tag := range tags {
		m[*tag.Key] = *tag.Value
	}
	return m
}

// Get returns a tag value from the map, or an empty string if the tag is not
// found in the map, similar to os.GetEnv. For convenience and cleanliness at
// caller.
func (m TagMap) Get(tag string) string {
	return m.GetWithDefault(tag, "")
}

// GetWithDefault returns a tag value from the map, or a specified default
// value if it is not found in the map.
func (m TagMap) GetWithDefault(tag, defval string) string {
	v, ok := m[tag]
	if ok {
		return v
	}
	return defval
}

// Has returns true if a tag exists in the map AND has the specified value
func (m TagMap) Has(tag, value string) bool {
	v, ok := m[tag]
	return ok && v == value
}
