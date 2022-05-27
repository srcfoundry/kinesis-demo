package kv

import (
	"context"
	"fmt"
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
	// even though KVMAp is an exported field, it is excluded from showing up (`json:"-"` ensures its ignored) for GET on the bucket URI. The reason for exporting
	// is to include it in calculating hash of the bucket and setting its ETag.
	KVMap map[string]Value `json:"-"`
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

		// if key already found in bucket then return
		if _, found := b.KVMap[key]; found {
			http.Error(w, "cannot add duplicate key: "+key, http.StatusConflict)
			return
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Println("obtained error while reading http body", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		b.KVMap[key] = body
		log.Println("successfully added", key, "within", b.GetName())

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"status":"added"}"`))

	case http.MethodPut:
		key := match.Vars["key"]

		// if key already found in bucket then check if ETag matches the bucket Etag
		if _, found := b.KVMap[key]; found {
			etags := []string{}
			if moreEtags, found := r.Header[http.CanonicalHeaderKey("ETag")]; found {
				etags = append(etags, moreEtags...)
			}
			if moreEtags, found := r.Header[http.CanonicalHeaderKey("etag")]; found {
				etags = append(etags, moreEtags...)
			}

			if len(etags) <= 0 {
				http.Error(w, "missing ETag header", http.StatusBadRequest)
				return
			}

			// just iterate to check if any of etag values match since doing binary slice search could be overkill
			foundEtag := false
			for _, etag := range etags {
				if etag == b.GetEtag() {
					foundEtag = true
					break
				}
			}

			if !foundEtag {
				http.Error(w, "cannot update due to mismatched ETag", http.StatusConflict)
				return
			}
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Println("obtained error while reading http body", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		b.KVMap[key] = body
		log.Println("successfully updated", key, "within", b.GetName())

		// SetComponentEtag for bucket
		component.SetComponentEtag(b)

		w.Header().Set("ETag", b.GetEtag())
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"status":"updated"}"`))
		return

	case http.MethodGet:
		// SetComponentEtag for bucket
		component.SetComponentEtag(b)

		// if URI does not contain any key path variable, then return marshalled bucket component
		if !isKeyPathMatched {
			component.MarshallToHttpResponseWriter(w, b)
			return
		}

		key := match.Vars["key"]
		value, found := b.KVMap[key]

		// if key not found in bucket then return
		if !found {
			http.Error(w, "cannot find value for key: "+key, http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("ETag", b.GetEtag())
		w.Write(value.([]byte))

	case http.MethodDelete:
		// if URI does not contain any key path variable, then return
		if !isKeyPathMatched {
			http.Error(w, "cannot find URI", http.StatusNotFound)
			return
		}

		key := match.Vars["key"]
		_, found := b.KVMap[key]

		// if key not found in bucket then return
		if !found {
			http.Error(w, "cannot find value for key: "+key, http.StatusNotFound)
			return
		}

		delete(b.KVMap, key)
		log.Println("successfully deleted", key, "within", b.GetName())

		// SetComponentEtag for bucket
		component.SetComponentEtag(b)

		if len(b.KVMap) <= 0 {
			log.Println("no more keys within", b.GetName(), "... proceeding to shutdown")
			errCh := make(chan error)
			b.Notify(func() (context.Context, interface{}, chan<- error) {
				return context.TODO(), component.Shutdown, errCh
			})
			err := <-errCh
			if err != nil {
				log.Println(fmt.Sprintf("obtained error %s while shutting down %v", err, b.GetName()))
			}
		} else {
			w.Header().Set("ETag", b.GetEtag())
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"deleted"}"`))
		return

	default:
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}
