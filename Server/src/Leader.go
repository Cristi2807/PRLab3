package main

import (
	"bytes"
	"encoding/json"
	"github.com/gorilla/mux"
	"io"
	"net/http"
	"strconv"
)

var srvDuplicate = ClusterSize/2 + 1
var saveIndex int

func storeDataLeader(w http.ResponseWriter, r *http.Request) {

	body, _ := io.ReadAll(r.Body)
	defer r.Body.Close()

	type Cell struct {
		Key   string `json:"key"`
		Value any    `json:"value"`
	}

	var cell Cell

	r.Body = io.NopCloser(bytes.NewReader(body))
	json.NewDecoder(r.Body).Decode(&cell)

	if len(getIndex(cell.Key)) > 0 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	for i := saveIndex; i < saveIndex+srvDuplicate; i++ {
		rbody := io.NopCloser(bytes.NewReader(body))
		req, _ := http.NewRequest(http.MethodPost, "http://server"+strconv.Itoa(mod(i, ClusterSize))+":8001/data", rbody)
		http.DefaultClient.Do(req)
	}

	saveIndex++
	w.WriteHeader(http.StatusCreated)

}

func updateDataLeader(w http.ResponseWriter, r *http.Request) {

	body, _ := io.ReadAll(r.Body)
	defer r.Body.Close()

	type Cell struct {
		Key   string `json:"key"`
		Value any    `json:"value"`
	}

	var cell Cell

	r.Body = io.NopCloser(bytes.NewReader(body))
	json.NewDecoder(r.Body).Decode(&cell)

	index := getIndex(cell.Key)

	if len(index) == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	for _, v := range index {
		rbody := io.NopCloser(bytes.NewReader(body))
		req, _ := http.NewRequest(http.MethodPut, "http://server"+strconv.Itoa(v)+":8001/data", rbody)
		http.DefaultClient.Do(req)
	}

	w.WriteHeader(http.StatusOK)

}

func removeDataLeader(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)

	index := getIndex(params["key"])

	if len(index) == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	for _, v := range index {
		req, _ := http.NewRequest(http.MethodDelete, "http://server"+strconv.Itoa(v)+":8001/data/"+params["key"], nil)
		http.DefaultClient.Do(req)
	}

	w.WriteHeader(http.StatusOK)

}

func getDataLeader(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)

	index := getIndex(params["key"])

	if len(index) == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	resp, _ := http.Get("http://server" + strconv.Itoa(index[0]) + ":8001/data/" + params["key"])

	body, _ := io.ReadAll(resp.Body)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(body)

}

func getIndex(key string) []int {
	var a []int

	for i := 0; i < ClusterSize; i++ {
		resp, err := http.Get("http://server" + strconv.Itoa(i) + ":8001/data/" + key)

		if err == nil && resp.StatusCode == http.StatusOK {
			a = append(a, i)
		}
	}
	return a
}

func mod(a int, b int) int {
	return (b + (a % b)) % b
}
