# mikoi

[![Build Status](https://travis-ci.org/nabeken/mikoi.svg)](https://travis-ci.org/nabeken/mikoi)

mikoi is a [HAProxy's proxy protocol](http://www.haproxy.org/download/1.5/doc/proxy-protocol.txt) enabler for command line tools.

It is designed to work to monitor ProxyProtocol enabled server with usual monitoring plugins such as nagios-plugins and sensu-plugins.

In normal situation, existing monitoring plugins can not monitor ProxyProtocol enabled server without modifications.

mikoi launches a proxy server lisening to an [ephemeral port](http://www.ncftp.com/ncftpd/doc/misc/ephemeral_ports.html) (dynamically allocated port) and forks a plugin with passing the port number to let the pluging connect to that port.

mikoi adds a ProxyProtocol header to traffic comes from a plugin and proxies to real server.

```text
        +----------+
  +---> |  plugin  | forked by mikoi
  |     +----------+
  |       /|\   |
  |        |    |
  |        |   \|/
  |      +---------+            +----------+
  |      |         | <--------- |          |
  +----- |  mikoi  |            |  server  | (proxy protocol enabled)
         |         | ---------> |          |
         +---------+ w/ header  +----------+
```

## Installation

Download from [releases](https://github.com/nabeken/mikoi/releases).

Or

```sh
go get -u github.com/nabeken/mikoi
```

## Usage

```sh
$ mikoi -h
Usage:
  mikoi [OPTIONS]

Application Options:
  -H, --hostname=   host name
  -p, --port=       port number
  -t, --timeout=    connection times out (10s)
  -V, --verbose     verbose (false)
  -P, --proxyproto  use ProxyProto (true)

Help Options:
  -h, --help        Show this help message
```

```sh
$ mikoi \
  -H smtp.example.com \
  -p 25 \
  -- /usr/lib/nagios/plugins/check_smtp -H 127.0.0.1 -p {} -w 0.5 -c 1.0
```

`{}` will be replaced with an ephemeral port that mikoi is listening to.

If you omit command line arguments for plugin, mikoi runs in proxy server mode:

```sh
$ mikoi -V -H smtp.example.com -p 25
cmd args: []
mikoi is now running as proxy server mode
mikoi is now listening to 63568
mikoi is launching server

// You can connect to 63568 by any clients
$ telnet 127.0.0.1 63568
```
