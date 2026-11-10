package javadiff

import (
	"reflect"
	"testing"
)

// CVE:F10: lines of field initializers and initializer blocks belong to
// <clinit> (static) or to every constructor (<init> with
// InitializerParams), the bytecode methods their code compiles into.
func TestInitializers(t *testing.T) {
	src := `package p;

import java.util.*;

public class MimeTypeUtils {
    private static final Random RANDOM =
            new Random();                           // line 6-7: <clinit>
    private final Map<String, Integer> cache = new HashMap<>(); // line 8: <init>
    @Deprecated(since = "5") private int plain;  // line 9: no initializer
    static {
        RANDOM.setSeed(1);                         // line 11: <clinit>
    }
    {
        cache.clear();                             // line 14: <init>
    }
    public MimeTypeUtils() {
        plain = 1;                                 // line 17: <init>()
    }
    boolean same(int a, int b) { return a == b; } // line 19: same
    interface Limits {
        int MAX = 10;                              // line 21: Limits <clinit>
        int max();
    }
}
`
	got := map[int]string{}
	for _, l := range []int{6, 7, 8, 9, 11, 14, 17, 19, 21} {
		ms := MethodsForLines(src, []int{l})
		if len(ms) == 1 {
			got[l] = ms[0].Key()
		}
	}
	want := map[int]string{
		6:  "MimeTypeUtils#<clinit>()",
		7:  "MimeTypeUtils#<clinit>()",
		8:  "MimeTypeUtils#<init>(*)",
		11: "MimeTypeUtils#<clinit>()",
		14: "MimeTypeUtils#<init>(*)",
		17: "MimeTypeUtils#<init>()",
		19: "MimeTypeUtils#same(int,int)",
		21: "MimeTypeUtils$Limits#<clinit>()",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("owners\n got %v\nwant %v", got, want)
	}
	// Declared lists the synthetic initializers once per class and kind.
	var keys []string
	for _, m := range Declared(src) {
		keys = append(keys, m.Key())
	}
	wantKeys := []string{"MimeTypeUtils#<clinit>()", "MimeTypeUtils#<init>(*)", "MimeTypeUtils#<init>()", "MimeTypeUtils#same(int,int)", "MimeTypeUtils$Limits#<clinit>()"}
	if !reflect.DeepEqual(keys, wantKeys) {
		t.Errorf("declared %v, want %v", keys, wantKeys)
	}
}
