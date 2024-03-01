package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	ics "github.com/arran4/golang-ical"
)

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

func main() {
	// Create a new HTTP client.
	client := &http.Client{}

	// Make a channel to pass results.
	ch := make(chan []Class, 42*10)

	var wg sync.WaitGroup

	// Loop through until failure (End of Sem).
	log.Println("Fetching Timetable")
	for i := 0; i < 42; i++ {
		wg.Add(1)
		go func(j int) {
			// Create request.
			req, err := http.NewRequest(http.MethodGet, fmt.Sprintf(apiURL, j*7), nil)
			if err != nil {
				log.Println("error creating request:", err)
				wg.Done()
				return
			}

			// Add authorisation.
			req.Header.Set("Authorization", authToken)

			// Send the request.
			resp, err := client.Do(req)
			if err != nil {
				log.Println("error sending request:", err)
				wg.Done()
				return
			}
			defer resp.Body.Close()

			// Read the response body.
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Println("error reading response body:", err)
				wg.Done()
				return
			}

			// Unmarshal the response.
			var v ApiResponse
			err = json.Unmarshal(body, &v)
			if err != nil {
				log.Println("Error unmarshalling json:", err)
				wg.Done()
				return
			}

			// log.Printf("status: %s\tnumrows: %d", v.Status, v.Data.Query.NumRows)

			ch <- v.Data.Query.Classes
			wg.Done()
		}(i)
	}

	// Wait to close the channel.
	wg.Wait()
	close(ch)

	// Read the data from the channel.
	var Classes []Class
	for classList := range ch {
		Classes = append(Classes, classList...)
	}

	log.Printf("Got timetable, %d events found", len(Classes))

	// Create a calendar.
	cal := ics.NewCalendar()

	for _, c := range Classes {
		ev := cal.AddEvent(generateUID(c.Start, c.End))
		ev.SetSummary(c.Course + " - " + c.Type)
		start, end, err := parseStartEnd(c.Start, c.End, c.Date)
		if err != nil {
			log.Println("failed parsing event time for event:", err)
			return
		}
		ev.SetCreatedTime(time.Now())
		ev.SetDtStampTime(time.Now())
		ev.SetModifiedAt(time.Now())
		ev.SetStartAt(start)
		ev.SetEndAt(end)
		ev.SetLocation(c.Room + " - " + c.Building)
	}

	file, err := os.Create("timetable.ics")
	if err != nil {
		log.Println("could not create calendar file:", err)
		return
	}

	err = cal.SerializeTo(file)
	if err != nil {
		log.Println("could not write calendar to file:", err)
		return
	}

	log.Println("Generated .ics file at timetable.ics")

}
