#!/bin/bash
rsyslogd
service postfix start

while true; do
  [ -f /var/log/mail.* ] && break
  sleep 1
done

tail -F /var/log/*.log
