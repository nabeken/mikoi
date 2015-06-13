# mikoi

:construction: :construction: :construction:

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
$ mikoi -- /usr/lib/nagios/plugins/check_smtp -H 127.0.0.1 -p {} -w 0.5 -c 1.0
```

`{}` will be replaced with an ephemeral port that mikoi is listening to.
