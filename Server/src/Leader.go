package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"hash/fnv"
	"io"
	"net/http"
	"strconv"
)

var replicaFactor = ClusterSize/2 + 1

var serverToHandle = 0

func storeDataLeader(w http.ResponseWriter, r *http.Request) {

	if leader == ID && serverToHandle != ID {
		body, _ := io.ReadAll(r.Body)
		defer r.Body.Close()

		rbody := io.NopCloser(bytes.NewReader(body))
		req, _ := http.NewRequest(http.MethodPost, "http://server"+strconv.Itoa(serverToHandle)+":8000/datastore", rbody)
		resp, _ := http.DefaultClient.Do(req)

		w.WriteHeader(resp.StatusCode)

		serverToHandle = mod(serverToHandle+1, ClusterSize)

		return
	}

	if leader == ID && serverToHandle == ID {
		fmt.Println("Server", ID, "serving")
		serverToHandle = mod(serverToHandle+1, ClusterSize)

		body, _ := io.ReadAll(r.Body)
		defer r.Body.Close()

		type Cell struct {
			Key    string `json:"key"`
			Value  any    `json:"value"`
			Server int    `json:"server"`
		}

		var cell Cell

		r.Body = io.NopCloser(bytes.NewReader(body))
		json.NewDecoder(r.Body).Decode(&cell)

		srvNumber := getServerForKey(cell.Key)

		cell.Server = srvNumber

		cellMarshalled, _ := json.Marshal(cell)

		var respSent = false

		for i := 0; i < replicaFactor; i++ {
			rBody := bytes.NewBuffer(cellMarshalled)
			req, _ := http.NewRequest(http.MethodPost, "http://server"+strconv.Itoa(mod(i+srvNumber, ClusterSize))+":8001/data", rBody)
			resp, err := http.DefaultClient.Do(req)

			if err == nil && resp.StatusCode == http.StatusUnprocessableEntity {
				w.WriteHeader(resp.StatusCode)
				return
			}

			if err == nil && respSent == false {
				w.WriteHeader(resp.StatusCode)
				respSent = true
			}

		}

		if respSent == false {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	if leader != ID {
		fmt.Println("Server", ID, "serving")
		body, _ := io.ReadAll(r.Body)
		defer r.Body.Close()

		type Cell struct {
			Key    string `json:"key"`
			Value  any    `json:"value"`
			Server int    `json:"server"`
		}

		var cell Cell

		r.Body = io.NopCloser(bytes.NewReader(body))
		json.NewDecoder(r.Body).Decode(&cell)

		srvNumber := getServerForKey(cell.Key)

		cell.Server = srvNumber

		cellMarshalled, _ := json.Marshal(cell)

		var respSent = false

		for i := 0; i < replicaFactor; i++ {
			rBody := bytes.NewBuffer(cellMarshalled)
			req, _ := http.NewRequest(http.MethodPost, "http://server"+strconv.Itoa(mod(i+srvNumber, ClusterSize))+":8001/data", rBody)
			resp, err := http.DefaultClient.Do(req)

			if err == nil && resp.StatusCode == http.StatusUnprocessableEntity {
				w.WriteHeader(resp.StatusCode)
				return
			}

			if err == nil && respSent == false {
				w.WriteHeader(resp.StatusCode)
				respSent = true
			}

		}

		if respSent == false {
			w.WriteHeader(http.StatusInternalServerError)
		}

		return
	}
}

func updateDataLeader(w http.ResponseWriter, r *http.Request) {

	if leader == ID && serverToHandle != ID {
		body, _ := io.ReadAll(r.Body)
		defer r.Body.Close()

		rbody := io.NopCloser(bytes.NewReader(body))
		req, _ := http.NewRequest(http.MethodPut, "http://server"+strconv.Itoa(serverToHandle)+":8000/datastore", rbody)
		resp, _ := http.DefaultClient.Do(req)

		w.WriteHeader(resp.StatusCode)

		serverToHandle = mod(serverToHandle+1, ClusterSize)

		return
	}

	if leader == ID && serverToHandle == ID {
		serverToHandle = mod(serverToHandle+1, ClusterSize)

		body, _ := io.ReadAll(r.Body)
		defer r.Body.Close()

		type Cell struct {
			Key    string `json:"key"`
			Value  any    `json:"value"`
			Server int    `json:"server"`
		}

		var cell Cell

		r.Body = io.NopCloser(bytes.NewReader(body))
		json.NewDecoder(r.Body).Decode(&cell)

		srvNumber := getServerForKey(cell.Key)

		cell.Server = srvNumber

		cellMarshalled, _ := json.Marshal(cell)

		var respSent = false

		for i := 0; i < replicaFactor; i++ {
			rBody := bytes.NewBuffer(cellMarshalled)
			req, _ := http.NewRequest(http.MethodPut, "http://server"+strconv.Itoa(mod(i+srvNumber, ClusterSize))+":8001/data", rBody)
			resp, err := http.DefaultClient.Do(req)

			if err == nil && resp.StatusCode == http.StatusNotFound {
				w.WriteHeader(resp.StatusCode)
				return
			}

			if err == nil && respSent == false {
				w.WriteHeader(resp.StatusCode)
				respSent = true
			}

		}

		if respSent == false {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	if leader != ID {
		body, _ := io.ReadAll(r.Body)
		defer r.Body.Close()

		type Cell struct {
			Key    string `json:"key"`
			Value  any    `json:"value"`
			Server int    `json:"server"`
		}

		var cell Cell

		r.Body = io.NopCloser(bytes.NewReader(body))
		json.NewDecoder(r.Body).Decode(&cell)

		srvNumber := getServerForKey(cell.Key)

		cell.Server = srvNumber

		cellMarshalled, _ := json.Marshal(cell)

		var respSent = false

		for i := 0; i < replicaFactor; i++ {
			rBody := bytes.NewBuffer(cellMarshalled)
			req, _ := http.NewRequest(http.MethodPut, "http://server"+strconv.Itoa(mod(i+srvNumber, ClusterSize))+":8001/data", rBody)
			resp, err := http.DefaultClient.Do(req)

			if err == nil && resp.StatusCode == http.StatusNotFound {
				w.WriteHeader(resp.StatusCode)
				return
			}

			if err == nil && respSent == false {
				w.WriteHeader(resp.StatusCode)
				respSent = true
			}

		}

		if respSent == false {
			w.WriteHeader(http.StatusInternalServerError)
		}

		return
	}
}

func removeDataLeader(w http.ResponseWriter, r *http.Request) {

	if leader == ID && serverToHandle != ID {

		params := mux.Vars(r)

		req, _ := http.NewRequest(http.MethodDelete, "http://server"+strconv.Itoa(serverToHandle)+":8000/datastore/"+params["key"], nil)
		resp, _ := http.DefaultClient.Do(req)

		w.WriteHeader(resp.StatusCode)

		serverToHandle = mod(serverToHandle+1, ClusterSize)

		return
	}

	if leader == ID && serverToHandle == ID {
		serverToHandle = mod(serverToHandle+1, ClusterSize)

		params := mux.Vars(r)

		srvNumber := getServerForKey(params["key"])

		var respSent = false

		for i := 0; i < replicaFactor; i++ {
			req, _ := http.NewRequest(http.MethodDelete,
				"http://server"+strconv.Itoa(mod(i+srvNumber, ClusterSize))+":8001/data/"+params["key"]+"/"+strconv.Itoa(srvNumber),
				nil)

			resp, err := http.DefaultClient.Do(req)

			if err == nil && resp.StatusCode == http.StatusNotFound {
				w.WriteHeader(resp.StatusCode)
				return
			}

			if err == nil && respSent == false {
				w.WriteHeader(resp.StatusCode)
				respSent = true
			}

		}

		if respSent == false {
			w.WriteHeader(http.StatusInternalServerError)
		}

		return
	}

	if leader != ID {
		params := mux.Vars(r)

		srvNumber := getServerForKey(params["key"])

		var respSent = false

		for i := 0; i < replicaFactor; i++ {
			req, _ := http.NewRequest(http.MethodDelete,
				"http://server"+strconv.Itoa(mod(i+srvNumber, ClusterSize))+":8001/data/"+params["key"]+"/"+strconv.Itoa(srvNumber),
				nil)

			resp, err := http.DefaultClient.Do(req)

			if err == nil && resp.StatusCode == http.StatusNotFound {
				w.WriteHeader(resp.StatusCode)
				return
			}

			if err == nil && respSent == false {
				w.WriteHeader(resp.StatusCode)
				respSent = true
			}

		}

		if respSent == false {
			w.WriteHeader(http.StatusInternalServerError)
		}

		return
	}
}

func getDataLeader(w http.ResponseWriter, r *http.Request) {

	if leader == ID && serverToHandle != ID {
		params := mux.Vars(r)

		resp, _ := http.Get("http://server" + strconv.Itoa(leader) + ":8000/datastore/" + params["key"])

		body, _ := io.ReadAll(resp.Body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(resp.StatusCode)
		w.Write(body)

		serverToHandle = mod(serverToHandle+1, ClusterSize)

		return
	}

	if leader == ID && serverToHandle == ID {
		serverToHandle = mod(serverToHandle+1, ClusterSize)

		params := mux.Vars(r)

		srvNumber := getServerForKey(params["key"])

		for i := 0; i < replicaFactor; i++ {
			req, _ := http.NewRequest(http.MethodGet,
				"http://server"+strconv.Itoa(mod(i+srvNumber, ClusterSize))+":8001/data/"+params["key"]+"/"+strconv.Itoa(srvNumber),
				nil)

			resp, err := http.DefaultClient.Do(req)

			if err == nil {
				body, _ := io.ReadAll(resp.Body)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(resp.StatusCode)
				w.Write(body)
				return
			}

		}

		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	if leader != ID {
		params := mux.Vars(r)

		srvNumber := getServerForKey(params["key"])

		for i := 0; i < replicaFactor; i++ {
			req, _ := http.NewRequest(http.MethodGet,
				"http://server"+strconv.Itoa(mod(i+srvNumber, ClusterSize))+":8001/data/"+params["key"]+"/"+strconv.Itoa(srvNumber),
				nil)

			resp, err := http.DefaultClient.Do(req)

			if err == nil {
				body, _ := io.ReadAll(resp.Body)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(resp.StatusCode)
				w.Write(body)
				return
			}

		}

		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func mod(a int, b int) int {
	return (b + (a % b)) % b
}

func getServerForKey(s string) int {
	h := fnv.New32a()
	h.Write([]byte(s))
	return int(h.Sum32() % ClusterSize)
}
