package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"log"
	"net/http"
	"strconv"
	"sync"
)

var ClusterSize = 4
var ActiveServers = 0
var leader = -1
var m sync.Mutex

func postVictory(w http.ResponseWriter, r *http.Request) {
	m.Lock()
	defer m.Unlock()

	var newLeader int
	json.NewDecoder(r.Body).Decode(&newLeader)

	if newLeader != leader {
		fmt.Println("Partition Leader SET")
	}

	leader = newLeader
}

func storeData(w http.ResponseWriter, r *http.Request) {

	if ActiveServers < ClusterSize {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	body, _ := io.ReadAll(r.Body)
	defer r.Body.Close()

	rbody := io.NopCloser(bytes.NewReader(body))
	req, _ := http.NewRequest(http.MethodPost, "http://server"+strconv.Itoa(leader)+":8000/datastore", rbody)
	resp, _ := http.DefaultClient.Do(req)

	w.WriteHeader(resp.StatusCode)

}

func updateData(w http.ResponseWriter, r *http.Request) {

	if ActiveServers < ClusterSize/2 {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	body, _ := io.ReadAll(r.Body)
	defer r.Body.Close()

	rbody := io.NopCloser(bytes.NewReader(body))
	req, _ := http.NewRequest(http.MethodPut, "http://server"+strconv.Itoa(leader)+":8000/datastore", rbody)
	resp, _ := http.DefaultClient.Do(req)

	w.WriteHeader(resp.StatusCode)

}

func removeData(w http.ResponseWriter, r *http.Request) {

	if ActiveServers < ClusterSize/2 {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	params := mux.Vars(r)

	req, _ := http.NewRequest(http.MethodDelete, "http://server"+strconv.Itoa(leader)+":8000/datastore/"+params["key"], nil)
	resp, _ := http.DefaultClient.Do(req)

	w.WriteHeader(resp.StatusCode)
}

func getData(w http.ResponseWriter, r *http.Request) {

	if ActiveServers < ClusterSize/2 {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	params := mux.Vars(r)

	resp, _ := http.Get("http://server" + strconv.Itoa(leader) + ":8000/datastore/" + params["key"])

	body, _ := io.ReadAll(resp.Body)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	w.Write(body)
}

func getActiveServers(w http.ResponseWriter, r *http.Request) {
	var updatedActiveSrvs int
	json.NewDecoder(r.Body).Decode(&updatedActiveSrvs)

	ActiveServers = updatedActiveSrvs
	//fmt.Println("Active servers:", ActiveServers)
}

func startServer() {
	router := mux.NewRouter()

	router.HandleFunc("/datastore/{key}", getData).Methods(http.MethodGet)
	router.HandleFunc("/datastore", storeData).Methods(http.MethodPost)
	router.HandleFunc("/datastore/{key}", removeData).Methods(http.MethodDelete)
	router.HandleFunc("/datastore", updateData).Methods(http.MethodPut)

	router.HandleFunc("/victory", postVictory).Methods(http.MethodPost)
	router.HandleFunc("/active", getActiveServers).Methods(http.MethodPost)

	fmt.Println("Proxy Server started")
	if err := http.ListenAndServe(":7000", router); err != nil {
		log.Fatal(err)
	}

}

func main() {

	go startServer()

	select {}
}
