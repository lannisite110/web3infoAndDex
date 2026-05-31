package config

import (
	"fmt"
	"net/url"
	"strings"
)

// NormalizeMySQLDSN converts Railway-style mysql:// URLs to go-sql-driver DSN.
func NormalizeMySQLDSN(dsn string) (string, error) {
	dsn = strings.TrimSpace(dsn)
	if dsn == "" {
		return "", fmt.Errorf("empty MYSQL_DSN")
	}
	if !strings.HasPrefix(strings.ToLower(dsn), "mysql://") {
		if !strings.Contains(dsn, "parseTime=") {
			sep := "?"
			if strings.Contains(dsn, "?") {
				sep = "&"
			}
			dsn += sep + "parseTime=true"
		}
		return dsn, nil
	}

	u, err := url.Parse(dsn)
	if err != nil {
		return "", fmt.Errorf("parse MYSQL_DSN: %w", err)
	}

	user := ""
	pass := ""
	if u.User != nil {
		user = u.User.Username()
		pass, _ = u.User.Password()
	}
	host := u.Hostname()
	port := u.Port()
	if port == "" {
		port = "3306"
	}
	dbName := strings.TrimPrefix(u.Path, "/")
	if dbName == "" {
		dbName = "railway"
	}

	q := u.Query()
	q.Set("parseTime", "true")
	if q.Get("tls") == "" {
		// Railway public proxy certs often do not match *.rlwy.net hostnames.
		if strings.Contains(host, "rlwy.net") || strings.Contains(host, "railway.app") {
			q.Set("tls", "skip-verify")
		} else {
			q.Set("tls", "true")
		}
	}

	out := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?%s", user, pass, host, port, dbName, q.Encode())
	return out, nil
}
