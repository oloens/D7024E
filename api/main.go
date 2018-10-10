package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type File struct {
	Hash               string
	Value              []byte
	Pin                bool
	TimeSinceRepublish int
}

// Typ pseudokod ish

// Store a specified file in the Kademlia network
func Store(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	var file File
	_ = json.NewDecoder(r.Body).Decode(&file)
	hash := kademlia.Hash(file)
	json.NewEncoder(w).Encode(hash)
}

// Display contents of a specified file
func Read(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	var file File
	for _, files := range kademlia.files {
		if files.Hash == params["id"] {
			json.NewEncoder(w).Encode(files)
			return
		}
	}
	json.NewEncoder(w).Encode("Error: no file found")
}

// Pins a file in the Kademlia network
func Pin(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	hash := params["id"]
	kademlia.Pin(hash)
	json.NewEncoder(w).Encode("Pinned file with hash: " + hash)
}

// Unpins a file in the Kademlia network
func Unpin(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	hash := params["id"]
	kademlia.Unpin(hash)
	json.NewEncoder(w).Encode("Unpinned file with hash: " + hash)
}

// Main function to handle routes
func main() {
	router := mux.NewRouter()
	//Right methods?
	router.HandleFunc("/api/{file}", Store).Methods("POST")
	router.HandleFunc("/api/{id}", Read).Methods("GET")
	router.HandleFunc("/api/{id}", Pin).Methods("PUT")
	router.HandleFunc("/api/{id}", Unpin).Methods("DELETE")
	//Does kademlia use port 8000?
	log.Fatal(http.ListenAndServe(":8000", router))
}
