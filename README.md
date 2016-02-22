# goasyncwebclient
go language asynchronous web client. Useful for simulating tens of thousands of simultaneous web users to exercise or break a web server.
## Downloads
[Windowsx86-64](downloads/windows86-64/goasyncwebclient.exe) 
[Ubuntux86-64](downloads/ubuntux86-64/goasyncwebclient)

## Description
This client will create up to c simultaneous connections
(separate tcp connections with unique client port numbers and blank initial cookie state)
to a defined url (or variant of a url with a unique index number appended per connection) and asynchronously request a response.
It will then read fully each response
(again in parallel) timing each one and logging the response time and any server supplied cookies to a file specified by ofile.

Usage of goasyncwebclient.exe:
*   -acktimeout int
        maximum time in milliseconds to receive some acknowledgement (headers) from server (default 2000)
*  -appendi
        true if the connection index should be appended to url
*  -c int
        concurrency e.g simultaneous connections number 7000 (default 1)
*  -domain string
        domain(e.g. host name) (default "127.0.0.1")
*  -minc int
        minimum concurrency to start from (default 1)
*  -n int
        total number of calls processed. should be a multiple of c (default 1)
*  -ofile string
        filename to write output to (default "/tmp/goout")
*  -path string
        path should start with a / and maybe end with one if also using the -appendi flag (default "/")
*  -port int
        port typically 80 or 443 (default 80)
*  -protocol string
        protocol prefix to domain e.g. http:// (default "http://")
*  -rampToFail
        default true and if true will ramp up the connections until first fail (default true)

## Examples
Suppose you have an ngix server listening on port 80 on a linux local host you could invoke goasyncwebclient with no arguments as:-
```
 ./goasyncwebclient

trialConcurrent=1 go look at file /tmp/goout  data speed=546.245283 KiB/s request rate=913.979035 Request/s
Test ended
```
To see the detail of the request response let's cat the log file.
```
 cat /tmp/goout
ackmillis=0 recmillis=1 http://127.0.0.1:80/ status: 200 OK
```
We see the url and the http status code 200. We see that the response was in 1 millisecond or less.
Now lets try with 600 simulated users.
```
 ./goasyncwebclient -c=600 -minc=600 -n=600

trialConcurrent=600 go look at file /tmp/goout  data speed=3028.152633 KiB/s request rate=5066.712902 Request/s
Test ended
```
Its all still good and the response time is still good as we can see by looking at the last few lines in the log.
```
tail -n 4 /tmp/goout
ackmillis=80 recmillis=80 http://127.0.0.1:80/ status: 200 OK

ackmillis=75 recmillis=75 http://127.0.0.1:80/ status: 200 OK
```
However this was actually at the limit which we see as we can let the program try 20% more connections each time starting an minc=300 and maximum c=9000. (It will stop as soon as there is an error or slowresponse).
```
./goasyncwebclient -c=9000 -minc=300 -n=9000

trialConcurrent=300 go look at file /tmp/goout  data speed=3731.619166 KiB/s request rate=6243.754944 Request/s
trialConcurrent=360 go look at file /tmp/goout  data speed=3883.965221 KiB/s request rate=6498.660761 Request/s
trialConcurrent=432 go look at file /tmp/goout  data speed=3409.761651 KiB/s request rate=5705.222109 Request/s
trialConcurrent=518 go look at file /tmp/goout  data speed=3407.418378 KiB/s request rate=5701.301339 Request/s
trialConcurrent=621 go look at file /tmp/goout  data speed=3154.072111 KiB/s request rate=5277.401702 Request/s
was an ERROR

Test ended
```
This was a disappointing limit as the speed was still good. It turned out that a default configuration of nginx was the problem and adding more configuration to nginx.conf fixed the problem.
```
events {
    worker_connections  30000;
}
```
Now a repeat test (starting at 7000 connections) gave:-
```
 ./goasyncwebclient -c=30000 -minc=7000 -n=30000
...........
trialConcurrent=7000 go look at file /tmp/goout  data speed=2243.497760 KiB/s request rate=3753.826318 Request/s.....
trialConcurrent=8400 go look at file /tmp/goout  data speed=2857.678926 KiB/s request rate=4781.475849 Request/s......
trialConcurrent=10080 go look at file /tmp/goout  data speed=2762.223186 KiB/s request rate=4621.759056 Request/s
was TOO SLOW (> acktimeout=2000)

Test ended
```
So it no longer fails with an error but merely gets too slow at 10000 simulated users.

Now to demonstrate the https to an external site we change the protocol, port and add a domain and path and in the results we see this site sets a cookie and is clearly a PHP site.
```
 ./goasyncwebclient -c=1 -minc=1 -n=1 -protocol=https:// -port=443 -domain=www.gladstonebrookes.co.uk -path=/online-claim-form/
.............................
trialConcurrent=1 go look at file /tmp/goout  data speed=24.584313 KiB/s request rate=0.683324 Request/s
Test ended
chris.turner@CT-LIN-1:~/work/bin$ cat /tmp/goout
ackmillis=1245 recmillis=1463 https://www.gladstonebrookes.co.uk:443/online-claim-form/ status: 200 OK
__cfduid=dc746905066c170c8e4f7e485723877a31456146462; Path=/; Domain=gladstonebrookes.co.uk; HttpOnly
PHPSESSID=72ca1440a84b3909ab49eee50e811f07; Path=/
Source=Direct; Path=/; Expires=Wed, 23 Mar 2016 13:07:43 GMT
```
This site is already slow at 1.4 seconds so lets increase the timeout to 4 seconds and see how may users it can take.
```
./goasyncwebclient -c=400 -minc=10 -n=400 -protocol=https:// -port=443 -domain=www.gladstonebrookes.co.uk -path=/online-claim-form/ -acktimeout=4000
...........................................................................
trialConcurrent=10 go look at file /tmp/goout  data speed=90.031498 KiB/s request rate=2.502436 Request/s...................................................................
trialConcurrent=12 go look at file /tmp/goout  data speed=116.106181 KiB/s request rate=3.227185 Request/s.....................................................................
trialConcurrent=14 go look at file /tmp/goout  data speed=132.061706 KiB/s request rate=3.670671 Request/s.........................................................................................
trialConcurrent=16 go look at file /tmp/goout  data speed=118.354553 KiB/s request rate=3.289679 Request/s
was TOO SLOW (> acktimeout=4000)

Test ended
```
So a mere 16 active users can push this sites response times past 4 seconds.
## Caveats
The program truely uses a unique connection and source port for each simulated user and it is easy to run out of open file handles. A default limit for windows might be 7000 and for linux 1000 so some changes of your OS limits may be desirable. E.g. on linux add some lines to /etc/security/limits.conf
```
* soft nofile 100000

* hard nofile 100000
```
Even with such changes the client source port range for the whole machine may be limited to 30000 or so. Also a finished tcp connection is not immediately available for reuse for a couple of minutes so repeated runs may still run out of connections.
For linux clients and servers some changes to /etc/sysctl.conf might be useful.
```
net.ipv4.tcp_max_syn_backlog = 16384
net.core.somaxconn = 16384
net.core.netdev_max_backlog = 16384
net.ipv4.ip_local_port_range = 18000    65535
```


