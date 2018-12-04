собрать
```
GOOS=linux GOARCH=mipsle go build
```
скопировать в /opt/bin

создать /opt/etc/init.d/S35rkn-redirect
```
#!/bin/sh

ENABLED=yes
PROCS=rkn-redirect
ARGS=""
PREARGS=""
DESC=$PROCS
PATH=/opt/sbin:/opt/bin:/opt/usr/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin

. /opt/etc/init.d/rc.func
```

```
chmod +x /opt/etc/init.d/S35rkn-redirect
```

Перенести админку с 80 порта
```
nvram set http_lanport 8080
nvram commit
```


Добавить в /etc/storage/started_script.sh
```
logger -t post_wan_script.sh "Loading ipset modules"
modprobe ip_set
modprobe ip_set_hash_ip
modprobe xt_set

logger -t rkn "Create ipset tables"
ipset create rkn hash:ip maxelem 16777216 family inet
ipset create rkn6 hash:ip maxelem 16777216 family inet6
```

Добавить в /etc/storage/post_iptables_script.sh
```
PROXY_PORT="9040"

iptables -t nat -I PREROUTING 1 -m set --match-set rkn src -p tcp --syn -j REDIRECT --to-ports $PROXY_PORT
iptables -t nat -I PREROUTING 1 -m set --match-set rkn dst -p tcp --syn -j REDIRECT --to-ports $PROXY_PORT
ip6tables -t nat -I PREROUTING 1 -m set --match-set rkn src -p tcp --syn -j REDIRECT --to-ports $PROXY_PORT
ip6tables -t nat -I PREROUTING 1 -m set --match-set rkn dst -p tcp --syn -j REDIRECT --to-ports $PROXY_PORT
```

Добавить в /etc/storage/post_wan_script.sh
```
if [ "$1" = "up" ]; then
    logger -t post_wan_script.sh "Starting tor"
    /opt/etc/init.d/S35tor start
    /opt/etc/init.d/S35rkn-redirect start
else
    logger -t post_wan_script.sh "Stopping tor"
    /opt/etc/init.d/S35tor stop
    /opt/etc/init.d/S35rkn-redirect stop
fi
```

установить Tor. Настройки тора /opt/etc/tor/torrc
```
User admin
TransPort 0.0.0.0:9040
DNSPort 9053
SOCKSPort 9050
ExitNodes {de}
DataDirectory /opt/var/lib/tor
ExitPolicy reject *:*
ExitPolicy reject6 *:*
Log notice syslog
```

Добавить в /etc/storage/dnsmasq/hosts
```
192.168.1.1 blackhole.beeline.ru
```
