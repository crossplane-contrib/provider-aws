package utils

import (
	"strconv"
	"strings"
)

// EngineVersion represents an AWS DocDB engine version.
type EngineVersion []any

// ParseEngineVersion from a raw string.
func ParseEngineVersion(raw string) EngineVersion {
	split := strings.Split(raw, ".")

	v := make(EngineVersion, len(split))
	for i, s := range split {
		d, err := strconv.Atoi(s)
		if err != nil {
			v[i] = s
		} else {
			v[i] = d
		}
	}
	return v
}

const (
	compareIsHigher = 1
	compareIsEqual  = 0
	compareIsLower  = -1
)

// Compare returns a positive value if v is represents a higher version number
// than other. A negative value is returned if other is higher than v.
// It returns 0 if both are considered equal.
func (v EngineVersion) Compare(other EngineVersion) int {
	if other == nil {
		return compareIsHigher
	}

	for i := 0; i < len(v); i++ {
		a := v.get(i)
		b := other.get(i)
		c := compareVersionComponents(a, b)
		if c != 0 {
			return c
		}
	}
	return compareIsEqual
}

func compareVersionComponents(a, b any) int {
	if a == b {
		return compareIsEqual
	}
	if b == nil {
		return compareIsHigher
	}
	aI, aIsInt := a.(int)
	bI, bIsInt := b.(int)
	if aIsInt {
		if bIsInt {
			return aI - bI
		}
		return compareIsHigher
	}
	if bIsInt {
		return compareIsLower
	}
	return compareIsEqual // We cannot decide if both are strings.
}

func (v EngineVersion) get(i int) any {
	if i >= 0 && i < len(v) {
		return v[i]
	}
	return nil
}

// CompareEngineVersions is a shortcut to compare two engine versions.
func CompareEngineVersions(a, b string) int {
	av := ParseEngineVersion(a)
	bv := ParseEngineVersion(b)
	return av.Compare(bv)
}
