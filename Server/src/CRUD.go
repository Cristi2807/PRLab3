package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"net/http"
)

func storeData(w http.ResponseWriter, r *http.Request) {

	body, _ := io.ReadAll(r.Body)
	defer r.Body.Close()

	type Cell struct {
		Key   any `json:"key"`
		Value any `json:"value"`
	}

	var cell Cell

	r.Body = io.NopCloser(bytes.NewReader(body))
	json.NewDecoder(r.Body).Decode(&cell)

	_, ok := data.Load(cell.Key)

	if ok == true {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	data.Store(cell.Key, cell.Value)
	fmt.Println("Data stored successfully")

	w.WriteHeader(http.StatusCreated)
}

func updateData(w http.ResponseWriter, r *http.Request) {

	body, _ := io.ReadAll(r.Body)
	defer r.Body.Close()

	type Cell struct {
		Key   any `json:"key"`
		Value any `json:"value"`
	}

	var cell Cell

	r.Body = io.NopCloser(bytes.NewReader(body))
	json.NewDecoder(r.Body).Decode(&cell)

	_, ok := data.Load(cell.Key)

	if ok == false {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	data.Store(cell.Key, cell.Value)
	fmt.Println("Data updated")

	w.WriteHeader(http.StatusOK)
}

func removeData(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)

	_, ok := data.Load(params["key"])

	if ok == false {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	data.Delete(params["key"])
	fmt.Println("Data deleted")

	w.WriteHeader(http.StatusOK)
}

func getData(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)

	value, ok := data.Load(params["key"])

	if ok == false {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	dataMarshalled, _ := json.Marshal(
		struct {
			Key   any `json:"key"`
			Value any `json:"value"`
		}{
			params["key"], value,
		},
	)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(dataMarshalled)
}
