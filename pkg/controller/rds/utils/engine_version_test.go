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
		"AuroraMySQLIsEqual": {
			args: args{
				spec:    "5",
				current: "5.7.mysql_aurora.2.07.0",
			},
			want: want{
				result: 0,
			},
		},
		"AuroraMySQLIsEqual2": {
			args: args{
				spec:    "5.7",
				current: "5.7.mysql_aurora.2.07.0",
			},
			want: want{
				result: 0,
			},
		},
		"AuroraMySQLIsHigher": {
			args: args{
				spec:    "8",
				current: "5.7.mysql_aurora.2.07.0",
			},
			want: want{
				result: 1,
			},
		},
		"AuroraPostgresIsEqual": {
			args: args{
				spec:    "14",
				current: "14.4",
			},
			want: want{
				result: 0,
			},
		},
		"AuroraPostgresIsLower": {
			args: args{
				spec:    "14",
				current: "15.2",
			},
			want: want{
				result: -1,
			},
		},
		"MariaDBIsEqual": {
			args: args{
				spec:    "10",
				current: "10.3.35",
			},
			want: want{
				result: 0,
			},
		},
		"MariaDBNotIsLower": {
			args: args{
				spec:    "10",
				current: "11.0",
			},
			want: want{
				result: -1,
			},
		},
		"MySQLIsEqual": {
			args: args{
				spec:    "5.7",
				current: "5.7.38",
			},
			want: want{
				result: 0,
			},
		},
		"MySQLIsHigher": {
			args: args{
				spec:    "8",
				current: "5.7.38",
			},
			want: want{
				result: 1,
			},
		},
		"OracleIsEqual": {
			args: args{
				spec:    "19",
				current: "19.0.0.0.ru-2023-01.rur-2023-01.r1",
			},
			want: want{
				result: 0,
			},
		},
		"OracleIsHigher": {
			args: args{
				spec:    "20",
				current: "19.0.0.0.ru-2023-01.rur-2023-01.r1",
			},
			want: want{
				result: 1,
			},
		},
		"PostgresIsEqual": {
			args: args{
				spec:    "10",
				current: "10.17",
			},
			want: want{
				result: 0,
			},
		},
		"PostgresIsHigher": {
			args: args{
				spec:    "14",
				current: "10.17",
			},
			want: want{
				result: 1,
			},
		},
		"SQLServerIsEqual": {
			args: args{
				spec:    "12",
				current: "12.00.6293.0.v1",
			},
			want: want{
				result: 0,
			},
		},
		"SQLServerIsHigher": {
			args: args{
				spec:    "14",
				current: "12.00.6293.0.v1",
			},
			want: want{
				result: 1,
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
