# goasyncwebclient
go language asynchronous web client. Useful for simulating tens of thousands of simultaneous web users to exercise or break a web server.

This client will create up to c simultaneous connections
(separate tcp connections with unique client port numbers and blank initial cookie state)
to a defined url (or variant of a url with a unique index number appended per connection) and asynchronously request a response.
It will then read fully each response
(again in parallel) timing each one and logging the response time and any server supplied cookies to a file specified by ofile.

Usage of goasyncwebclient.exe:
   -acktimeout int
        maximum time in milliseconds to receive some acknowledgement (headers) from server (default 2000)
  -appendi
        true if the connection index should be appended to url
  -c int
        concurrency e.g simultaneous connections number 7000 (default 1)
  -domain string
        domain(e.g. host name) (default "127.0.0.1")
  -minc int
        minimum concurrency to start from (default 1)
  -n int
        total number of calls processed. should be a multiple of c (default 1)
  -ofile string
        filename to write output to (default "/tmp/goout")
  -path string
        path should start with a / and maybe end with one if also using the -appendi flag (default "/")
  -port int
        port typically 80 or 443 (default 80)
  -protocol string
        protocol prefix to domain e.g. http:// (default "http://")
  -rampToFail
        default true and if true will ramp up the connections until first fail (default true)


