package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"html/template"
	"log"
	"net/http"
)

var (
	DBConn *sql.DB
)

type QueryValue struct {
	Identifier string
}

type DataOutput struct {
	Identifier string    `json:"identifier"`
	Data       []float64 `json:"data"`
}

func main() {

	OpenDatabase("../storage.db")

	http.HandleFunc("/", index)
	http.HandleFunc("/identifiers", getPossibleIdentifiers)
	http.HandleFunc("/entries", getEntries)

	err := http.ListenAndServe(":8081", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func index(w http.ResponseWriter, r *http.Request) {

	t := template.New("index.html").Delims("[[", "]]")
	t, _ = t.ParseFiles("templates/index.html")

	t.ExecuteTemplate(w, "index.html", nil)
}

// Route handlers
//
//

func getPossibleIdentifiers(w http.ResponseWriter, r *http.Request) {

	identifiers := GetIdentifiers()

	entriesJSON, _ := json.Marshal(identifiers)
	w.Header().Set("Content-Type", "application/json")
	w.Write(entriesJSON)

}

func getEntries(w http.ResponseWriter, r *http.Request) {
	valuesRaw := r.URL.Query().Get("values")
	var queryValues []QueryValue

	err := json.Unmarshal([]byte(valuesRaw), &queryValues)
	if err != nil {
		fmt.Println(err)
	}

	var identifiers []string
	for _, queryValue := range queryValues {
		identifiers = append(identifiers, queryValue.Identifier)
	}

	entries := GetLatestEntries(identifiers, 200)

	entriesJSON, _ := json.Marshal(entries)
	w.Header().Set("Content-Type", "application/json")
	w.Write(entriesJSON)

}

// Database
//
//

func OpenDatabase(dbName string) {
	var err error
	DBConn, err = sql.Open("sqlite3", dbName)
	if err != nil {
		fmt.Sprintf("Issue with db: %s", err)
	}
}

func GetIdentifiers() []string {
	var output []string
	rows, err := DBConn.Query("SELECT handler_identifier FROM entries GROUP BY handler_identifier")

	if err != nil {
		panic(err)
	}

	for rows.Next() {
		var value string
		if err := rows.Scan(&value); err != nil {
			panic(err)
		}
		output = append(output, value)
	}

	return output

}

func GetLatestEntries(identifiers []string, limit int) []DataOutput {
	var output []DataOutput

	stmt, err := DBConn.Prepare(`
		SELECT output 
		FROM entries 
		WHERE handler_identifier = ?
		ORDER BY id DESC 
		LIMIT ?`)

	if err != nil {
		panic(err)
	}

	for _, identifier := range identifiers {
		var dataSingle []float64

		rows, err := stmt.Query(identifier, limit)
		if err != nil {
			panic(err)
		}

		for rows.Next() {
			var value float64
			if err := rows.Scan(&value); err != nil {
				panic(err)
			}
			dataSingle = append(dataSingle, value)
		}

		output = append(output, DataOutput{
			Identifier: identifier,
			Data:       reverse(dataSingle),
		})

	}

	return output
}

// Misc
//
//

// Reverses the provided array values
func reverse(numbers []float64) []float64 {
	newNumbers := make([]float64, len(numbers))
	for i, j := 0, len(numbers)-1; i < j; i, j = i+1, j-1 {
		newNumbers[i], newNumbers[j] = numbers[j], numbers[i]
	}
	return newNumbers
}
