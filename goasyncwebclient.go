package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"
)

var urls = []string{}

type HttpResponse struct {
	url        string
	response   *http.Response
	err        error
	body       []byte
	err2       error
	receivedat time.Time //time when response body has been fully read
	ackat      time.Time //time when first response headers
	sentat     time.Time //time at point of send
}

func (r *HttpResponse) ackDelayMillis() int64 {
	return r.ackat.Sub(r.sentat).Nanoseconds() / time.Millisecond.Nanoseconds()
}

func (r *HttpResponse) recDelayMillis() int64 {
	return r.receivedat.Sub(r.sentat).Nanoseconds() / time.Millisecond.Nanoseconds()
}

func (r *HttpResponse) generalDelayMillis(generalTime time.Time) int64 {
	return generalTime.Sub(r.sentat).Nanoseconds() / time.Millisecond.Nanoseconds()
}

func (r *HttpResponse) bodyLen() int {
	if r.body != nil {
		return len(r.body)
	}
	return 0
}

func asyncHttpGets(urls []string, appendt bool, cookiename string, cookievalue string) ([]*HttpResponse, bool) {
	ch := make(chan *HttpResponse)
	responses := []*HttpResponse{}
	wasError := false
	timeout := time.Duration(time.Duration(90) * time.Second)
	client := http.Client{
		Timeout: timeout,
	}
	//client := http.Client{}
	for _, url := range urls {
		go func(url string, appendt bool, cookiename string, cookievalue string) {
			//		fmt.Printf("Fetching %s \n", url)
			//TODO i would like to return early if error but we are committed to do all urls elsewhere
			clientTime := time.Now().UnixNano() / 1000000
			if appendt {
				url = fmt.Sprintf("%s%d", url, clientTime)
			}
			req, err := http.NewRequest("GET", url, nil)
			if err == nil {
				if(cookiename != "" && cookievalue != ""){
					cookie := &http.Cookie{ Name: cookiename ,
						Value: cookievalue, 
					}
					req.AddCookie(cookie);
				}
				sentat := time.Now()

				resp, err := client.Do(req)
				ackat := time.Now()
				if err == nil && resp != nil {
					defer resp.Body.Close()
					body, err2 := ioutil.ReadAll(resp.Body)
					ch <- &HttpResponse{url, resp, err, body, err2, time.Now(), ackat, sentat}
					if err2 != nil {
						wasError = true
					}
					//examine those headers and cookies
					//for ci, c := range resp.Cookies() {

					//}
				} else {
					ch <- &HttpResponse{url, resp, err, nil, nil, time.Now(), ackat, sentat}
					if err != nil {
						wasError = true
					}
				}
			} else {
				sentat := time.Now()
				ch <- &HttpResponse{url, nil, err, nil, nil, sentat, sentat, sentat}
				if err != nil {
					wasError = true
				}
			}
		}(url, appendt, cookiename, cookievalue)
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
				return responses, wasError
			}
		case <-time.After(50 * time.Millisecond):
			fmt.Printf(".")
		}
		if len(responses) == len(urls) {
			return responses, wasError
		}
	}
	return responses, wasError
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

///increment in 20% each time

func computeTrialIncrement(trialConcurrent int) int {
	trialIncrement := (trialConcurrent * 20) / 100
	if trialIncrement < 1 { //but check for quantisation error
		trialIncrement = 1
	}
	return trialIncrement
}

