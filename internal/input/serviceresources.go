package input

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
)

// serviceFilePrefix is where a Maven-ecosystem jar's declarative
// ServiceLoader provider-configuration files live (JAR File
// Specification / java.util.ServiceLoader's own "Deploying service
// providers on the class path" documentation). The file's own name,
// after this prefix, is already the service interface's fully qualified
// binary name in dot form — unlike a .class entry, it is never
// slash-separated.
const serviceFilePrefix = "META-INF/services/"

// ServiceFileResource is one META-INF/services/<Interface> resource read
// from a jar: SERVICE_LOADING S0's "resource index" — a generic
// non-.class artifact, not yet parsed or interpreted. Content and
// SHA256 are kept so the parser (internal/serviceload) and any report
// can cite the exact source, and so two artifacts shipping the same
// file can be told apart from two that merely share an interface name.
type ServiceFileResource struct {
	Jar       string // the jar this was read from
	Interface string // the file's own name, e.g. "java.sql.Driver"
	Content   []byte
	SHA256    string
}

// ListServiceFiles reads every META-INF/services/<Interface> entry from
// jarPath (the directory entry itself, if the zip has one, is skipped).
// It does not parse or validate content — just resource access, mirroring
// inspectJar's own style for pom.properties.
func ListServiceFiles(jarPath string) ([]ServiceFileResource, error) {
	zr, err := zip.OpenReader(jarPath)
	if err != nil {
		return nil, fmt.Errorf("opening %s as a zip: %w", jarPath, err)
	}
	defer zr.Close()

	var out []ServiceFileResource
	for _, f := range zr.File {
		if !strings.HasPrefix(f.Name, serviceFilePrefix) {
			continue
		}
		iface := strings.TrimPrefix(f.Name, serviceFilePrefix)
		if iface == "" || strings.HasSuffix(f.Name, "/") {
			continue // the META-INF/services/ directory entry itself
		}
		rc, err := f.Open()
		if err != nil {
			return nil, fmt.Errorf("opening %s in %s: %w", f.Name, jarPath, err)
		}
		content, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			return nil, fmt.Errorf("reading %s in %s: %w", f.Name, jarPath, err)
		}
		sum := sha256.Sum256(content)
		out = append(out, ServiceFileResource{
			Jar:       jarPath,
			Interface: iface,
			Content:   content,
			SHA256:    hex.EncodeToString(sum[:]),
		})
	}
	return out, nil
}
