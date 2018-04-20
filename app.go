package main

import (
	"fmt"
	"net/http"
	"bufio"
	"regexp"
	"os"
	"io"
)
const addr = "localhost:55555"

var pattern *regexp.Regexp

func main() {
	pattern = regexp.MustCompile("[<,>,:,\",/,\\,|,?,*]")

	http.HandleFunc("/", handle)

	err := http.ListenAndServe(addr, nil)
	fmt.Println(err.Error())
}

func handle(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Request from ", r.RemoteAddr, "with method", r.Method)

	if r.Method == http.MethodPost {
		scanner := bufio.NewScanner(r.Body)
		total := 0
		jobs := make(chan string, 100)
		results := make(chan string, 100)

		for i:=1; i<=3; i++ {
			go fetchWorker(i, jobs, results)
		}

		for scanner.Scan() {
			jobs <- scanner.Text()
			total++
		}

		fmt.Println("### SCANNED ###")

		close(jobs)

		if total > 0 {
			for j := 1; j<=total; j++ {
				<- results
			}
		}

		if err := scanner.Err(); err != nil {
			http.Error(w, "Error during request body reading", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)

	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func safePath(url string) string {
	return pattern.ReplaceAllString(url, "_")
}

func fetchWorker(id int, jobs <- chan string, results chan <- string) {
	fmt.Println("Worker", id, "started")
	var count int = 0

	for j := range jobs {
		count ++
		fmt.Println("Worker", id, "Starting JOB", j)
		result := fetch(j)
		fmt.Println("Worker", id, "Finished JOB", j)

		fmt.Println(result)
		fmt.Println("-----")
		results <- result
	}
	fmt.Println("*** worker ",id, "runs ", count, " times" )
	fmt.Println("Worker", id, "Finished")
}

func fetch(url string) (result string) {
	response, err := http.Get(url)

	if err != nil {
		result = "Failed to fetch due to Error " + url + " " + err.Error()
	} else if response.StatusCode != http.StatusOK {
		result = "Failed to fetch due to StatusCode " + url + " " + fmt.Sprintf("%d", response.StatusCode)
	} else {
		defer response.Body.Close()

		fname := "tmp/" + safePath(url)
		f, err := os.Create(fname)

		if err != nil {
			result = "Error during creating file" +  fname + " " + err.Error()
		} else {
			defer f.Close()

			_, err := io.Copy(f, response.Body)

			if err != nil {
				result = "Error during copying response of " + url + " " + err.Error()
			} else {
				result = "Fetched " + url + " as " + fname
			}
		}
	}

	return
}