package mode

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestConfigUnmarshalSupportsInfoFindField(t *testing.T) {
	var conf Config
	err := yaml.Unmarshal([]byte("infoFind:\n  Email:\n    - test@example\\.com\n"), &conf)
	if err != nil {
		t.Fatal(err)
	}

	if got := conf.InfoFind["Email"]; len(got) != 1 || got[0] != "test@example\\.com" {
		t.Fatalf("InfoFind = %#v, want Email pattern from infoFind", conf.InfoFind)
	}
}

func TestConfigUnmarshalSupportsLegacyInfoFilerField(t *testing.T) {
	var conf Config
	err := yaml.Unmarshal([]byte("infoFiler:\n  Email:\n    - test@example\\.com\n"), &conf)
	if err != nil {
		t.Fatal(err)
	}

	if got := conf.InfoFind["Email"]; len(got) != 1 || got[0] != "test@example\\.com" {
		t.Fatalf("InfoFind = %#v, want Email pattern from legacy infoFiler", conf.InfoFind)
	}
}
