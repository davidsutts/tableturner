package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
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

type Calendar []Class

const (
	statusSuccess = "success"
	apiURL        = ""
	authToken     = ""
)

func main() {
	// Create a new HTTP client.
	client := &http.Client{}

	// Create Calendar.
	var Cal Calendar

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
	for classes := range ch {
		Cal = append(Cal, classes...)
	}

}
