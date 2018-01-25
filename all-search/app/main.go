package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

type foo struct {
	Name        string   `json:"name"`
	Mail        string   `datastore:",noindex" json:"mail"`
	Description string   `datastore:",noindex" json:"description"`
	Index       []string `json:"-"`
}

func (f *foo) setIndex() {
	var (
		names = nGram(2, f.Name)
		mails = nGram(2, f.Mail)
		descs = nGram(2, f.Description)
	)
	f.Index = append(f.Index, names...)
	f.Index = append(f.Index, mails...)
	f.Index = append(f.Index, descs...)
}

func init() {
	http.HandleFunc("/put", putFoos)
	http.HandleFunc("/get", getFoos)
}

func putFoos(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	foos := []*foo{
		{Name: "テスト1", Mail: "foo1@sample.com", Description: "全文検索のテスト。"},
		{Name: "2番めのデータ", Mail: "foo2@sample.com", Description: "これが検索可能か？"},
		{Name: "名前", Mail: "foo3@sample.com", Description: "テストします"},
		{Name: "名前４マルチバイトにちゃんと対応できるか", Mail: "foo4@sample.com", Description: "あまり"},
	}
	for _, foo := range foos {
		foo.setIndex()
	}
	keys := []*datastore.Key{
		datastore.NewIncompleteKey(ctx, "foo", nil),
		datastore.NewIncompleteKey(ctx, "foo", nil),
		datastore.NewIncompleteKey(ctx, "foo", nil),
		datastore.NewIncompleteKey(ctx, "foo", nil),
	}

	if _, err := datastore.PutMulti(ctx, keys, foos); err != nil {
		log.Errorf(ctx, err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "put datas : %v", len(foos))
}

func getFoos(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	queries := nGram(2, r.FormValue("q"))
	query := datastore.NewQuery("foo")
	for _, q := range queries {
		query = query.Filter("Index=", q)
	}
	var foos []*foo
	if _, err := query.Order("Name").GetAll(ctx, &foos); err != nil {
		log.Errorf(ctx, err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp, err := json.MarshalIndent(foos, "", "  ")
	if err != nil {
		log.Errorf(ctx, err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(resp)
}

func nGram(n int, s string) []string {
	var (
		runes = []rune(s)
		grams = make([]string, 0, len(runes))
	)
	for left, right := 0, n; right <= len(runes); right++ {
		grams = append(grams, string(runes[left:right]))
		left++
	}
	return grams
}
