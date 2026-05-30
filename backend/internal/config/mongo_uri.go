package config

import (
	"fmt"
	"strings"
)

// ValidateMongoURI checks Atlas connection strings before dial.
func ValidateMongoURI(uri string) error {
	if !strings.HasPrefix(uri, "mongodb://") && !strings.HasPrefix(uri, "mongodb+srv://") {
		return fmt.Errorf("MONGODB_URI must start with mongodb:// or mongodb+srv://")
	}
	lower := strings.ToLower(uri)
	if strings.Contains(lower, "<password>") {
		return fmt.Errorf("MONGODB_URI still contains Atlas placeholder <password>")
	}
	if strings.Count(uri, "@") != 1 {
		return fmt.Errorf("MONGODB_URI appears malformed (expected exactly one @)")
	}
	return nil
}

// MongoURIRedacted returns a URI safe for logs (password hidden).
func MongoURIRedacted(uri string) string {
	schemeEnd := strings.Index(uri, "://")
	if schemeEnd < 0 {
		return "***"
	}
	rest := uri[schemeEnd+3:]
	at := strings.Index(rest, "@")
	if at < 0 {
		return uri
	}
	creds := rest[:at]
	host := rest[at:]
	colon := strings.Index(creds, ":")
	if colon < 0 {
		return uri[:schemeEnd+3] + creds + host
	}
	return uri[:schemeEnd+3] + creds[:colon+1] + "***" + host
}
