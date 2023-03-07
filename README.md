# Internal DNS

## Overview

Fun small project to get used to NS1 API. Go script to add a public DNS record with the system's private IP address, to allow for psuedo-service-discovery and easier addressing of devices on the LAN.

## Building

```bash
go build internal-dns.go
```

## Usage 

Example:

```bash
~ -> ip --brief -4 a s wlp0s20f3
wlp0s20f3        UP             192.168.1.234/24
~ -> ZONE='ns1.work.tjhop.io'
~ -> DOMAIN='test'
~ -> hostname                                                            
ns1-dell
~ -> ./internal-dns -zone "$ZONE"
time="2023-03-07T16:35:34-05:00" level=info msg="Log level set to: info"
time="2023-03-07T16:35:34-05:00" level=info msg="Internal DNS server started" build_date= commit= go_version=go1.20.1 version=
time="2023-03-07T16:35:36-05:00" level=info msg="Internal DNS record set" domain=ns1-dell zone=ns1.work.tjhop.io
time="2023-03-07T16:35:36-05:00" level=info msg="Internal DNS exited" build_date= commit= go_version=go1.20.1 version=
~ -> echo "$(hostname).$ZONE"; dig "$(hostname).$ZONE" A @8.8.8.8 +short
ns1-dell.ns1.work.tjhop.io
192.168.1.234
~ -> ./internal-dns -zone "$ZONE" -domain "$DOMAIN" -log-format json
{"level":"info","msg":"Log level set to: info","time":"2023-03-07T16:37:01-05:00"}
{"build_date":"","commit":"","go_version":"go1.20.1","level":"info","msg":"Internal DNS server started","time":"2023-03-07T16:37:01-05:00","versio
n":""}
{"domain":"test","level":"info","msg":"Internal DNS record set","time":"2023-03-07T16:37:04-05:00","zone":"ns1.work.tjhop.io"}
{"build_date":"","commit":"","go_version":"go1.20.1","level":"info","msg":"Internal DNS exited","time":"2023-03-07T16:37:04-05:00","version":""}
~ -> echo "$DOMAIN.$ZONE"; dig "$DOMAIN.$ZONE" A @8.8.8.8 +short
test.ns1.work.tjhop.io
192.168.1.234
``` 
