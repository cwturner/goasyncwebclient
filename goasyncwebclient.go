package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

var urls = []string{
//  "http://www.rubyconf.org/",
//  "http://golang.org/",
//  "http://matt.aimonetti.net/",
}

type HttpResponse struct {
	url        string
	response   *http.Response
	err        error
	body       []byte
	err2       error
	receivedat time.Time
}

func asyncHttpGets(urls []string) []*HttpResponse {
	ch := make(chan *HttpResponse)
	responses := []*HttpResponse{}
	timeout := time.Duration(time.Duration(90) * time.Second)
	client := http.Client{
		Timeout: timeout,
	}
	//client := http.Client{}
	for _, url := range urls {
		go func(url string) {
			//		fmt.Printf("Fetching %s \n", url)
			resp, err := client.Get(url)
			if err != nil && resp != nil {
				defer resp.Body.Close()
				body, err2 := ioutil.ReadAll(resp.Body)
				ch <- &HttpResponse{url, resp, err, body, err2, time.Now()}
			} else {

				ch <- &HttpResponse{url, resp, err, nil, nil, time.Now()}
			}
		}(url)
	}

	for {
		select {
		case r := <-ch:
			//			fmt.Printf("%s was fetched\n", r.url)
			if r.err != nil {
				//				fmt.Println("with an error", r.err)
			}
			responses = append(responses, r)
			if len(responses) == len(urls) {
				return responses
			}
		case <-time.After(50 * time.Millisecond):
			fmt.Printf(".")
		}
	}
	return responses
}

func main() {
	t0 := time.Now()
	for i := 0; i < 7000; i++ {
		urls = append(urls, fmt.Sprintf("%s%d", "http://ct-lin-1.pgitech.local:5555/", i))
		//urls = append(urls, fmt.Sprintf("%s", "http://ct-lin-1.pgitech.local/"))
	}

	results := asyncHttpGets(urls)
	for _, result := range results {
		if result != nil && result.response != nil {
			fmt.Printf("millis=%d %s status: %s\n", result.receivedat.Sub(t0)/time.Millisecond, result.url,
				result.response.Status)
		} else if result != nil {
			fmt.Printf("millis=%d %s with error : %s\n", result.receivedat.Sub(t0)/time.Millisecond, result.url,
				result.err)
		}
	}
}
