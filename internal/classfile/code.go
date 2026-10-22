package classfile

import (
	"bytes"
	"fmt"
)

// ExceptionHandler is one exception_table entry (JVM spec 4.7.3): the
// bytecode range [StartPC, EndPC) is guarded by HandlerPC for
// CatchType, or for every exception when CatchType == "" (catch_type
// index 0 — used both for a compiled `finally` block and an explicit
// `catch (Throwable ...)`, which the class file format does not
// distinguish). DATA_FLOW D1's CFG builder turns these into exceptional
// edges, sharing this data with JVM_GRAPH G3's own class-initialization
// work only in spirit (G3 needed none of this; it's named here only
// because the plan calls out "uzgodnić współdzielenie z JVM G3" and the
// answer is: there is nothing to share, G3's scope never touched the
// exception table).
type ExceptionHandler struct {
	StartPC   int
	EndPC     int
	HandlerPC int
	CatchType string // "" for catch-all
}

// LineNumberEntry is one LineNumberTable entry (JVM spec 4.7.12):
// source line Line starts at bytecode offset PC. Optional presentation
// data only (DATA_FLOW D1: "zachować opcjonalne numery linii do
// prezentacji; brak debug data nie zatrzymuje analizy") — nothing in
// this package requires it to be present.
type LineNumberEntry struct {
	PC   int
	Line int
}

// CodeAttr is a Code attribute's structured content (JVM spec 4.7.3).
// DATA_FLOW D1's CFG builder needs MaxStack/MaxLocals (stack-category
// bookkeeping) and ExceptionTable (exceptional edges) beyond the raw
// Code bytes classfile.ExtractCalls already used.
type CodeAttr struct {
	MaxStack       int
	MaxLocals      int
	Code           []byte
	ExceptionTable []ExceptionHandler
	// LineNumbers is nil when the class was compiled without debug info
	// (javac -g:none, or any compiler that omits it) — absence is not an
	// error.
	LineNumbers []LineNumberEntry
}

// parseCodeAttribute parses a Code attribute's bytes (JVM spec 4.7.3):
// the raw bytecode, the exception table, and — if present — the nested
// LineNumberTable attribute. Every other nested attribute
// (StackMapTable, LocalVariableTable, ...) is skipped, same as before
// DATA_FLOW D1: nothing in this package recomputes a verifier's own
// stack map, and local variable names are not needed for reachability.
func parseCodeAttribute(data []byte, p pool) (CodeAttr, error) {
	rd := newReader(bytes.NewReader(data))
	maxStack := rd.u2()
	maxLocals := rd.u2()
	codeLen := rd.u4()
	code := rd.bytes(codeLen)

	excCount := rd.u2()
	var handlers []ExceptionHandler
	for i := uint16(0); i < excCount && rd.err == nil; i++ {
		startPC := rd.u2()
		endPC := rd.u2()
		handlerPC := rd.u2()
		catchTypeIdx := rd.u2()
		catchType := ""
		if catchTypeIdx != 0 {
			var err error
			catchType, err = p.className(catchTypeIdx)
			if err != nil {
				return CodeAttr{}, fmt.Errorf("exception_table[%d] catch_type: %w", i, err)
			}
		}
		handlers = append(handlers, ExceptionHandler{
			StartPC:   int(startPC),
			EndPC:     int(endPC),
			HandlerPC: int(handlerPC),
			CatchType: catchType,
		})
	}

	var lineNumbers []LineNumberEntry
	attrCount := rd.u2()
	for i := uint16(0); i < attrCount && rd.err == nil; i++ {
		nameIdx := rd.u2()
		length := rd.u4()
		attrData := rd.bytes(length)
		if rd.err != nil {
			break
		}
		name, err := p.utf8(nameIdx)
		if err != nil {
			return CodeAttr{}, fmt.Errorf("Code sub-attribute %d name: %w", i, err)
		}
		if name == "LineNumberTable" {
			lineNumbers, err = parseLineNumberTable(attrData)
			if err != nil {
				return CodeAttr{}, fmt.Errorf("LineNumberTable: %w", err)
			}
		}
	}
	if rd.err != nil {
		return CodeAttr{}, rd.err
	}
	return CodeAttr{
		MaxStack:       int(maxStack),
		MaxLocals:      int(maxLocals),
		Code:           code,
		ExceptionTable: handlers,
		LineNumbers:    lineNumbers,
	}, nil
}

func parseLineNumberTable(data []byte) ([]LineNumberEntry, error) {
	rd := newReader(bytes.NewReader(data))
	count := rd.u2()
	entries := make([]LineNumberEntry, 0, count)
	for i := uint16(0); i < count && rd.err == nil; i++ {
		pc := rd.u2()
		line := rd.u2()
		entries = append(entries, LineNumberEntry{PC: int(pc), Line: int(line)})
	}
	if rd.err != nil {
		return nil, rd.err
	}
	return entries, nil
}
