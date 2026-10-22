package classfile

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
)

// JarClassReader re-reads individual class entries from one or more
// jars, keeping each jar's own *zip.ReadCloser open across repeated
// calls. Several consumers (SERVICE_LOADING S1, ASYNC_CALLBACKS A1) need
// a specific class's real *ClassFile — constant pool and all — after
// hierarchy.Build has already discarded it (see hierarchy.Method's own
// doc comment for why): this is the shared way to get it back without
// each consumer re-opening the same jar once per class it cares about.
// The zero value is ready to use; call Close when done.
type JarClassReader struct {
	open map[string]*zip.ReadCloser
}

// ReadClass parses the class at entryName inside the jar at jarPath.
func (r *JarClassReader) ReadClass(jarPath, entryName string) (*ClassFile, error) {
	if r.open == nil {
		r.open = map[string]*zip.ReadCloser{}
	}
	zr, ok := r.open[jarPath]
	if !ok {
		var err error
		zr, err = zip.OpenReader(jarPath)
		if err != nil {
			return nil, err
		}
		r.open[jarPath] = zr
	}
	for _, f := range zr.File {
		if f.Name != entryName {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return nil, err
		}
		data, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			return nil, err
		}
		return Parse(bytes.NewReader(data))
	}
	return nil, fmt.Errorf("entry %q not found in %s", entryName, jarPath)
}

// Close closes every jar this reader opened. Safe to call on a zero
// value that never read anything.
func (r *JarClassReader) Close() error {
	var firstErr error
	for _, zr := range r.open {
		if err := zr.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	r.open = nil
	return firstErr
}
