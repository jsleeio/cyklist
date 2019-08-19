package ec2discovery

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func TestTagFilter(t *testing.T) {
	tc := TagFilter("testkey", "testvalue")
	if *tc.Name != "tag:testkey" {
		t.Errorf("TagFilter name should be %q, got %q", "tag:testkey", *tc.Name)
	}
	if *tc.Values[0] != "testvalue" {
		t.Errorf("TagFilter first value should be %q, got %q", "testvalue", *tc.Values[0])
	}
	if len(tc.Values) != 1 {
		t.Errorf("TagFilter should have exactly one value, got %d", len(tc.Values))
	}
}

func TestNotImageId(t *testing.T) {
	tc := &ec2.Instance{ImageId: aws.String("abcdef")}
	shouldfalse := NotImageId("abcdef")
	if shouldfalse(tc) {
		t.Errorf("NotImageId funcs should return false for ImageId == %q when ImageId is %q", "abcdef", *tc.ImageId)
	}
	shouldtrue := NotImageId("ghijkl")
	if !shouldtrue(tc) {
		t.Errorf("NotImageId funcs should return true for ImageId == %q when ImageId is %q", "ghijkl", *tc.ImageId)
	}
}

func TestAgeAtLeast(t *testing.T) {
	now := time.Now().UTC()
	testtime := now.Add(-5 * time.Hour)
	instance := &ec2.Instance{LaunchTime: &testtime}
	testcases := []struct {
		age    time.Duration
		result bool
	}{
		{1 * time.Minute, false},
		{-1 * time.Hour, false},
		{-5 * time.Hour, true},
		{-6 * time.Hour, true},
	}
	for _, tc := range testcases {
		f := AgeAtLeast(tc.age)
		if !f(instance) {
			truedelta := now.Sub(testtime).String()
			t.Errorf("AgeAtLeast should return false for %v being a larger delta than %v", tc.age, truedelta)
		}
	}
}
