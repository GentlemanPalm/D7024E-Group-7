#!/bin/bash
./bootstrap_addr.sh
/usr/local/go/bin/go run main.go --bsip $(cat bootstrap_host)
