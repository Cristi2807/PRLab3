package main

import (
	"bytes"
	"encoding/json"
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

var ID, _ = strconv.Atoi(os.Getenv("ID"))

var leader = -1

var leaderBeat = make(chan bool)
var followerBeat = make(chan bool)

var mutex sync.Mutex

var listening = false

func Election() {

	for i := ClusterSize - 1; i >= 0; i-- {
		_, err := net.ResolveTCPAddr("tcp", "server"+strconv.Itoa(i)+":3000")

		if err == nil {
			//fmt.Println("Leader", i, time.Now())
			for j := 0; j <= i; j++ {
				tcpAddr, err1 := net.ResolveTCPAddr("tcp", "server"+strconv.Itoa(j)+":3000")

				if err1 == nil {

					conn, _ := net.DialTCP("tcp", nil, tcpAddr)

					objMarshalled, _ := json.Marshal(i)

					conn.Write(objMarshalled)

					conn.Close()

				}
			}

			leaderMarshalled, _ := json.Marshal(i)
			rBody := bytes.NewBuffer(leaderMarshalled)
			http.Post("http://proxyserver:7000/victory", "application/json", rBody)

			return
		}
	}

}

func postVictory(conn net.Conn) {
	mutex.Lock()
	defer mutex.Unlock()

	var newLeader int
	d := json.NewDecoder(conn)

	d.Decode(&newLeader)
	//fmt.Println("OLD:", leader, "NEW:", newLeader, time.Now())

	if newLeader != ID && leader == ID {
		leaderBeat <- true

		//fmt.Println("heartbeat stop, listening", time.Now())
		go listenHeartBeat()
	}

	if newLeader == ID && leader != ID {
		go leaderHeartBeat()
	}

	if newLeader != ID && leader != ID && listening == false {
		//fmt.Println("listening")
		go listenHeartBeat()
	}

	leader = newLeader
}

func getHeartBeat(conn *net.UDPConn) {
	d := json.NewDecoder(conn)

	type Obj struct {
		Heartbeat string
	}

	var obj Obj

	d.Decode(&obj)

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

	udpAddr, _ := net.ResolveUDPAddr("udp", "239.0.0.0:5000")

	conn, _ := net.DialUDP("udp", nil, udpAddr)

	obj := struct {
		Heartbeat string
	}{
		Heartbeat: "heartbeat",
	}

	objMarshalled, _ := json.Marshal(obj)

	conn.Write(objMarshalled)

	conn.Close()
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
