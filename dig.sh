#!/bin/bash

# Configuration
NUM_REQUESTS=9999                # Total number of DNS requests (default: 100)
QUERY_TYPES=(A AAAA NS PTR MX SOA CNAME TXT SRV NAPTR)    # List of DNS query types to randomize
DOMAIN_LIST=(example.com example.org example.net test.example.com opensource.org eff.org fsf.org apache.org)  # List of base domains
SUBDOMAIN_LIST=(www www2 app log blog mail smtp imap pop3 ns1 ns2 ns3 ns4 ns5)  # List of subdomains
PORT=25353                       # Target port for dig requests
LOCALHOST=127.0.0.1              # Target host for dig requests
FAST_MODE=false                  # Default mode (not fast)

# Parse arguments
while [[ "$1" != "" ]]; do
  case $1 in
    --fast)
      FAST_MODE=true
      ;;
  esac
  shift
done

# Generate a random domain name
function random_domain {
  local base_domain=${DOMAIN_LIST[$RANDOM % ${#DOMAIN_LIST[@]}]}
  local subdomain=${SUBDOMAIN_LIST[$RANDOM % ${#SUBDOMAIN_LIST[@]}]}
  echo "${subdomain}.${base_domain}"
}

# Generate a random PTR query target
function random_ptr {
  echo "$((RANDOM % 256)).$((RANDOM % 256)).$((RANDOM % 256)).$((RANDOM % 256)).in-addr.arpa"
}

# Send DNS requests
for ((i = 0; i < NUM_REQUESTS; i++)); do
  query_type=${QUERY_TYPES[$RANDOM % ${#QUERY_TYPES[@]}]}  # Random query type

  if [ "$query_type" == "PTR" ]; then
    domain=$(random_ptr)  # Use a random PTR target
  else
    domain=$(random_domain)  # Use a random domain name
  fi

  echo "Query #$((i + 1)): dig @$LOCALHOST -p $PORT $domain $query_type"
  dig @$LOCALHOST -p $PORT $domain $query_type +short

  # Add a small random delay
  if [ "$FAST_MODE" == false ]; then
    sleep $(awk "BEGIN {print $RANDOM/32768}")
  fi

done