func main() {
	protocolPtr := flag.String("protocol", "http://", "protocol prefix to domain e.g. http://")
	domainPtr := flag.String("domain", "127.0.0.1", "domain(e.g. host name)")
	portPtr := flag.Int("port", 80, "port typically 80 or 443")
	concurrentPtr := flag.Int("c", 1, "concurrency e.g simultaneous connections number 7000")
	pathPtr := flag.String("path", "/", "path should start with a / and maybe end with one if also using the -appendi flag")
	nPtr := flag.Int("n", 1, "total number of calls processed. should be a multiple of c")
	appendiPtr := flag.Bool("appendi", false, "true if the connection index should be appended to url")
	appendtPtr := flag.Bool("appendt", false, "true if the send timestamp (milliseconds since epoch) should be appended to url")
	filenamePtr := flag.String("ofile", "/tmp/goout", "filename to write output to")
	rampToFailPtr := flag.Bool("rampToFail", true, "default true and if true will ramp up the connections until first fail")
	minConcurrentPtr := flag.Int("minc", 1, "minimum concurrency to start from")
	ackTimeoutPtr := flag.Int64("acktimeout", 2000, "maximum time in milliseconds to receive some acknowledgement (headers) from server")
    cookieNamePtr := flag.String("cookieName","SESSION","cookie header name e.g. SESSION");
    cookieValuePtr := flag.String("cookieValue","","cookie header value e.g. 894b4e8a-f830-4d0c-bdbf-d9084eaaa986");
 
 
	t0 := time.Now()
	flag.Parse()

	f, err := os.Create(*filenamePtr)
	check(err)
	w := bufio.NewWriter(f)

	if *nPtr < *concurrentPtr {
		*concurrentPtr = *nPtr
	}

	//outer ramploop here
	wasError := false
	wasTooSlow := false
	trialConcurrent := *concurrentPtr
	if *rampToFailPtr {
		trialConcurrent = *minConcurrentPtr
	}

	for trialConcurrent <= *concurrentPtr {

		urls = make([]string, trialConcurrent)

		for i := 0; i < trialConcurrent; i++ {
			var urlPtr string = fmt.Sprintf("%s%s:%d%s", *protocolPtr, *domainPtr, *portPtr, *pathPtr)
			if *appendiPtr {
				urlPtr = fmt.Sprintf("%s%d", urlPtr, i)
			}
			urls[i] = urlPtr

		}

		results, wasTrialError := asyncHttpGets(urls, *appendtPtr, *cookieNamePtr, *cookieValuePtr)
		if wasTrialError {
			wasError = true
		}

		var totalBodyLen int64 = 0
		var earliestStartTime time.Time = time.Now()
		var latestRecTime time.Time = t0
		var totalGoodResponses int = 0

		for _, result := range results {
			var messPtr string
			if result != nil && result.response != nil {
				ackDelay := result.ackDelayMillis()
				totalBodyLen += (int64)(result.bodyLen())
				totalGoodResponses++
				if ackDelay > *ackTimeoutPtr {
					wasTooSlow = true
				}
				if result.sentat.Before(earliestStartTime) {
					earliestStartTime = result.sentat
				}
				if result.receivedat.After(latestRecTime) {
					latestRecTime = result.receivedat
				}

				var serverDelayMillis int64 = 0

				if result.bodyLen() > 1 && bytes.IndexRune(result.body, '{') == 0 {
					//its a json object
					//see if it has a pong
					
					var dat map[string]interface{}
					if err := json.Unmarshal(result.body, &dat); err != nil {
						panic(err)
					}
					if val, ok := dat["pong"]; ok {
						var serverTime string = val.(string)
						var serverTimeInt int64 = 0
						
						serverTimeInt, err = strconv.ParseInt(serverTime, 10, 64)
						if err == nil {
							serverDelayMillis = result.generalDelayMillis(time.Unix(serverTimeInt/int64(1000), (serverTimeInt%int64(1000))*int64(1000000)))
						} else {
							panic(err)
						}

					}

				}
				messPtr = fmt.Sprintf("ackmillis=%d recmillis=%d serverDelayMillis=%d bodyLen=%d %s status: %s\n", ackDelay, result.recDelayMillis(), serverDelayMillis, result.bodyLen(), result.url,
					result.response.Status)
				//examine those headers and cookies
				var messCookies string = ""
				for _, c := range result.response.Cookies() {
					messCookies = fmt.Sprintf("%s%s\n", messCookies, c.String())
				}
				messPtr = fmt.Sprintf("%s%s\n", messPtr, messCookies)

			} else if result != nil {
				ackDelay := result.ackDelayMillis()
				if ackDelay > *ackTimeoutPtr {
					wasTooSlow = true
				}
				messPtr = fmt.Sprintf("millis=%d recmillis=%d %s with error : %s\n", ackDelay, result.recDelayMillis(), result.url,
					result.err)
			} else {
				messPtr = "nil result\n"
			}
			w.WriteString(messPtr)
		}
		w.Flush()
		var speed float64 = (float64(totalBodyLen) / (latestRecTime.Sub(earliestStartTime).Seconds())) / 1024.0
		var rate float64 = float64(totalGoodResponses) / (latestRecTime.Sub(earliestStartTime).Seconds())

		fmt.Printf("\ntrialConcurrent=%d go look at file %s  data speed=%f KiB/s request rate=%f Request/s", trialConcurrent, *filenamePtr, speed, rate)

		//reached the end of the trial loop. need to computre the next iteration if any
		if wasError {
			fmt.Printf("\nwas an ERROR\n")
			break //stop on error
		}
		if wasTooSlow {
			fmt.Printf("\nwas TOO SLOW (> acktimeout=%d)\n", *ackTimeoutPtr)
			break //stop on error
		}
		trialConcurrent += computeTrialIncrement(trialConcurrent)

	}
	fmt.Printf("\nTest ended\n")
}
