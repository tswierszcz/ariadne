package javadiff

// stripCommentsAndStrings blanks out every comment, string literal, text
// block and char literal in a Java source file, replacing their content
// with spaces (newlines kept as newlines) so brace matching never gets
// confused by a stray '{'/'}'/';' inside one, while every real line
// number stays exactly where it was.
func stripCommentsAndStrings(src string) string {
	out := make([]byte, len(src))
	copy(out, src)
	n := len(out)

	blank := func(from, to int) { // [from, to), leaving newlines alone
		for i := from; i < to; i++ {
			if out[i] != '\n' {
				out[i] = ' '
			}
		}
	}

	for i := 0; i < n; i++ {
		switch {
		case out[i] == '/' && i+1 < n && out[i+1] == '/':
			start := i
			for i < n && out[i] != '\n' {
				i++
			}
			blank(start, i)
			// leave i at the '\n' (or EOF); the loop's i++ moves past it

		case out[i] == '/' && i+1 < n && out[i+1] == '*':
			start := i
			i += 2
			for i+1 < n && !(out[i] == '*' && out[i+1] == '/') {
				i++
			}
			end := i + 2
			if end > n {
				end = n
			}
			blank(start, end)
			i = end - 1

		case out[i] == '"' && i+2 < n && out[i+1] == '"' && out[i+2] == '"':
			start := i
			i += 3
			for i+2 < n && !(out[i] == '"' && out[i+1] == '"' && out[i+2] == '"') {
				if out[i] == '\\' && i+1 < n {
					i++
				}
				i++
			}
			end := i + 3
			if end > n {
				end = n
			}
			blank(start, end)
			i = end - 1

		case out[i] == '"':
			start := i
			i++
			for i < n && out[i] != '"' && out[i] != '\n' {
				if out[i] == '\\' && i+1 < n {
					i++
				}
				i++
			}
			end := i + 1
			if end > n {
				end = n
			}
			blank(start, end)
			i = end - 1

		case out[i] == '\'':
			start := i
			i++
			for i < n && out[i] != '\'' && out[i] != '\n' {
				if out[i] == '\\' && i+1 < n {
					i++
				}
				i++
			}
			end := i + 1
			if end > n {
				end = n
			}
			blank(start, end)
			i = end - 1
		}
	}
	return string(out)
}
