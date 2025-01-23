# NS Checker

A tool for domain name security analysis with two main components:

1. DNS Typo Checker - Identifies similar domain names for typosquatting detection
2. DNS Listener - A configurable DNS sinkhole for monitoring DNS queries

## Features

### DNS Listener Features

1. DNS Server Capabilities:
   - UDP DNS server with configurable port
   - Worker pool with automatic scaling
   - Request buffering and rate limiting
   - Message parsing and validation

2. Infrastructure:
   - Health monitoring system
   - Caching with TTL management
   - Resource optimization
   - Graceful operations handling
   - Security validations

### DNS Typo Checker Features

1. Domain Analysis:
   - Character manipulation detection
   - TLD variation checking
   - Multiple domain processing
   - WHOIS data integration

2. Operations:
   - Structured logging system
   - Real-time reporting
   - Configuration management
   - Batch processing support

## Quick Start

### Installation

```bash
go run . help
```

### DNS Typo Checker

Edit the `typo-tlds.txt` file to add the typo you want to check.

```bash
go run . check
```

This will check the typo of the domain name in the `typo-tlds.txt` file.
It will print the domain name and the typo domain name.
It will also log the query message to a file `[Date]_dns_typo_checker.log` and not registered domains to `[Date]_dns_typo_checker_not_registered.log` into the `logs` directory.

The console output will be like this:

```bash
Checking typos for domain: nsone.net
Valid DNS found for typo: sone.net
Valid DNS found for typo: snone.net
Valid DNS found for typo: none.net
No DNS record for: nosne.net
Valid DNS found for typo: nsne.net
No DNS record for: nsnoe.net
Valid DNS found for typo: nsoe.net
No DNS record for: nsoen.net
No DNS record for: nson.net
Valid DNS found for typo: nsone.com
Valid DNS found for typo: nsone.org
No DNS record for: nsone.ne
Valid DNS found for typo: nsone.co
```

### DNS Listener

```bash
go run . listen
```

This will start a DNS server on port 25353, you can use `dig` to query the server.

```bash
go run . listen 5353
```

The console output will be like this:

```bash
Initializing with health check port: 8088

╔═══════════════════════════════════╗
║         DNS Listener Active       ║
╚═══════════════════════════════════╝

=== DNS Listener Configuration ===
► Port: 25353
► Worker Pool Size: 4 workers
► Request Channel Buffer: 80 requests
► Rate Limit: 100000 requests/second (burst: 1000)
► DNS Message Buffer Size: 512 bytes
► Cache TTL: 30m0s
► Cache Cleanup Interval: 1m0s
===================================
UDP server listening on 0.0.0.0:25353
TCP server listening on 0.0.0.0:25353
```

And Runtime Statistics will be like this:

```bash
=== Runtime Statistics ===
► System Health:
  • CPU Usage: 0.0%
  • Memory Usage: 4.9%
  • Uptime: 30s
  • Last GC: 482727h18m ago
  • GC Pause: 0
► Cache:
  • Size: 2 entries (104 B)
  • Hit Ratio: 0.0% (0/2)
  • Evictions: 0
► Processing:
  • Channel Load: 0/80 (0% utilized)
  • Total Requests: 2 (0.1/sec avg)
  • Goroutines: 0
  • Heap Usage: 0 B
► Performance:
  • Request Rate: 0.1/sec current
  • Response Times:
    - Avg: 6.00ms
    - P95: 9.00ms
    - P99: 9.00ms
► Rate Limiting:
  • Limited Requests: 0
  • Active Clients: 2 (0% of limit)
  • Burst Usage: 0.1%
► Validation:
  • Success Rate: 100.0% (2/2 total)
  • Invalid Queries: 0
  • Invalid Responses: 0
=========================
```

This will start a DNS server on port 5353, you can use `dig` to query the server.

It will always response a A record with the IP `127.0.0.1` to the query.
It will also print the query message.
It will also log the query message to a file `[Date]_dns_listener.log` in the `logs` directory.

```bash
ns-checker on  main [✘!+⇡] via 🐹 v1.23.5 via 💎 v3.0.0 
❯ dig @127.0.0.1 -p 25353 example.org SOA

; <<>> DiG 9.20.4-3-Debian <<>> @127.0.0.1 -p 25353 example.org SOA
; (1 server found)
;; global options: +cmd
;; Got answer:
;; ->>HEADER<<- opcode: QUERY, status: NOERROR, id: 11884
;; flags: qr rd ad; QUERY: 1, ANSWER: 0, AUTHORITY: 0, ADDITIONAL: 1
;; WARNING: recursion requested but not available

;; OPT PSEUDOSECTION:
; EDNS: version: 0, flags:; udp: 1232
; COOKIE: 1ff27d473bdecb10 (echoed)
;; QUESTION SECTION:
;example.org.   IN SOA

;; Query time: 8 msec
;; SERVER: 127.0.0.1#25353(127.0.0.1) (UDP)
;; WHEN: Sat Jan 25 16:13:31 CET 2025
;; MSG SIZE  rcvd: 52
```

