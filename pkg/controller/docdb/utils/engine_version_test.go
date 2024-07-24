package utils

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestEngineVersionIsHigherOrEqual(t *testing.T) {
	type args struct {
		spec    string
		current string
	}

	type want struct {
		result int
	}

	cases := map[string]struct {
		args
		want
	}{
		"DocDBIsEqual": {
			args: args{
				spec:    "5.0.0",
				current: "5.0.0",
			},
			want: want{
				result: 0,
			},
		},
		"DocDBIsEqual2": {
			args: args{
				spec:    "5.0",
				current: "5.0.0",
			},
			want: want{
				result: 0,
			},
		},
		"DocDBIsHigher": {
			args: args{
				spec:    "5.0",
				current: "4.0.0",
			},
			want: want{
				result: 1,
			},
		},
		"DocDBIsLower": {
			args: args{
				spec:    "4.0.0",
				current: "5.0.0",
			},
			want: want{
				result: -1,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			specV := ParseEngineVersion(tc.args.spec)
			curV := ParseEngineVersion(tc.args.current)

			res := specV.Compare(curV)
			resSign := sign(res)
			if diff := cmp.Diff(tc.want.result, resSign); diff != "" {
				t.Errorf("r: -want, +got:\n%q\n%q\n%s", tc.args.spec, tc.args.current, diff)
			}
		})
	}
}

func sign(x int) int {
	if x < 0 {
		return -1
	}
	if x > 0 {
		return 1
	}
	return 0
}
