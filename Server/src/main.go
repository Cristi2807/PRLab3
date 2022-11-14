package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"sync"
	"time"
)

var ClusterSize = 4

var data sync.Map

func startServer() {
	router := mux.NewRouter()

	router.HandleFunc("/data/{key}", getData).Methods(http.MethodGet)
	router.HandleFunc("/data", storeData).Methods(http.MethodPost)
	router.HandleFunc("/data/{key}", removeData).Methods(http.MethodDelete)
	router.HandleFunc("/data", updateData).Methods(http.MethodPut)

	router.HandleFunc("/victory", postVictory).Methods(http.MethodPost)
	router.HandleFunc("/test", getTest).Methods(http.MethodGet)
	router.HandleFunc("/heartbeat", getHeartBeat).Methods(http.MethodGet)

	fmt.Println("Server started")
	if err := http.ListenAndServe(":8001", router); err != nil {
		log.Fatal(err)
	}

}

func startLeaderServer() {
	router := mux.NewRouter()
	router.HandleFunc("/datastore/{key}", getDataLeader).Methods(http.MethodGet)
	router.HandleFunc("/datastore", storeDataLeader).Methods(http.MethodPost)
	router.HandleFunc("/datastore/{key}", removeDataLeader).Methods(http.MethodDelete)
	router.HandleFunc("/datastore", updateDataLeader).Methods(http.MethodPut)

	saveIndex = 0

	fmt.Println("Partition Leader started")
	//fmt.Println("Partition leader: http://localhost:80" + strconv.Itoa(ID+1) + "0")
	if err := http.ListenAndServe(":8000", router); err != nil {
		log.Fatal(err)
	}
}

func main() {

	go startServer()

	time.Sleep(time.Second)

	go Election()

	select {}
}
