package dataflow

import "testing"

// TestMayDeclareOutside: a supertype outside the input blocks
// VirtualCallers unless it provably cannot declare the method — a marker
// interface, or a JDK type and a descriptor naming a non-JDK type.
func TestMayDeclareOutside(t *testing.T) {
	cases := []struct {
		super, desc string
		want        bool
	}{
		{"java.io.Serializable", "(Ljava/lang/Object;)V", false},
		{"java.lang.ref.ReferenceQueue", "(Ljava/lang/ClassLoader;Ljava/util/concurrent/Callable;)Ljava/lang/Class;", true},
		{"java.lang.ref.ReferenceQueue", "(Lorg/example/Provider;)Ljava/util/Collection;", false},
		{"java.util.AbstractList", "(I)[Lcom/acme/Row;", false},
		{"javax.servlet.GenericServlet", "(Lorg/example/X;)V", true}, // not the JDK: may come from any jar
		{"org.other.Base", "(Lorg/example/X;)V", true},
	}
	for _, c := range cases {
		if got := mayDeclareOutside(c.super, c.desc); got != c.want {
			t.Errorf("mayDeclareOutside(%s, %s) = %v, want %v", c.super, c.desc, got, c.want)
		}
	}
}
