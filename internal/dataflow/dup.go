package dataflow

// The dup/dup2 family (JVMS 6.5 dup_x1, dup_x2, dup2, dup2_x1, dup2_x2):
// pure stack reshuffles, width-dependent. Each form below is named after
// the spec's own "Form N" split by operand widths. No new Values are
// created here — only existing stack entries move, so the analysis never
// loses precision at a dup, only at the (separately handled) ops that
// actually discard it.

func stepDupX1(st *absState) string {
	vs, ok := st.popN(2)
	if !ok || vs[0].width != 1 || vs[1].width != 1 {
		return "dup_x1 on fewer than two category-1 values"
	}
	v2, v1 := vs[0], vs[1] // v1 was on top
	st.pushEntry(v1)
	st.pushEntry(v2)
	st.pushEntry(v1)
	return ""
}

func stepDupX2(st *absState) string {
	top, ok := st.pop()
	if !ok || top.width != 1 {
		return "dup_x2 on a missing or category-2 top-of-stack value"
	}
	next, ok := st.pop()
	if !ok {
		return "dup_x2 with fewer than two stack entries"
	}
	if next.width == 2 {
		// Form 2: v1(cat1), v2(cat2) -> v1, v2, v1
		st.pushEntry(top)
		st.pushEntry(next)
		st.pushEntry(top)
		return ""
	}
	third, ok := st.pop()
	if !ok || third.width != 1 {
		return "dup_x2 form 1 requires three category-1 values"
	}
	// Form 1: v1,v2,v3(all cat1) -> v1, v3, v2, v1
	st.pushEntry(top)
	st.pushEntry(third)
	st.pushEntry(next)
	st.pushEntry(top)
	return ""
}

func stepDup2(st *absState) string {
	top, ok := st.pop()
	if !ok {
		return "stack underflow at dup2"
	}
	if top.width == 2 {
		// Form 2: v1(cat2) -> v1, v1
		st.pushEntry(top)
		st.pushEntry(top)
		return ""
	}
	second, ok := st.pop()
	if !ok || second.width != 1 {
		return "dup2 requires two category-1 values when the top is category-1"
	}
	// Form 1: v1,v2(cat1) -> v2, v1, v2, v1
	st.pushEntry(second)
	st.pushEntry(top)
	st.pushEntry(second)
	st.pushEntry(top)
	return ""
}

func stepDup2X1(st *absState) string {
	top, ok := st.pop()
	if !ok {
		return "stack underflow at dup2_x1"
	}
	if top.width == 2 {
		below, ok := st.pop()
		if !ok || below.width != 1 {
			return "dup2_x1 form 2 requires a category-1 value beneath a category-2 top"
		}
		// Form 2: v1(cat2), v2(cat1) -> v1, v2, v1
		st.pushEntry(top)
		st.pushEntry(below)
		st.pushEntry(top)
		return ""
	}
	second, ok := st.pop()
	if !ok || second.width != 1 {
		return "dup2_x1 requires two category-1 values on top"
	}
	third, ok := st.pop()
	if !ok || third.width != 1 {
		return "dup2_x1 form 1 requires a third category-1 value"
	}
	// Form 1: v1,v2,v3(all cat1) -> v2, v1, v3, v2, v1
	st.pushEntry(second)
	st.pushEntry(top)
	st.pushEntry(third)
	st.pushEntry(second)
	st.pushEntry(top)
	return ""
}

func stepDup2X2(st *absState) string {
	top, ok := st.pop()
	if !ok {
		return "stack underflow at dup2_x2"
	}
	if top.width == 2 {
		below, ok := st.pop()
		if !ok {
			return "stack underflow at dup2_x2 (category-2 top)"
		}
		if below.width == 2 {
			// Form 4: v1(cat2), v2(cat2) -> v1, v2, v1
			st.pushEntry(top)
			st.pushEntry(below)
			st.pushEntry(top)
			return ""
		}
		third, ok := st.pop()
		if !ok || third.width != 1 {
			return "dup2_x2 form 3 requires two category-1 values beneath a category-2 top"
		}
		// Form 3: v1(cat2), v2,v3(cat1) -> v1, v3, v2, v1
		st.pushEntry(top)
		st.pushEntry(third)
		st.pushEntry(below)
		st.pushEntry(top)
		return ""
	}
	second, ok := st.pop()
	if !ok || second.width != 1 {
		return "dup2_x2 requires two category-1 values on top"
	}
	third, ok := st.pop()
	if !ok {
		return "stack underflow at dup2_x2"
	}
	if third.width == 2 {
		// Form 2: v1,v2(cat1), v3(cat2) -> v2, v1, v3, v2, v1
		st.pushEntry(second)
		st.pushEntry(top)
		st.pushEntry(third)
		st.pushEntry(second)
		st.pushEntry(top)
		return ""
	}
	fourth, ok := st.pop()
	if !ok || fourth.width != 1 {
		return "dup2_x2 form 1 requires four category-1 values"
	}
	// Form 1: v1,v2,v3,v4(all cat1) -> v2, v1, v4, v3, v2, v1
	st.pushEntry(second)
	st.pushEntry(top)
	st.pushEntry(fourth)
	st.pushEntry(third)
	st.pushEntry(second)
	st.pushEntry(top)
	return ""
}
