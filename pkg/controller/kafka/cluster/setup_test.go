package cluster

import (
	"testing"
)

func TestSortBootstrapBrokerString(t *testing.T) {
	type args struct {
		input string
	}

	type want struct {
		result string
	}

	cases := map[string]struct {
		args
		want
	}{
		"EmptyString": {
			args: args{
				input: "",
			},
			want: want{
				result: "",
			},
		},

		"SingleEndpointString": {
			args: args{
				input: "b-1.test.0abcde.c6.kafka.eu-central-1.amazonaws.com:9098",
			},
			want: want{
				result: "b-1.test.0abcde.c6.kafka.eu-central-1.amazonaws.com:9098",
			},
		},

		"MultipleEndpointsString": {
			args: args{
				input: "b-3.test.0abcde.c6.kafka.eu-central-1.amazonaws.com:9098,b-2.test.0abcde.c6.kafka.eu-central-1.amazonaws.com:9098,b-1.test.0abcde.c6.kafka.eu-central-1.amazonaws.com:9098",
			},
			want: want{
				result: "b-1.test.0abcde.c6.kafka.eu-central-1.amazonaws.com:9098,b-2.test.0abcde.c6.kafka.eu-central-1.amazonaws.com:9098,b-3.test.0abcde.c6.kafka.eu-central-1.amazonaws.com:9098",
			},
		},

		"WithSpacesEndpointsString": {
			args: args{
				input: "b-3.test.0abcde.c6.kafka.eu-central-1.amazonaws.com:9098, b-2.test.0abcde.c6.kafka.eu-central-1.amazonaws.com:9098, b-1.test.0abcde.c6.kafka.eu-central-1.amazonaws.com:9098",
			},
			want: want{
				result: "b-1.test.0abcde.c6.kafka.eu-central-1.amazonaws.com:9098,b-2.test.0abcde.c6.kafka.eu-central-1.amazonaws.com:9098,b-3.test.0abcde.c6.kafka.eu-central-1.amazonaws.com:9098",
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			out := sortBootstrapBrokerString(tc.input)
			if out != tc.result {
				t.Errorf("For input '%s', expected '%s', but got '%s'", tc.input, tc.result, out)
			}
		})
	}
}
