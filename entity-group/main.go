package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

func init() {
	r := mux.NewRouter()

	r.Path("/example").Methods(http.MethodPost).HandlerFunc(putSamples)

	fcRoute := r.PathPrefix("/foo/{fooID}/foochild")
	fcWithID := fcRoute.Path("/{fcID}")
	fcWithID.Methods(http.MethodGet).HandlerFunc(getFooChild)

	http.Handle("/", r)
}

type foo struct {
	Name string
}

type fooChild struct {
	Name string
}

func putSamples(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	var (
		f      = foo{Name: "SampleFoo"}
		fooKey = datastore.NewIncompleteKey(ctx, "foo", nil)
	)
	newFooKey, err := datastore.Put(ctx, fooKey, &f)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var (
		fc    = fooChild{Name: "SampleChile"}
		fcKey = datastore.NewIncompleteKey(ctx, "fooChild", newFooKey)
	)
	newFCKey, err := datastore.Put(ctx, fcKey, &fc)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Foo ID: %v\nFooChild ID: %v", newFooKey.IntID(), newFCKey.IntID())
}

// XXX EGを組んでるエンティティに対して、子エンティティのIDだけを指定してエンティティを取得することはできない
func getFooChildOnlyChildID(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	fcKeyID := mux.Vars(r)["id"]
	id, _ := strconv.ParseInt(fcKeyID, 10, 64)

	var (
		fc    = new(fooChild)
		fcKey = datastore.NewKey(ctx, "fooChild", "", id, nil)
	)
	if err := datastore.Get(ctx, fcKey, fc); err == datastore.ErrNoSuchEntity {
		http.Error(w, "Not found data", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf8")
	json.NewEncoder(w).Encode(fc)
}

func getFooChild(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	log.Infof(ctx, "NewContext; Type : %T, Values : %+v", ctx, ctx)
	log.Infof(ctx, "ReqContext; Type : %T, Values : %+v", r.Context(), r.Context())

	var (
		sFooID  = mux.Vars(r)["fooID"]
		sFcID   = mux.Vars(r)["fcID"]
		fID, _  = strconv.ParseInt(sFooID, 10, 64)
		fcID, _ = strconv.ParseInt(sFcID, 10, 64)
	)

	var (
		fc    = new(fooChild)
		fcKey = datastore.NewKey(ctx, "fooChild", "", fcID, datastore.NewKey(ctx, "foo", "", fID, nil))
	)
	if err := datastore.Get(ctx, fcKey, fc); err == datastore.ErrNoSuchEntity {
		http.Error(w, "Not found data", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp, err := json.MarshalIndent(fc, "", "  ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf8")
	fmt.Fprintf(w, "%s\n", resp)
}
