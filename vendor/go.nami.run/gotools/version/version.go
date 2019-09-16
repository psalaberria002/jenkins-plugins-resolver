package version

import (
	"encoding/json"
	"fmt"

	"github.com/juamedgod/semver"
)

// Version models a component version extracted from the upstream
type Version struct {
	sv  *semver.Version
	ctx map[string]string
}

// Context provides additional information about the parsed version such as its raw format
func (v *Version) Context() map[string]string {
	return v.ctx
}

// MarshalJSON Allows serializing the Version as a string
func (v *Version) MarshalJSON() (data []byte, err error) {
	return json.Marshal(v.String())
}

// UnmarshalJSON populates the Version from a JSON byte slice
func (v *Version) UnmarshalJSON(data []byte) error {
	str := ""
	if err := json.Unmarshal(data, &str); err != nil {
		return fmt.Errorf("malformed JSON data: %v", err)
	}
	v2, err := New(str)
	if err != nil {
		return err
	}
	v.sv = v2.sv
	return nil
}

// String returns the string representation of the Version
func (v *Version) String() string {
	if v == nil || v.sv == nil {
		return ""
	}
	return v.sv.String()
}

// Less return true if the version is lower than the provided version, false otherwise
func (v *Version) Less(version *Version) bool {
	return v.sv.Less(version.sv)
}

// Greater return true if the version is greater than the provided version, false otherwise
func (v *Version) Greater(v2 *Version) bool {
	if v2 == nil {
		return true
	}
	return v.sv.Greater(v2.sv)
}

// Major returns the version major component
func (v *Version) Major() int {
	return int(v.sv.Major)
}

// Minor returns the version minor component
func (v *Version) Minor() int {
	return int(v.sv.Minor)
}

// Patch returns the version patch component
func (v *Version) Patch() int {
	return int(v.sv.Patch)
}

// New returns a new Version from the provided string representation in data
// It allows providing a list of maps containing some context about the raw format of the version.
// For example, a version 2016.20.3 may have been parsed from a raw string 2016203, the context may contain
// information about how it was originally constructed.
func New(data string, contextList ...map[string]string) (*Version, error) {
	v := &Version{ctx: make(map[string]string, 0)}
	for _, ctx := range contextList {
		for key, value := range ctx {
			v.ctx[key] = value
		}
	}
	if sv, err := semver.ParseVersion(data); err == nil {
		v.sv = sv
	} else if sv, err := semver.ParsePermissiveVersion(data); err == nil {
		v.sv = sv
	} else {
		return nil, fmt.Errorf("error parsing semver version %q", data)
	}
	v.sv.Hack(semver.SupportRevisionsInPreRelease)
	return v, nil
}

// Matches checks if the Version matches the provided Semantic Version Expression
func (v *Version) Matches(expr semver.Expression) bool {
	if v.sv != nil {
		// Semver ranges always fail when pre-releases are involved but we will do our best here for now.
		cleanedVersion := semver.NewVersion(v.sv.Major, v.sv.Minor, v.sv.Patch)
		return expr.Matches(cleanedVersion)
	}
	return false
}
