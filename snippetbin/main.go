package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

var (
	addr = flag.String("addr", "127.0.0.1:8080", "http service address")
)

func openDB() *gorm.DB {
	conn, err := gorm.Open("sqlite3", "./snippetbin.db")
	if err != nil {
		log.Fatal(err)
	}
	return conn
}

func getSnippets(db *gorm.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var snippets []Snippet
		if err := db.Find(&snippets).Error; err != nil {
			fmt.Errorf("Error retrieving snippets:", err)
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		response, err := json.Marshal(snippets)
		if err != nil {
			fmt.Errorf("Marshal error:", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(response)
	}
}

func getSnippet(db *gorm.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		snippetID := vars["id"]

		var snippet Snippet
		if err := db.First(&snippet, snippetID).Error; err != nil {
			fmt.Errorf("Error retrieving snippet:", err)
			http.Error(w, fmt.Sprintf("Snippet with ID: %s not found", snippetID), http.StatusNotFound)
			return
		}

		response, err := json.Marshal(snippet)
		if err != nil {
			fmt.Errorf("Marshal error:", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(response)
	}
}

func createSnippet(db *gorm.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var snippet Snippet

		err := json.NewDecoder(r.Body).Decode(&snippet)
		if err != nil {
			fmt.Errorf("Decode error: ", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err = db.Create(&snippet).Error; err != nil {
			fmt.Errorf("Error creating snippet: ", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		response, err := json.Marshal(snippet)
		if err != nil {
			fmt.Errorf("Marshal error:", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write(response)
	}
}

func updateSnippet(db *gorm.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var snippet Snippet
		var updated Snippet

		vars := mux.Vars(r)
		snippetID := vars["id"]

		if err := db.First(&snippet, snippetID).Error; err != nil {
			fmt.Errorf("Error retrieving snippet:", err)
			http.Error(w, fmt.Sprintf("Snippet with ID: %s not found", snippetID), http.StatusNotFound)
			return
		}

		err := json.NewDecoder(r.Body).Decode(&updated)
		if err != nil {
			fmt.Errorf("Decode error: ", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := db.Model(&snippet).Updates(updated).Error; err != nil {
			fmt.Errorf("Error updating snippet: ", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		response, err := json.Marshal(snippet)
		if err != nil {
			fmt.Errorf("Marshal error:", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(response)
	}
}

func deleteSnippet(db *gorm.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		snippetID := vars["id"]

		var snippet Snippet
		if err := db.First(&snippet, snippetID).Error; err != nil {
			fmt.Errorf("Error retrieving snippet:", err)
			http.Error(w, fmt.Sprintf("Snippet with ID: %s not found", snippetID), http.StatusNotFound)
			return
		}

		if err := db.Delete(&snippet).Error; err != nil {
			fmt.Errorf("Error deleting snippet: ", err)
			http.Error(w, fmt.Sprint("Error deleting snippet: ", err), http.StatusBadRequest)
			return
		}

		response, err := json.Marshal(snippet)
		if err != nil {
			fmt.Errorf("Marshal error:", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(response)
	}
}

func main() {
	flag.Parse()

	dbConn := openDB()
	defer dbConn.Close()

	dbConn.AutoMigrate(&Snippet{})

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/snippets", getSnippets(dbConn)).Methods("GET")
	router.HandleFunc("/snippets", createSnippet(dbConn)).Methods("POST")
	router.HandleFunc("/snippets/{id}", getSnippet(dbConn)).Methods("GET")
	router.HandleFunc("/snippets/{id}", updateSnippet(dbConn)).Methods("PATCH")
	router.HandleFunc("/snippets/{id}", deleteSnippet(dbConn)).Methods("DELETE")

	log.Fatal(http.ListenAndServe(*addr, router))
}
