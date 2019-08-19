package ec2helpers

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type hastestcase struct {
	Input      []*ec2.Tag
	WantKey    string
	WantValue  string
	WantResult bool
}

func TestHas(t *testing.T) {
	notags := []*ec2.Tag{}
	onetag := []*ec2.Tag{&ec2.Tag{Key: aws.String("thing"), Value: aws.String("xyz")}}
	testcases := []hastestcase{
		hastestcase{Input: notags, WantKey: "xyz", WantValue: "abc", WantResult: false},
		hastestcase{Input: notags, WantKey: "", WantValue: "", WantResult: false},
		hastestcase{Input: onetag, WantKey: "thing", WantValue: "xyz", WantResult: true},
		hastestcase{Input: onetag, WantKey: "thing", WantValue: "blargh", WantResult: false},
		hastestcase{Input: onetag, WantKey: "blorp", WantValue: "blorp", WantResult: false},
		hastestcase{Input: onetag, WantKey: "", WantValue: "", WantResult: false},
	}
	for _, test := range testcases {
		m := MapTags(test.Input)
		if r := m.Has(test.WantKey, test.WantValue); r != test.WantResult {
			t.Errorf("Has check for %q == %q wanted %v, got %v",
				test.WantKey, test.WantValue, test.WantResult, r)
		}
	}
}

type gettestcase struct {
	Input     []*ec2.Tag
	WantKey   string
	WantValue string
}

func TestGet(t *testing.T) {
	notags := []*ec2.Tag{}
	onetag := []*ec2.Tag{&ec2.Tag{Key: aws.String("thing"), Value: aws.String("xyz")}}
	testcases := []gettestcase{
		gettestcase{notags, "", ""},
		gettestcase{notags, "thing", ""},
		gettestcase{onetag, "thing", "xyz"},
		gettestcase{onetag, "", ""},
	}
	for _, test := range testcases {
		m := MapTags(test.Input)
		if r := m.Get(test.WantKey); r != test.WantValue {
			t.Errorf("Get check for %q == %q, got %q", test.WantKey, test.WantValue, r)
		}
	}
}

func TestGetWithDefault(t *testing.T) {
	notags := []*ec2.Tag{}
	onetag := []*ec2.Tag{&ec2.Tag{Key: aws.String("thing"), Value: aws.String("xyz")}}
	testcases := []gettestcase{
		gettestcase{notags, "", "-"},
		gettestcase{notags, "thing", "-"},
		gettestcase{onetag, "thing", "xyz"},
		gettestcase{onetag, "", "-"},
	}
	for _, test := range testcases {
		m := MapTags(test.Input)
		if r := m.GetWithDefault(test.WantKey, "-"); r != test.WantValue {
			t.Errorf("GetWithDefault check for %q == %q, got %q", test.WantKey, test.WantValue, r)
		}
	}
}
