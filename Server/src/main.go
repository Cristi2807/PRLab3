package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
)

var ID, _ = strconv.Atoi(os.Getenv("ID"))
var data sync.Map

type Cluster []struct {
	ID  int    `json:"id"`
	URL string `json:"url"`
}

var cluster Cluster
var leader = -1
var ch = make(chan int, 1)

func ParseCluster() {
	serversFile, _ := os.Open("servers.json")
	jsonParser := json.NewDecoder(serversFile)
	jsonParser.Decode(&cluster)
}

func sendToAll(method string, body []byte, key string) {

	for i := 0; i < len(cluster); i++ {
		if i != ID {
			switch method {
			case http.MethodPost:
				rbody := io.NopCloser(bytes.NewReader(body))
				req, _ := http.NewRequest(http.MethodPost, cluster[i].URL+"/datastore", rbody)
				http.DefaultClient.Do(req)
			case http.MethodPut:
				rbody := io.NopCloser(bytes.NewReader(body))
				req, _ := http.NewRequest(http.MethodPut, cluster[i].URL+"/datastore", rbody)
				http.DefaultClient.Do(req)
			case http.MethodDelete:
				req, _ := http.NewRequest(http.MethodDelete, cluster[i].URL+"/datastore/"+key, nil)
				http.DefaultClient.Do(req)
			}

		}
	}

}

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
	} else {
		data.Store(cell.Key, cell.Value)
		fmt.Println("Data stored successfully")

		w.WriteHeader(http.StatusCreated)

		if leader == ID {
			sendToAll(http.MethodPost, body, "")
		}

	}
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

	if ok == true {
		data.Store(cell.Key, cell.Value)
		w.WriteHeader(http.StatusOK)
		fmt.Println("Data updated")

		if leader == ID {
			sendToAll(http.MethodPut, body, "")
		}

	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func removeData(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)

	_, ok := data.Load(params["key"])

	if ok == true {
		data.Delete(params["key"])
		fmt.Println("Data deleted")

		w.WriteHeader(http.StatusOK)

		if leader == ID {
			sendToAll(http.MethodDelete, nil, params["key"])
		}

	} else {
		w.WriteHeader(http.StatusNotFound)
	}

}

func getData(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)

	value, ok := data.Load(params["key"])

	if ok == true {

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
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func Election() {

	for i := len(cluster) - 1; i >= 0; i-- {
		_, err := http.Get(cluster[i].URL + "/heartbeat")

		if err == nil {
			for j := 0; j <= i; j++ {
				leaderMarshalled, _ := json.Marshal(i)
				rBody := bytes.NewBuffer(leaderMarshalled)
				http.Post(cluster[j].URL+"/victory", "application/json", rBody)
			}

			break
		}
	}

}

func getHeartbeat(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func postVictory(w http.ResponseWriter, r *http.Request) {
	var i int
	json.NewDecoder(r.Body).Decode(&i)
	leader = i

	if leader == ID {
		fmt.Println("Partition leader:", cluster[leader].URL)
	}
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/datastore/{key}", getData).Methods(http.MethodGet)
	router.HandleFunc("/datastore", storeData).Methods(http.MethodPost)
	router.HandleFunc("/datastore/{key}", removeData).Methods(http.MethodDelete)
	router.HandleFunc("/datastore", updateData).Methods(http.MethodPut)

	router.HandleFunc("/victory", postVictory).Methods(http.MethodPost)
	router.HandleFunc("/heartbeat", getHeartbeat).Methods(http.MethodGet)

	ParseCluster()

	go Election()

	fmt.Printf("Server started\n")
	if err := http.ListenAndServe(os.Getenv("ADDRESS"), router); err != nil {
		log.Fatal(err)
	}

}
