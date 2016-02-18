package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
	"flag"
	"os"
	"bufio"
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
		if len(responses) == len(urls) {
				return responses
			}
	}
	return responses
}

func check(e error) {
    if e != nil {
        panic(e)
    }
}

func main() {
	protocolPtr := flag.String("protocol","http://","protocol prefix to domain e.g. http://")
	domainPtr := flag.String("domain", "172.16.100.225", "domain(e.g. host name)")
	portPtr := flag.Int("port",80,"port typically 80 or 443")
	concurrentPtr := flag.Int("c",1,"concurrency e.g simultaneous connections number 7000")
	pathPtr := flag.String("path","/","path should start with a / and maybe end with one if also using the -appendi flag")
	nPtr := flag.Int("n",1,"total number of calls processed. should be a multiple of c")
	appendiPtr := flag.Bool("appendi", false, "true if the connection index should be appended to url")
	filenamePtr := flag.String("ofile","/tmp/goout","filename to write output to");
	
	flag.Parse()
	
	if *nPtr < *concurrentPtr {
		*concurrentPtr = *nPtr;
	}
	
	t0 := time.Now()
	for i := 0; i < *concurrentPtr; i++ {
		var urlPtr string = fmt.Sprintf("%s%s:%d%s",*protocolPtr, *domainPtr,*portPtr,*pathPtr);
		if *appendiPtr {
			urlPtr = fmt.Sprintf("%s%d", urlPtr, i)
		}
		urls = append(urls, urlPtr)
		
		//urls = append(urls, fmt.Sprintf("%s%d", "http://ct-lin-1.pgitech.local:5555/", i))
		//urls = append(urls, fmt.Sprintf("%s", "http://ct-lin-1.pgitech.local/"))
	}

	results := asyncHttpGets(urls)
	
	f, err := os.Create(*filenamePtr)
    check(err)	
    w := bufio.NewWriter(f)
     
	for _, result := range results {
		var messPtr string
		if result != nil && result.response != nil {
			messPtr  = fmt.Sprintf("millis=%d %s status: %s\n", result.receivedat.Sub(t0)/time.Millisecond, result.url,
				result.response.Status)
		} else if result != nil {
			messPtr = fmt.Sprintf("millis=%d %s with error : %s\n", result.receivedat.Sub(t0)/time.Millisecond, result.url,
				result.err)
		} else {
			messPtr = "nil result\n";
		}
		w.WriteString(messPtr);
	}
	w.Flush()
	fmt.Printf("go look at file %s\n", *filenamePtr)
}
