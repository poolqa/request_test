# Transactions Request Pressure Test
This is a simple tool for Http Pressure test.

You can use it on Windows system, and don't be afraid to encounter run out the port.

It can reuse the connection.

If this tool can help you, please don't forget to star me. :) 

### TODO
This tool is a semifinished product, if you want add some features, you can clone and fix it, or tall me.

### get it
```bash
go get github.com/poolqa/request_test
```
or
```bash
git clone https://github.com/poolqa/request_test
```
### build it
```bash
go build
```
### Run it

```bash
request_test -c 3 -n 5 -u https://www.google.com -w

Thread count: 3
Starting at: 2020/08/03 22:20:40.711046
Finished at: 2020/08/03 22:20:41.067851

performance statistics :
LE  1 Sec: count: 15
-----------------------
Total Count: 15

Use times: 356.8042ms
Tps:       42.040 per/sec
Availability:  100.00%
Failed:  0
Connection Times
                         min            max            avg
Response time:           47.0104ms      116.3257ms     7.755046ms
Transaction time:        47.0104ms      116.3257ms     7.755046ms
Build cli time:          0s             0s             0s

```
### Usage
```bash
request_test -h
Transactions Request Pressure Test/ v0.0.1
Usage: pressure -c concurrency -n requests -t timeLimit -m method -u url -[dh]

Options:
  -c concurrency
        go routine(concurrency) count at same time. (default 1)
  -d    show debug log
  -h    this help
  -i int
        print interval for process request count time. (default 1)
  -m method
        http method name. (default "GET")
  -n requests
        requests at every concurrency
  -r    show those routine's requests detail at report log,
        but it wall use more and more memory.
  -t timeLimit
        timeLimit Seconds to max. to spend on benchmarking, timer is start at all routine wake up.
  -u url
        target url
  -w    waiting all routines stand-by.

```

### Copyright
request_test is completely free. Please mark the source of request_test in your commercial product if possible.
