package kv

import (
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/srcfoundry/kinesis/component"
)

type KV struct {
	component.Container
}

func (k *KV) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	/*
		Proceeding to dynamically match and extract bucket & key path variables from URI. Hint obtained from
		https://github.com/gorilla/mux/blob/master/mux_test.go#L228.
		POST & PUT request creates a new bucket and key or could be used to replace an existing key, in which case
		should pass the bucket ETag as header. All updates (PUT) to existing keys are handled within the bucket component.
		Expected POST/PUT URI path would be "/kv/<bucket>/<key>"
	*/
	case http.MethodPost, http.MethodPut:
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
			log.Println("obtained error while reading http body", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		bucket := Bucket{Created: time.Now().Format("May 02, 2006 15:04:05"), KVMap: map[string]Value{key: body}}
		bucket.Name = bucketName
		bucket.RWMutex = &sync.RWMutex{}

		notifyCh := make(chan interface{}, 1)
		defer close(notifyCh)
		bucket.Subscribe(bucket.Name+".activate.listener", notifyCh)

		err = k.Add(&bucket)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

	listenerloop:
		for notification := range notifyCh {
			switch notification {
			case component.Started:
				bucket.Unsubscribe(bucket.Name + ".activate.listener")
				break listenerloop
			}
		}
		log.Println("successfully created bucket", bucket.GetName())
		log.Println("successfully added", key, "within", bucket.GetName())

		w.Header().Add("ETag", bucket.GetEtag())
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"status":"created"}"`))
	case http.MethodGet:
		// route request to the embedded SimpleComponent ServeHTTP method
		k.SimpleComponent.ServeHTTP(w, r)
	default:
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}
