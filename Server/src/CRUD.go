package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"net/http"
	"strconv"
)

var recovered = false

func storeData(w http.ResponseWriter, r *http.Request) {

	body, _ := io.ReadAll(r.Body)
	defer r.Body.Close()

	type Cell struct {
		Key    any `json:"key"`
		Value  any `json:"value"`
		Server int `json:"server"`
	}

	var cell Cell

	r.Body = io.NopCloser(bytes.NewReader(body))
	json.NewDecoder(r.Body).Decode(&cell)

	_, ok := data[cell.Server].Load(cell.Key)

	if ok == true {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	data[cell.Server].Store(cell.Key, cell.Value)
	fmt.Println("Data", cell, "stored successfully")

	w.WriteHeader(http.StatusCreated)
}

func updateData(w http.ResponseWriter, r *http.Request) {

	body, _ := io.ReadAll(r.Body)
	defer r.Body.Close()

	type Cell struct {
		Key    any `json:"key"`
		Value  any `json:"value"`
		Server int `json:"server"`
	}

	var cell Cell

	r.Body = io.NopCloser(bytes.NewReader(body))
	json.NewDecoder(r.Body).Decode(&cell)

	_, ok := data[cell.Server].Load(cell.Key)

	if ok == false {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	data[cell.Server].Store(cell.Key, cell.Value)
	fmt.Println("Data", cell, "updated")

	w.WriteHeader(http.StatusOK)
}

func removeData(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)

	srvNumber, _ := strconv.Atoi(params["server"])

	_, ok := data[srvNumber].Load(params["key"])

	if ok == false {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	data[srvNumber].Delete(params["key"])
	fmt.Println("Data deleted")

	w.WriteHeader(http.StatusOK)
}

func getData(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)

	srvNumber, _ := strconv.Atoi(params["server"])

	value, ok := data[srvNumber].Load(params["key"])

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

func setRecovered(w http.ResponseWriter, r *http.Request) {
	recovered = true
}

func recoverServer(w http.ResponseWriter, r *http.Request) {
	if recovered == true {

		params := mux.Vars(r)

		replica, _ := strconv.Atoi(params["replica"])

		var m = make(map[string]string)

		data[replica].Range(func(k any, v any) bool {

			k1 := k.(string)
			v1 := v.(string)

			m[k1] = v1

			return true
		})

		dataMarshalled, _ := json.Marshal(m)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(dataMarshalled)

	} else {
		w.WriteHeader(http.StatusExpectationFailed)
	}
}
