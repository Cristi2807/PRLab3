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
	"time"
)

const ClusterSize = 4

var recovering []int

var leader = -1
var m sync.Mutex
var mutex sync.Mutex
var t time.Time

func postVictory(w http.ResponseWriter, r *http.Request) {
	m.Lock()
	defer m.Unlock()

	var newLeader int
	json.NewDecoder(r.Body).Decode(&newLeader)

	if newLeader != leader {
		fmt.Println("Partition Leader SET Server", newLeader, time.Now())
	}

	leader = newLeader
}

func storeData(w http.ResponseWriter, r *http.Request) {

	if len(recovering) == 0 {
		body, _ := io.ReadAll(r.Body)
		defer r.Body.Close()

		rbody := io.NopCloser(bytes.NewReader(body))
		req, _ := http.NewRequest(http.MethodPost, "http://server"+strconv.Itoa(leader)+":8000/datastore", rbody)
		resp, _ := http.DefaultClient.Do(req)

		w.WriteHeader(resp.StatusCode)

		return
	}

	w.WriteHeader(http.StatusInternalServerError)
}

func updateData(w http.ResponseWriter, r *http.Request) {

	if len(recovering) == 0 {
		body, _ := io.ReadAll(r.Body)
		defer r.Body.Close()

		rbody := io.NopCloser(bytes.NewReader(body))
		req, _ := http.NewRequest(http.MethodPut, "http://server"+strconv.Itoa(leader)+":8000/datastore", rbody)
		resp, _ := http.DefaultClient.Do(req)

		w.WriteHeader(resp.StatusCode)

		return
	}

	w.WriteHeader(http.StatusInternalServerError)
}

func removeData(w http.ResponseWriter, r *http.Request) {

	if len(recovering) == 0 {
		params := mux.Vars(r)

		req, _ := http.NewRequest(http.MethodDelete, "http://server"+strconv.Itoa(leader)+":8000/datastore/"+params["key"], nil)
		resp, _ := http.DefaultClient.Do(req)

		w.WriteHeader(resp.StatusCode)

		return
	}

	w.WriteHeader(http.StatusInternalServerError)
}

func getData(w http.ResponseWriter, r *http.Request) {

	if len(recovering) == 0 {
		params := mux.Vars(r)

		resp, _ := http.Get("http://server" + strconv.Itoa(leader) + ":8000/datastore/" + params["key"])

		body, _ := io.ReadAll(resp.Body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(resp.StatusCode)
		w.Write(body)

		return
	}

	w.WriteHeader(http.StatusInternalServerError)
}

func startServer() {
	router := mux.NewRouter()

	router.HandleFunc("/datastore/{key}", getData).Methods(http.MethodGet)
	router.HandleFunc("/datastore", storeData).Methods(http.MethodPost)
	router.HandleFunc("/datastore/{key}", removeData).Methods(http.MethodDelete)
	router.HandleFunc("/datastore", updateData).Methods(http.MethodPut)

	router.HandleFunc("/victory", postVictory).Methods(http.MethodPost)

	router.HandleFunc("/recovering", setRecovering).Methods(http.MethodPost)
	router.HandleFunc("/recovering", setRecovering).Methods(http.MethodDelete)

	fmt.Println("Proxy Server started")
	if err := http.ListenAndServe(":7000", router); err != nil {
		log.Fatal(err)
	}

}

func setRecovering(w http.ResponseWriter, r *http.Request) {
	mutex.Lock()
	defer mutex.Unlock()

	var moreThanZero = len(recovering) > 0
	var equalZero = len(recovering) == 0

	if r.Method == http.MethodPost {
		recovering = append(recovering, 1)

	}

	if r.Method == http.MethodDelete {
		recovering = recovering[1:]

	}

	var nowMoreThanZero = len(recovering) > 0
	var nowEqualZero = len(recovering) == 0

	if equalZero && nowMoreThanZero {
		t = time.Now()
		fmt.Println("STOPPING request handling! Recovering in process", time.Now())
	}

	if moreThanZero && nowEqualZero {
		fmt.Println("STARTING request handling! Recovering finished", time.Now())
		fmt.Println("Recovering lasted:", time.Now().Sub(t))
	}
}

func recoverAllServersOnStartUp() {
	for i := 0; i < ClusterSize; i++ {
		http.Post("http://server"+strconv.Itoa(i)+":8001/recover", "", nil)
	}
}

func main() {

	go startServer()

	time.Sleep(500 * time.Millisecond)

	go recoverAllServersOnStartUp()

	select {}
}

func simRequests() {
	for i := 0; i < 1000; i++ {
		type Cell struct {
			Key   any `json:"key"`
			Value any `json:"value"`
		}

		var cell Cell

		cell.Key = strconv.Itoa(i)
		cell.Value = "THE SAME VALUE"

		cellMarshalled, _ := json.Marshal(cell)
		rBody := bytes.NewBuffer(cellMarshalled)

		req, _ := http.NewRequest(http.MethodPost,
			"http://proxyserver:7000/datastore", rBody)

		http.DefaultClient.Do(req)
	}

}
