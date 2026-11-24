package research

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"unicode/utf8"
)

var (
	identifierRE = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_.:-]{0,159}$`)
	hashRE       = regexp.MustCompile(`^sha256:[a-f0-9]{64}$`)
)

// checkDuplicates rejects a JSON document with a repeated object member
// at any depth (CONTRACT.md §2: duplicates are rejected before schema
// validation), invalid UTF-8, a BOM or trailing data.
func checkDuplicates(data []byte) error {
	if bytes.HasPrefix(data, []byte{0xEF, 0xBB, 0xBF}) {
		return fmt.Errorf("document starts with a byte order mark")
	}
	if !utf8.Valid(data) {
		return fmt.Errorf("document is not valid UTF-8")
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	if err := walkValue(dec, ""); err != nil {
		return err
	}
	if _, err := dec.Token(); err != io.EOF {
		return fmt.Errorf("trailing data after the JSON document")
	}
	return nil
}

func walkValue(dec *json.Decoder, path string) error {
	tok, err := dec.Token()
	if err != nil {
		return err
	}
	d, ok := tok.(json.Delim)
	if !ok {
		return nil
	}
	switch d {
	case '{':
		seen := map[string]bool{}
		for dec.More() {
			kt, err := dec.Token()
			if err != nil {
				return err
			}
			k := kt.(string)
			if seen[k] {
				return fmt.Errorf("duplicate member %q at %s", k, pointerOrRoot(path))
			}
			seen[k] = true
			if err := walkValue(dec, path+"/"+k); err != nil {
				return err
			}
		}
	case '[':
		for i := 0; dec.More(); i++ {
			if err := walkValue(dec, fmt.Sprintf("%s/%d", path, i)); err != nil {
				return err
			}
		}
	}
	_, err = dec.Token() // closing delimiter
	return err
}

func pointerOrRoot(p string) string {
	if p == "" {
		return "/"
	}
	return p
}

// decodeStrict checks duplicates, then decodes data into v rejecting
// unknown fields.
func decodeStrict(data []byte, v any) error {
	if err := checkDuplicates(data); err != nil {
		return err
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	return dec.Decode(v)
}