```bash
ns-checker on  main [✘!+⇡] via 🐹 v1.23.5 via 💎 v3.0.0 
❯ dig +tcp @::1 -p 25353 example.org AAAA

; <<>> DiG 9.20.4-3-Debian <<>> +tcp @::1 -p 25353 example.org AAAA
; (1 server found)
;; global options: +cmd
;; Got answer:
;; ->>HEADER<<- opcode: QUERY, status: NOERROR, id: 12071
;; flags: qr rd ad; QUERY: 1, ANSWER: 0, AUTHORITY: 0, ADDITIONAL: 1
;; WARNING: recursion requested but not available

;; OPT PSEUDOSECTION:
; EDNS: version: 0, flags:; udp: 1232
; COOKIE: c17e3e69f9d52ece (echoed)
;; QUESTION SECTION:
;example.org.   IN AAAA

;; Query time: 12 msec
;; SERVER: ::1#25353(::1) (TCP)
;; WHEN: Sat Jan 25 16:14:25 CET 2025
;; MSG SIZE  rcvd: 52
```

`[Date]_dns_listener.log` file will be like this:

```text
Created response for 127.0.0.1:51775 (52 bytes) [::1]
Transaction ID: 2f27
Flags: 0120
Questions: 1
Question: example.org
Type: AAAA
Class: IN
Raw Query (Hex):
00000000  2f 27 01 20 00 01 00 00  00 00 00 01 07 65 78 61  |/'. .........exa|
00000010  6d 70 6c 65 03 6f 72 67  00 00 1c 00 01 00 00 29  |mple.org.......)|
00000020  04 d0 00 00 00 00 00 0c  00 0a 00 08 c1 7e 3e 69  |.............~>i|
00000030  f9 d5 2e ce                                       |....|

Created response for [::1]:44685 (52 bytes)
[2025-01-25 15:14:59.642] [UDP] Client: 127.0.0.1:51775
Protocol: UDP
Client IP: 127.0.0.1
Transaction ID: cb17
Flags: 0120
Questions: 1
Question: example.org
Type: SOA
Class: IN
Raw Query (Hex):
00000000  cb 17 01 20 00 01 00 00  00 00 00 01 07 65 78 61  |... .........exa|
00000010  6d 70 6c 65 03 6f 72 67  00 00 06 00 01 00 00 29  |mple.org.......)|
00000020  04 d0 00 00 00 00 00 0c  00 0a 00 08 62 aa cc 22  |............b.."|
00000030  7f ca d3 c8                                       |....|

Created response for 127.0.0.1:51775 (52 bytes)
```

## Build & Run

You can use the Makefile to build and run the application:

### Build Options

```bash
make build-linux-amd64
make build-linux-armv7
```

### Run

```bash
make run
```

and test with `dig`

```bash
for i in {1..10000}; do echo "Iteration $i"; dig @127.0.0.1 -p 25353 example.com SOA; done
```

or

```bash
dig @localhost -p 25353 example.com
```

## Docker

```bash
docker build . -t ns-checker:latest
docker run -d -p 25353:25353/udp ns-checker:latest
```

or use the `docker-compose.yml` file

```bash
docker-compose up --build -d --remove-orphans
```

## Environment Configuration

```bash
# DNS Listener Server Configuration
export DNS_LISTENER_PORT=25353                  # Main DNS server port (UDP/TCP)
export DNS_LISTENER_HEALTH_PORT=8080            # Health check server port
export DNS_LISTENER_TCP_ENABLED=true            # Enable TCP protocol support
export DNS_LISTENER_RESPONSE_IP=127.0.0.1       # Default response IP address
export DNS_LISTENER_RESPONSE_TTL=300            # TTL for DNS responses in seconds

# Performance Configuration
export DNS_LISTENER_MAX_WORKERS=8               # Maximum number of worker goroutines
export DNS_LISTENER_RATE_LIMIT=100000           # Requests per second limit
export DNS_LISTENER_RATE_BURST=1000             # Burst capacity for rate limiting

# Cache Configuration
export DNS_LISTENER_CACHE_TTL=1800              # Cache TTL in seconds
export DNS_LISTENER_CLEANUP_INTERVAL=60         # Cache cleanup interval in seconds

# Logging Configuration
export DNS_LISTENER_LOGS_DIR=./logs             # Directory for log files
export DNS_LISTENER_LOG_FILE=dns_listener.log   # Main log file name
export DNS_LISTENER_DEBUG_LEVEL=info            # Debug level (debug|info|warn|error)

# Metrics Configuration
export DNS_LISTENER_METRICS_ENABLED=true        # Enable metrics collection
```

Docker environment configuration:

```bash
docker run -d -p 25353:25353/udp -e DNS_LISTENER_PORT=25353 -e DNS_LISTENER_HEALTH_PORT=8080 -e DNS_LISTENER_MAX_WORKERS=4 -e DNS_LISTENER_CACHE_TTL=1800 -e DNS_LISTENER_CLEANUP_INTERVAL=60 -e DNS_LISTENER_RATE_LIMIT=100000 -e DNS_LISTENER_RATE_BURST=1000 -e DNS_LISTENER_LOGS_DIR=./logs -e DNS_LISTENER_LOG_FILE=dns_listener.log ns-checker:latest
```

Docker Compose environment configuration:

```bash
mv .env.example .env
docker-compose up --build -d --remove-orphans
```
