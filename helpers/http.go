package helpers

import (
	"net/http"
	"strings"

	"github.com/pkg/errors"
)

// GetFileNameFromHTTPHeader returns the first filename found in content-disposition header.
func GetFileNameFromHTTPHeader(headers http.Header) (string, error) {
	contentDispositionHeader := headers.Get("content-disposition")
	for _, stmt := range strings.Split(contentDispositionHeader, ";") {
		stmt = strings.TrimSpace(stmt)
		const filenamePrefix = "filename="
		if strings.HasPrefix(stmt, filenamePrefix) {
			return strings.Trim(stmt[len(filenamePrefix):], `"`), nil
		}
	}
	return "", errors.Errorf(`could not resolve filename from content-disposition header<%s>`, contentDispositionHeader)
}
