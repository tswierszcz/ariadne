package flowsum

// The catalog states the exact dependence of a few JDK methods a data
// flow into a deserialization entry (or a taint analysis) passes through,
// so they are not answered by the conservative policy: a decoder's
// result depends on its argument, not on the decoder, and is not a new
// source. Keyed by the owner the bytecode names and the method name (any
// descriptor). Anything else is unmodeled: the conservative policy.

type flow struct {
	recv bool // the receiver
	args bool // every argument
}

type catalogEntry struct {
	result, intoRecv flow
}

var (
	fromArgs      = catalogEntry{result: flow{args: true}}
	fromRecv      = catalogEntry{result: flow{recv: true}}
	fromBoth      = catalogEntry{result: flow{recv: true, args: true}}
	wrapsArgs     = catalogEntry{intoRecv: flow{args: true}}                           // a constructor keeping its arguments
	appendsArgs   = catalogEntry{result: flow{recv: true}, intoRecv: flow{args: true}} // a builder's append
	noDependences = catalogEntry{}
)

var catalog = map[string]catalogEntry{}

func init() {
	add := func(e catalogEntry, owner string, names ...string) {
		for _, n := range names {
			catalog[owner+"#"+n] = e
		}
	}
	for _, sb := range []string{"java.lang.StringBuilder", "java.lang.StringBuffer"} {
		add(wrapsArgs, sb, "<init>")
		add(appendsArgs, sb, "append", "insert")
		add(fromRecv, sb, "toString", "substring", "charAt")
	}
	add(wrapsArgs, "java.lang.String", "<init>")
	add(fromArgs, "java.lang.String", "valueOf", "copyValueOf", "format", "join")
	add(fromBoth, "java.lang.String", "concat", "replace", "replaceAll", "replaceFirst", "substring", "trim", "strip",
		"toLowerCase", "toUpperCase", "getBytes", "toCharArray", "toString", "intern", "split", "subSequence")
	add(fromArgs, "java.util.Base64$Decoder", "decode")
	add(fromArgs, "java.util.Base64$Encoder", "encode", "encodeToString")
	add(noDependences, "java.util.Base64", "getDecoder", "getEncoder", "getMimeDecoder", "getMimeEncoder", "getUrlDecoder", "getUrlEncoder")
	for _, w := range []string{"java.io.ByteArrayInputStream", "java.io.ObjectInputStream", "java.io.InputStreamReader",
		"java.io.BufferedReader", "java.io.StringReader", "java.io.BufferedInputStream", "java.io.DataInputStream",
		"java.util.zip.GZIPInputStream", "java.util.zip.InflaterInputStream", "java.io.CharArrayReader"} {
		add(wrapsArgs, w, "<init>")
		add(fromRecv, w, "read", "readAllBytes", "readNBytes", "readLine", "readObject", "readUnshared", "readUTF")
	}
	add(fromRecv, "java.io.InputStream", "read", "readAllBytes", "readNBytes")
	add(fromRecv, "java.io.Reader", "read")
	add(fromArgs, "java.util.Optional", "of", "ofNullable")
	add(fromBoth, "java.util.Optional", "get", "orElse", "orElseThrow")
	add(fromArgs, "java.util.Objects", "requireNonNull", "toString")
	for _, n := range []string{"java.lang.Integer", "java.lang.Long", "java.lang.Short", "java.lang.Byte", "java.lang.Boolean", "java.lang.Double"} {
		add(fromArgs, n, "valueOf", "parseInt", "parseLong", "parseShort", "parseByte", "parseBoolean", "parseDouble", "toString")
	}
	add(fromArgs, "java.net.URLDecoder", "decode")
	add(noDependences, "java.lang.Object", "<init>")
	// Configuring a stream does not change the data it reads.
	add(noDependences, "java.io.ObjectInputStream", "setObjectInputFilter")
	add(fromArgs, "java.io.ObjectInputFilter$Config", "createFilter")
}

func lookupCatalog(owner, name string) (catalogEntry, bool) {
	e, ok := catalog[owner+"#"+name]
	return e, ok
}

func (f flow) apply(recv Dep, args []Dep) Dep {
	var d Dep
	if f.recv {
		d = union(d, recv)
	}
	if f.args {
		for _, a := range args {
			d = union(d, a)
		}
	}
	return d
}
