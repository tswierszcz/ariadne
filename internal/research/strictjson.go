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

// requireMembers reports the first of names missing from a JSON object:
// required nullable fields must be present (CONTRACT.md §2), which a Go
// pointer field alone cannot tell from an absent one.
func requireMembers(data []byte, names ...string) error {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	for _, n := range names {
		if _, ok := m[n]; !ok {
			return fmt.Errorf("missing required field %q", n)
		}
	}
	return nil
}

// scanObject validates a one-line record like checkDuplicates and returns
// its top-level member names and its record_type string.
func scanObject(data []byte) (map[string]bool, string, error) {
	if !utf8.Valid(data) {
		return nil, "", fmt.Errorf("record is not valid UTF-8")
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	tok, err := dec.Token()
	if err != nil {
		return nil, "", err
	}
	if d, ok := tok.(json.Delim); !ok || d != '{' {
		return nil, "", fmt.Errorf("record is not a JSON object")
	}
	top := map[string]bool{}
	typ := ""
	for dec.More() {
		kt, err := dec.Token()
		if err != nil {
			return nil, "", err
		}
		k := kt.(string)
		if top[k] {
			return nil, "", fmt.Errorf("duplicate member %q at /", k)
		}
		top[k] = true
		if k == "record_type" {
			vt, err := dec.Token()
			if err != nil {
				return nil, "", err
			}
			s, ok := vt.(string)
			if !ok {
				return nil, "", fmt.Errorf("record_type is not a string")
			}
			typ = s
			continue
		}
		if err := walkValue(dec, "/"+k); err != nil {
			return nil, "", err
		}
	}
	if _, err := dec.Token(); err != nil {
		return nil, "", err
	}
	if _, err := dec.Token(); err != io.EOF {
		return nil, "", fmt.Errorf("trailing data after the record")
	}
	return top, typ, nil
}
