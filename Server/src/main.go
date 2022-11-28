package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"
)

const ClusterSize = 4

var data [ClusterSize]sync.Map

func startServer() {
	router := mux.NewRouter()

	router.HandleFunc("/data/{key}/{server}", getData).Methods(http.MethodGet)
	router.HandleFunc("/data", storeData).Methods(http.MethodPost)
	router.HandleFunc("/data/{key}/{server}", removeData).Methods(http.MethodDelete)
	router.HandleFunc("/data", updateData).Methods(http.MethodPut)

	router.HandleFunc("/recover", setRecovered).Methods(http.MethodPost)
	router.HandleFunc("/recover/{replica}", recoverServer).Methods(http.MethodGet)

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

	fmt.Println("Partition Leader started")
	//fmt.Println("Partition leader: http://localhost:80" + strconv.Itoa(ID+1) + "0")
	if err := http.ListenAndServe(":8000", router); err != nil {
		log.Fatal(err)
	}
}

func startUDPServer() {

	udpAddr, _ := net.ResolveUDPAddr("udp4", "239.0.0.0:5000")

	// setup listener for incoming UDP connection
	ln, _ := net.ListenMulticastUDP("udp4", nil, udpAddr)

	//fmt.Println("UDP server up and listening on port 8000")

	defer ln.Close()

	for {
		// wait for UDP client to connect
		getHeartBeat(ln)
	}
}

func startTCPServer() {
	// Listen for incoming connections.
	l, _ := net.Listen("tcp", "server"+strconv.Itoa(ID)+":"+"3000")

	// Close the listener when the application closes.
	defer l.Close()

	for {
		// Listen for an incoming connection.
		conn, _ := l.Accept()

		// Handle connections in a new goroutine.
		go postVictory(conn)
	}
}

func recoverMe() {
	req, _ := http.NewRequest(http.MethodPost, "http://proxyserver:7000/recovering", nil)
	http.DefaultClient.Do(req)

	for i := 0; i < replicaFactor; i++ {
		for j := 0; j < replicaFactor; j++ {
			if i != j {
				resp, err := http.Get("http://server" + strconv.Itoa(mod(ID-i+j, ClusterSize)) + ":8001/recover/" + strconv.Itoa(mod(ID-i, ClusterSize)))

				if err == nil && resp.StatusCode == http.StatusOK {

					go func() {
						m1 := make(map[string]string)

						json.NewDecoder(resp.Body).Decode(&m1)

						for s, s2 := range m1 {
							data[mod(ID-i, ClusterSize)].Store(s, s2)
						}
					}()

					continue
				}
			}
		}
	}

	recovered = true

	req1, _ := http.NewRequest(http.MethodDelete, "http://proxyserver:7000/recovering", nil)
	http.DefaultClient.Do(req1)

	fmt.Println("Server", ID, "successfully recovered")
}

func main() {

	go startServer()
	go startUDPServer()
	go startTCPServer()
	go startLeaderServer()

	//Await all servers to wake up and be ready for Election
	time.Sleep(time.Second)

	go Election()

	if recovered == false {
		go recoverMe()
	}

	select {}
}
