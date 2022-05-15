package kv

import (
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/srcfoundry/kinesis/component"
)

type KV struct {
	component.Container
	buckets map[string]bool
}

func (k *KV) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:

		// proceeding to dynamically match and extract bucket & key path variables from URI. Hint obtained from
		// https://github.com/gorilla/mux/blob/master/mux_test.go#L228
		// expected POST URI path would be "/kv/<bucket>/<key>"
		postPath := k.GetURI() + "/{bucket:[a-zA-Z0-9_.-]+}/{key:[a-zA-Z0-9_.-]+}"
		route := new(mux.Route).PathPrefix(postPath)
		var match mux.RouteMatch

		isPathMatched := route.Match(r, &match)
		if !isPathMatched {
			http.Error(w, "URI path not matching: "+postPath, http.StatusBadRequest)
			return
		}

		if len(match.Vars) <= 0 {
			http.Error(w, "unable to derive bucket or key as path variables", http.StatusBadRequest)
			return
		}

		bucketName, key := match.Vars["bucket"], match.Vars["key"]
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Println("obtained error while reading POST", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		bucket := Bucket{Created: time.Now().Format("May 02, 2006 15:04:05"), kvMap: map[string]Value{key: body}}
		bucket.Name = bucketName

		err = k.Add(&bucket)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"status":"created"}"`))
	case http.MethodGet:
		kvComponent, _ := component.GetComponentCopy(k)
		component.MarshallToHttpResponseWriter(w, kvComponent)
	default:
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}
