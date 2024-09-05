package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/Masterminds/sprig"
)

func init() {
	// setup mongoDB
	// mongodb.ConnectDB()
	ConnectDB()
}

// Handler function to render the paginated data
func handleData(w http.ResponseWriter, r *http.Request) {
	// Parse the page number from URL query, default is 1
	pageStr := r.URL.Query().Get("page")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	response, err := getAnomalyData(page) // Fetch your actual data here
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Create pagination links (up to 10 page links shown at a time)
	paginationLinks := []int{}
	for i := 1; i <= response.TotalPages; i++ {
		paginationLinks = append(paginationLinks, i)
	}

	// Template data including paginated results and total page info
	type TemplateData struct {
		Data            []AnomalyDataVM
		Page            int
		TotalPages      int
		PaginationLinks []int
	}

	tmplData := TemplateData{
		Data:            response.Items,
		Page:            response.PageNumber,
		TotalPages:      response.TotalPages,
		PaginationLinks: paginationLinks,
	}

	// Use a template with sprig functions
	tmpl := template.Must(template.New("index.html").Funcs(sprig.FuncMap()).ParseFiles("templates/index.html"))
	err = tmpl.Execute(w, tmplData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func main() {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("GET /data", handleData)

	http.HandleFunc("GET /ping", func(w http.ResponseWriter, r *http.Request) {
		now := time.Now()
		w.WriteHeader(http.StatusOK)
		w.Header().Add("Content-Type", "text/csv")
		w.Write([]byte("GET OK"))
		logRequestData(http.StatusOK, time.Until(now), now, r)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("Starting server on :" + port + "...")
	err := http.ListenAndServe(":" + port, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func logRequestData(code int, timeElapsed time.Duration, timeForReceivedRequest time.Time, req *http.Request) {
	fmt.Printf("%v\t%v\t%v\t%v\t%v\t%v\n",
		timeForReceivedRequest.Format(time.RFC3339),
		code,
		fmt.Sprint(timeElapsed.Seconds()*-1)+"s",
		req.RemoteAddr,
		req.Method,
		req.URL.Path,
	)
}
