package session

import (
	"net/http"
	"sync"
)

// HeaderJar map between hosts and headers
type HeaderJar struct {
	headers *sync.Map
}

// NewHeaderJar returns an empty HeaderJar
func NewHeaderJar() *HeaderJar {
	return &HeaderJar{
		headers: &sync.Map{},
	}
}

// SetHeader set header for a particular host
func (hj *HeaderJar) SetHeader(host string, header http.Header) {
	hj.headers.Store(host, header)
}

// GetHeader returns the header for the given host
func (hj *HeaderJar) GetHeader(host string) http.Header {
	value, found := hj.headers.Load(host)
	if !found {
		return nil
	}
	return value.(http.Header)
}
