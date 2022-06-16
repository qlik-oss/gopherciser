package session

import (
	"fmt"
	"net/http"
	"sync"
)

// HeaderJar map between hosts and headers
type HeaderJar struct {
	headers        *sync.Map
	allRestMethods []RestMethod
}

// hostMethodPair acts as a key in the map
type hostMethodPair struct {
	host   string
	method RestMethod
}

func (hmp *hostMethodPair) string() string {
	return fmt.Sprintf("%s:%s", hmp.host, hmp.method.String())
}

// NewHeaderJar returns an empty HeaderJar
func NewHeaderJar() *HeaderJar {
	restMethodInts := RestMethod(0).GetEnumMap().AsInt()
	allRestMethods := make([]RestMethod, 0, len(restMethodInts))
	for _, restMethod := range restMethodInts {
		allRestMethods = append(allRestMethods, RestMethod(restMethod))
	}

	return &HeaderJar{
		headers:        &sync.Map{},
		allRestMethods: allRestMethods,
	}
}

// SetHeader set header for a particular host
func (hj *HeaderJar) SetHeader(host string, header http.Header) {
	if len(hj.allRestMethods) == 0 {
		panic("HeaderJar may only be initialized using NewHeaderJar()")
	}
	hj.SetHeaderForMethods(host, header, hj.allRestMethods)
}

// SetHeaderForMethods set header for a particular host and set of methods
func (hj *HeaderJar) SetHeaderForMethods(host string, header http.Header, methods []RestMethod) {
	for _, method := range methods {
		hmp := hostMethodPair{host, method}
		hj.headers.Store(hmp, header)
	}
}

// GetHeader returns the header for the given host
func (hj *HeaderJar) GetHeader(host string) http.Header {
	return hj.GetHeaderForMethod(host, GET)
}

// GetHeaderForMethod returns the header for the given host and specific method
func (hj *HeaderJar) GetHeaderForMethod(host string, method RestMethod) http.Header {
	hmp := hostMethodPair{host, method}
	value, found := hj.headers.Load(hmp)
	if !found {
		return nil
	}
	return value.(http.Header)
}
