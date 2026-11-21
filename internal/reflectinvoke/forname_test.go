package reflectinvoke

import (
	"fmt"
	"reflect"
	"testing"

	"ariadne/internal/hierarchy"
)

// TestFindForNameInits (REFLECTION R4 r4-1) on testdata/reflection: the
// literal Class.forName sites that initialize a class with an initializer
// — directly, through the superclass, and with a non-constant initialize —
// and none for initialize=false or a class with no initializer.
func TestFindForNameInits(t *testing.T) {
	h, _, err := hierarchy.Build("../../testdata/reflection/app.jar", nil)
	if err != nil {
		t.Fatal(err)
	}
	inits, st, warnings := FindForNameInits(h)
	if len(warnings) != 0 {
		t.Errorf("warnings: %v", warnings)
	}
	var got []string
	for _, i := range inits {
		got = append(got, fmt.Sprintf("%s#%s %s -> %s maybe=%v", i.CallerFQCN, i.CallerName, i.Named, i.InitFQCN, i.MaybeInit))
	}
	want := []string{
		"fx.r.Main#forNameInit fx.r.Inited -> fx.r.Inited maybe=false",
		"fx.r.Main#forNameSuper fx.r.InitedSub -> fx.r.InitedBase maybe=false",
		"fx.r.Main#forNameMaybe fx.r.MaybeInited -> fx.r.MaybeInited maybe=true",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("inits\n got %q\nwant %q", got, want)
	}
	// 6 sites, all literal: Inited, NotInited (initialize=false), InitedSub,
	// MaybeInited, Target and Made (no initializer).
	if want := (ForNameStats{Sites: 6, Literal: 6, NoInit: 1, NoInitializer: 2, Inits: 3, MaybeInit: 1}); st != want {
		t.Errorf("stats %+v, want %+v", st, want)
	}
}
