package testdata

import (
	"fmt"
	"testing"
)

func Test_TableIsMap(t *testing.T) {
	type args struct {
		arg1 string
		arg2 int
	}

	cases := map[string]struct {
		args    args
		want    string
		wantErr bool
	}{
		"test1": {
			args: args{
				arg1: "test1",
				arg2: 10,
			},
			want: "test1",
		},
		"test2 test2": {
			args: args{
				arg1: "test2",
				arg2: 20,
			},
			wantErr: false,
		},
	}

	for name, tt := range cases {
		tt := tt

		t.Run(name, func(t *testing.T) {
			t.Parallel()
			fmt.Println(tt)
		})
	}
}
