package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

var ID, _ = strconv.Atoi(os.Getenv("ID"))

var leader = -1
var ActiveServers = 1

var leaderBeat = make(chan bool)
var followerBeat = make(chan bool)

var m sync.Mutex

var listening = false

func Election() {

	for i := ClusterSize - 1; i >= 0; i-- {
		_, err := http.Get("http://server" + strconv.Itoa(i) + ":8001/test")

		if err == nil {
			//fmt.Println("Leader", i, time.Now())
			for j := 0; j <= i; j++ {
				leaderMarshalled, _ := json.Marshal(i)
				rBody := bytes.NewBuffer(leaderMarshalled)
				http.Post("http://server"+strconv.Itoa(j)+":8001/victory", "application/json", rBody)
			}

			leaderMarshalled, _ := json.Marshal(i)
			rBody := bytes.NewBuffer(leaderMarshalled)
			http.Post("http://proxyserver:7000/victory", "application/json", rBody)

			return
		}
	}

}

func getTest(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func postVictory(w http.ResponseWriter, r *http.Request) {
	m.Lock()
	defer m.Unlock()

	var newLeader int
	json.NewDecoder(r.Body).Decode(&newLeader)
	//fmt.Println("OLD:", leader, "NEW:", newLeader, time.Now())

	if newLeader != ID && leader == ID {
		leaderBeat <- true
		//fmt.Println("heartbeat stop, listening", time.Now())
		go listenHeartBeat()
	}

	if newLeader == ID && leader != ID {
		go leaderHeartBeat()

		go startLeaderServer()
	}

	if newLeader != ID && leader != ID && listening == false {
		//fmt.Println("listening")
		go listenHeartBeat()
	}

	leader = newLeader
}

func getHeartBeat(w http.ResponseWriter, r *http.Request) {
	followerBeat <- true
}

func listenHeartBeat() {
	listening = true
	for {
		select {
		case <-followerBeat:
			//fmt.Println("heartbeat received", time.Now())
		case <-time.After(3 * time.Second):
			//fmt.Println("stop listening... Election")
			go Election()
			listening = false
			return
		}
	}
}

func sendHeartBeat() {
	ActiveServers = 1
	for i := 0; i < ClusterSize; i++ {
		if i != ID {
			_, err := http.Get("http://server" + strconv.Itoa(i) + ":8001/heartbeat")

			if err == nil {
				ActiveServers++
			}
		}
	}

	ActiveServersMarshalled, _ := json.Marshal(ActiveServers)
	rBody := bytes.NewBuffer(ActiveServersMarshalled)
	http.Post("http://proxyserver:7000/active", "application/json", rBody)

}

func leaderHeartBeat() {
	for {
		select {
		case <-leaderBeat:
			return
		default:
			go sendHeartBeat()
			//fmt.Println("heartbeat", time.Now())
			time.Sleep(time.Second)
		}
	}
}
