package tests

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	testNameFinder "github.com/takaaa220/test-name-finder"
)

func Test_Integration(t *testing.T) {
	type args struct {
		fileName    string
		lineNumber  int
		startCursor int
		endCursor   int
	}

	cases := map[string]struct {
		args    args
		want    *testNameFinder.TestName
		wantErr bool
	}{
		"failure_file_not_found": {
			args: args{
				fileName:    "not_found.go",
				lineNumber:  1,
				startCursor: 0,
				endCursor:   0,
			},
			wantErr: true,
		},
		"failure_line_number_is_zero": {
			args: args{
				fileName:    "table_is_map_test.go",
				lineNumber:  0,
				startCursor: 0,
				endCursor:   0,
			},
			wantErr: true,
		},
		"failure_start_cursor_is_minus": {
			args: args{
				fileName:    "table_is_map_test.go",
				lineNumber:  1,
				startCursor: -1,
				endCursor:   0,
			},
			wantErr: true,
		},
		"failure_start_cursor_is_greater_than_end_cursor": {
			args: args{
				fileName:    "table_is_map_test.go",
				lineNumber:  1,
				startCursor: 1,
				endCursor:   0,
			},
			wantErr: true,
		},
		"failure_test_name_not_found": {
			args: args{
				fileName:    "table_is_map_test.go",
				lineNumber:  7,
				startCursor: 0,
				endCursor:   0,
			},
			wantErr: true,
		},
		"success_table_is_map_1": {
			args: args{
				fileName:    "table_is_map_test.go",
				lineNumber:  19,
				startCursor: 3,
				endCursor:   8,
			},
			want: &testNameFinder.TestName{
				FuncName: "Test_TableIsMap",
				TestCase: "test1",
			},
		},
		"success_table_is_map_2": {
			args: args{
				fileName:    "table_is_map_test.go",
				lineNumber:  27,
				startCursor: 10,
				endCursor:   17,
			},
			want: &testNameFinder.TestName{
				FuncName: "Test_TableIsMap",
				TestCase: "test2 test2",
			},
		},
		"success_table_is_map_3": {
			args: args{
				fileName:    "table_is_map_test.go",
				lineNumber:  19,
				startCursor: 0,
				endCursor:   0,
			},
			want: &testNameFinder.TestName{
				FuncName: "Test_TableIsMap",
				TestCase: "test1",
			},
		},
		"success_table_is_map_4": {
			args: args{
				fileName:    "table_is_map_test.go",
				lineNumber:  25,
				startCursor: 4,
				endCursor:   4,
			},
			want: &testNameFinder.TestName{
				FuncName: "Test_TableIsMap",
				TestCase: "test1",
			},
		},
		"success_table_is_map_out_of_map_1": {
			args: args{
				fileName:    "table_is_map_test.go",
				lineNumber:  18,
				startCursor: 3,
				endCursor:   3,
			},
			want: &testNameFinder.TestName{
				FuncName: "Test_TableIsMap",
				TestCase: "",
			},
		},
		"success_table_is_map_out_of_map_2": {
			args: args{
				fileName:    "table_is_map_test.go",
				lineNumber:  33,
				startCursor: 2,
				endCursor:   2,
			},
			want: &testNameFinder.TestName{
				FuncName: "Test_TableIsMap",
				TestCase: "",
			},
		},
		"success_table_is_slice_1": {
			args: args{
				fileName:    "table_is_slice_test.go",
				lineNumber:  29,
				startCursor: 10,
				endCursor:   21,
			},
			want: &testNameFinder.TestName{
				FuncName: "Test_TableIsSlice",
				TestCase: "test2 test2",
			},
		},
		"success_table_is_slice_2": {
			args: args{
				fileName:    "table_is_slice_test.go",
				lineNumber:  20,
				startCursor: 0,
				endCursor:   0,
			},
			want: &testNameFinder.TestName{
				FuncName: "Test_TableIsSlice",
				TestCase: "test1",
			},
		},
		"success_table_is_slice_3": {
			args: args{
				fileName:    "table_is_slice_test.go",
				lineNumber:  35,
				startCursor: 4,
				endCursor:   4,
			},
			want: &testNameFinder.TestName{
				FuncName: "Test_TableIsSlice",
				TestCase: "test2 test2",
			},
		},
		"success_table_is_slice_out_of_slice_1": {
			args: args{
				fileName:    "table_is_slice_test.go",
				lineNumber:  19,
				startCursor: 3,
				endCursor:   3,
			},
			want: &testNameFinder.TestName{
				FuncName: "Test_TableIsSlice",
				TestCase: "",
			},
		},
		"success_table_is_slice_out_of_slice_2": {
			args: args{
				fileName:    "table_is_slice_test.go",
				lineNumber:  36,
				startCursor: 1,
				endCursor:   1,
			},
			want: &testNameFinder.TestName{
				FuncName: "Test_TableIsSlice",
				TestCase: "",
			},
		},
		"success_table_is_not_supported_slice": {
			args: args{
				fileName:    "table_is_not_supported_slice_test.go",
				lineNumber:  21,
				startCursor: 10,
				endCursor:   13,
			},
			want: &testNameFinder.TestName{
				FuncName: "Test_TableIsNotSupportedSlice",
				TestCase: "",
			},
		},
	}

	for name, tt := range cases {
		tt := tt
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			pwd := os.Getenv("PWD")
			filePath := filepath.Join(pwd, "testdata", tt.args.fileName)

			out, err := testNameFinder.FindTestName(filePath, testNameFinder.Selection{
				LineNumber:  tt.args.lineNumber,
				StartCursor: tt.args.startCursor,
				EndCursor:   tt.args.endCursor,
			})
			if err != nil {
				if !tt.wantErr {
					t.Errorf("err should be nil, but got %v", err)
				}

				return
			}

			if diff := cmp.Diff(tt.want, out); diff != "" {
				t.Errorf("out mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
