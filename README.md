# NS Checker

This is a simple tool to check the availability of a domain that look similar to the given name. Searching for typos in the domain name.

Also i provides a simple DNS listener that always returns the same IP address for any query and logs the query message.

## Usage

```bash
go run . help
```

## DNS Typo Checker

Edit the `typo-tlds.txt` file to add the typo you want to check.

```bash
go run . check
```

This will check the typo of the domain name in the `typo-tlds.txt` file.
It will print the domain name and the typo domain name.
It will also log the query message to a file `dns_typo_checker.log` and not registered domains to `dns_typo_checker_not_registered.log`.

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

## DNS Listener

```bash
go run . listen
```

This will start a DNS server on port 25353, you can use `dig` to query the server.

```bash
go run . listen 5353
```

This will start a DNS server on port 5353, you can use `dig` to query the server.

It will always response a A record with the IP `127.0.0.1` to the query.
It will also print the query message.
It will also log the query message to a file `dns_listener.log`.

```bash
ns-checker/source/ns-checker via 🐹 v1.23.5 via 💎 v3.0.0 
❯ dig @127.0.0.1 -p 25353 example.com
;; Warning: Message parser reports malformed message packet.

; <<>> DiG 9.20.4-3-Debian <<>> @127.0.0.1 -p 25353 example.com
; (1 server found)
;; global options: +cmd
;; Got answer:
;; ->>HEADER<<- opcode: QUERY, status: NOERROR, id: 4095
;; flags: qr rd ra; QUERY: 1, ANSWER: 1, AUTHORITY: 0, ADDITIONAL: 1

;; QUESTION SECTION:
;example.com.   IN A

;; ANSWER SECTION:
.   0 CLASS1232 OPT 10 8 qqkMOpAvhZo=

;; ADDITIONAL SECTION:
example.com.  300 IN A 127.0.0.1

;; Query time: 0 msec
;; SERVER: 127.0.0.1#25353(127.0.0.1) (UDP)
;; WHEN: Wed Jan 22 20:31:30 CET 2025
;; MSG SIZE  rcvd: 68
```

```bash
ns-checker/source/ns-checker via 🐹 v1.23.5 via 💎 v3.0.0 
❯ dig +tcp @::1 -p 25353 example.com    
;; Warning: Message parser reports malformed message packet.

; <<>> DiG 9.20.4-3-Debian <<>> +tcp @::1 -p 25353 example.com
; (1 server found)
;; global options: +cmd
;; Got answer:
;; ->>HEADER<<- opcode: QUERY, status: NOERROR, id: 30715
;; flags: qr rd ra; QUERY: 1, ANSWER: 1, AUTHORITY: 0, ADDITIONAL: 1

;; QUESTION SECTION:
;example.com.   IN A

;; ANSWER SECTION:
.   0 CLASS1232 OPT 10 8 sB394KJa8RY=

;; ADDITIONAL SECTION:
example.com.  300 IN A 127.0.0.1

;; Query time: 0 msec
;; SERVER: ::1#25353(::1) (TCP)
;; WHEN: Wed Jan 22 20:31:38 CET 2025
;; MSG SIZE  rcvd: 68

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

Copy the `env.sample` environment file to `.env` and adjust as needed:

```text
# DNS server port (default: 25353)
DNS_PORT=25353

# Cache TTL in seconds (default: 600)
CACHE_TTL=600

# Maximum number of workers (default: auto-calculated)
MAX_WORKERS=12

# Rate limit per second (default: 100000)
RATE_LIMIT=100000

# Rate limit burst (default: 1000)
RATE_BURST=1000
```
