package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	ics "github.com/arran4/golang-ical"
)

const apiURL = "https://api.adelaide.edu.au/api/generic-query-structured/v1/?target=/system/TIMETABLE_WEEKLY/queryx/%s,%d&MaxRows=9999"

type ApiResponse struct {
	Status string `json:"status"`
	Data   struct {
		Query struct {
			NumRows   int     `json:"numrows"`
			Queryname string  `json:"queryname"`
			Classes   []Class `json:"rows"`
		} `json:"query"`
	} `json:"data"`
}

type Class struct {
	Type     string `json:"D.XLATLONGNAME"`
	Start    string `json:"START_TIME"`
	End      string `json:"END_TIME"`
	Course   string `json:"B.DESCR"`
	Building string `json:"F.DESCR"`
	Room     string `json:"E.ROOM"`
	Date     string `json:"DATE"`
}

var tmpl *template.Template

func main() {

	http.Handle("/dist/", http.StripPrefix("/dist/", http.FileServer(http.Dir("./dist/"))))

	http.HandleFunc("/api/", apiHandler)
	http.HandleFunc("/", indexHandler)

	tmpl = template.Must(tmpl.ParseFiles("./src/html/index.html"))

	http.ListenAndServe("0.0.0.0:10000", nil)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)

	tmpl.ExecuteTemplate(w, "index.html", nil)
}

func apiHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)

	if r.Method != http.MethodPost {
		http.Error(w, "bad method to api route", http.StatusBadRequest)
		return
	}

	// Parse the form data from the request
	err := r.ParseMultipartForm(16 * 2 * 10)
	if err != nil {
		http.Error(w, fmt.Sprintln("error parsing form:", err), http.StatusBadRequest)
		return
	}

	// Extract the auth-token and student-id fields from the form data
	auth := r.FormValue("auth-token")
	id := r.FormValue("student-id")

	log.Printf("id: %s\tauth: %s", id, auth)

	err = writeCalendar(w, id, auth)
	if err != nil {
		log.Println("could not write calendar:", err)
	}

	w.Header().Add("content-type", "text/calendar")
}

func writeCalendar(w io.Writer, studentID, auth string) error {
	// Create a new HTTP client.
	client := &http.Client{}

	// Make a channel to pass results.
	ch := make(chan []Class, 42*10)

	// Make a channel to catch errors.
	errCh := make(chan error)

	var wg sync.WaitGroup

	// Loop through until end of year.
	log.Println("Fetching Timetable")
	for i := 0; i < 42; i++ {
		wg.Add(1)
		go func(j int) {
			// Create request.
			req, err := http.NewRequest(http.MethodGet, fmt.Sprintf(apiURL, studentID, j*7), nil)
			if err != nil {
				errCh <- fmt.Errorf("error creating request: %w", err)
				wg.Done()
				return
			}

			// Add authorisation.
			req.Header.Set("Authorization", auth)

			// Send the request.
			resp, err := client.Do(req)
			if err != nil {
				errCh <- fmt.Errorf("error sending request: %w", err)
				wg.Done()
				return
			}
			defer resp.Body.Close()

			// Read the response body.
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				errCh <- fmt.Errorf("error reading response body: %w", err)
				wg.Done()
				return
			}

			// Unmarshal the response.
			var v ApiResponse
			err = json.Unmarshal(body, &v)
			if err != nil {
				errCh <- fmt.Errorf("error unmarshalling json: %w", err)
				wg.Done()
				return
			}

			ch <- v.Data.Query.Classes
			wg.Done()
		}(i)
	}

	// Wait to close the channel.
	wg.Wait()
	close(ch)
	close(errCh)

	if <-errCh != nil {
		log.Println("ERROR")
	}

	// Read the data from the channel.
	var Classes []Class
	for classList := range ch {
		Classes = append(Classes, classList...)
	}

	log.Printf("Got timetable, %d events found", len(Classes))

	if len(Classes) == 0 {
		return fmt.Errorf("no classes found, try updating authToken")
	}

	// Create a calendar.
	cal := ics.NewCalendar()

	for _, c := range Classes {
		ev := cal.AddEvent(generateUID(c.Start, c.End))
		ev.SetSummary(c.Course + " - " + c.Type)
		start, end, err := parseStartEnd(c.Start, c.End, c.Date)
		if err != nil {
			return fmt.Errorf("failed parsing event time for event: %w", err)
		}
		ev.SetCreatedTime(time.Now())
		ev.SetDtStampTime(time.Now())
		ev.SetModifiedAt(time.Now())
		ev.SetStartAt(start)
		ev.SetEndAt(end)
		ev.SetLocation(c.Room + " - " + c.Building)
	}

	err := cal.SerializeTo(w)
	if err != nil {
		return fmt.Errorf("could not write calendar to io.writer: %w", err)
	}

	return nil
}
