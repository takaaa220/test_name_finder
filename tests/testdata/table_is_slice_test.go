package testdata

import (
	"fmt"
	"testing"
)

func Test_TableIsSlice(t *testing.T) {
	type args struct {
		arg1 string
		arg2 int
	}

	cases := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "test1",
			args: args{
				arg1: "test1",
				arg2: 10,
			},
			want: "test1",
		},
		{
			name: "test2 test2",
			args: args{
				arg1: "test2",
				arg2: 20,
			},
			wantErr: false,
		},
	}

	for _, tt := range cases {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fmt.Println(tt)
		})
	}
}
