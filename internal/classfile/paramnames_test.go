package classfile

import (
	"reflect"
	"testing"
)

func TestParameterNames(t *testing.T) {
	for _, tc := range []struct {
		desc   string
		static bool
		attr   []string
		locals map[int]string
		want   []string
	}{
		// MethodParameters wins.
		{"(JLjava/lang/String;)V", false, []string{"id", "name"}, map[int]string{1: "x", 3: "y"}, []string{"id", "name"}},
		// LocalVariableTable by slot: this=0, long takes 1-2.
		{"(JLjava/lang/String;)V", false, nil, map[int]string{0: "this", 1: "id", 3: "name"}, []string{"id", "name"}},
		{"(ILjava/lang/String;)V", true, nil, map[int]string{0: "n", 1: "s"}, []string{"n", "s"}},
		// An unnamed MethodParameters entry falls back to the table.
		{"(I)V", true, []string{""}, map[int]string{0: "n"}, []string{"n"}},
		// Incomplete: nil.
		{"(II)V", true, nil, map[int]string{0: "a"}, nil},
		{"()V", false, nil, nil, nil},
	} {
		if got := parameterNames(tc.desc, tc.static, tc.attr, tc.locals); !reflect.DeepEqual(got, tc.want) {
			t.Errorf("%s static=%v: %v, want %v", tc.desc, tc.static, got, tc.want)
		}
	}
}
