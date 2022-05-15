package kv

import (
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/srcfoundry/kinesis/component"
)

type Value interface{}

type Bucket struct {
	component.SimpleComponent
	Created string `json:"created"`
	kvMap   map[string]Value
}

func (b *Bucket) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// proceeding to dynamically match and extract key path variable from URI.
	pathPrefix := b.GetURI() + "/{key:[a-zA-Z0-9_.-]+}"
	route := new(mux.Route).PathPrefix(pathPrefix)

	var match mux.RouteMatch
	isKeyPathMatched := route.Match(r, &match)
	if isKeyPathMatched && len(match.Vars) <= 0 {
		http.Error(w, "unable to derive key as path variable", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodPost:
		key := match.Vars["key"]
		var found bool

		// if key already found in bucket then return
		if _, found = b.kvMap[key]; found {
			http.Error(w, "cannot add duplicate key: "+key, http.StatusConflict)
			return
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Println("obtained error while reading POST body", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		b.kvMap[key] = body
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"status":"added"}"`))
	case http.MethodGet:
		// get component copy for obtaining etag
		bCopy, _ := component.GetComponentCopy(b)

		// if URI does not contain any key path variable, then return marshalled bucket component
		if !isKeyPathMatched {
			component.MarshallToHttpResponseWriter(w, bCopy)
			return
		}

		key := match.Vars["key"]
		var (
			found bool
			value interface{}
		)

		// if key not found in bucket then return
		if value, found = b.kvMap[key]; !found {
			http.Error(w, "cannot find value for key: "+key, http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("ETag", bCopy.GetEtag())
		w.Write(value.([]byte))
	default:
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}
