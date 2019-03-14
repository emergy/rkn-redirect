# собрать
```
GOOS=linux GOARCH=mipsle GOMIPS=softfloat go build
```

# Padavan
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
ip6tables -t nat -I PREROUTING 1 -m set --match-set rkn6 src -p tcp --syn -j REDIRECT --to-ports $PROXY_PORT
ip6tables -t nat -I PREROUTING 1 -m set --match-set rkn6 dst -p tcp --syn -j REDIRECT --to-ports $PROXY_PORT
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
TransPort 9040
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

# OpenWRT

перенести админку с 80 порта в файле /etc/config/uhttpd

создать /etc/init.d/rkn-redirect
```
#!/bin/sh /etc/rc.common

START=50
STOP=50

USE_PROCD=1

start_service() {
    procd_open_instance
    procd_set_param command /usr/bin/rkn-redirect -syslog
    procd_close_instance
}
```

```
chmod +x /etc/init.d/rkn-redirect
/etc/init.d/rkn-redirect enable
/etc/init.d/rkn-redirect start
```


```
opkg update
opkg install ipset kmod-nf-nat6
```

дефолтный dnsmasq не умеер ipset
```
opkg remove dnsmasq
opkg install dnsmasq-full
```

добавить в /etc/config/dhcp
```
config domain
        option name 'blackhole.beeline.ru'
        option ip '192.168.1.1'
```

/etc/config/firewall
```
config ipset
        option name rkn
        option storage hash
        option match ip
        option maxelem 16777216
        option hashsize 1024
        option enabled 1
        option family ipv4

config ipset
        option name rkn6
        option storage hash
        option match ip
        option maxelem 16777216
        option hashsize 1024
        option enabled 1
        option family ipv6
```

/etc/firewall.user
```
PROXY_PORT="9040"
iptables -t nat -I PREROUTING 1 -m set --match-set rkn src -p tcp --syn -j REDIRECT --to-ports $PROXY_PORT
iptables -t nat -I PREROUTING 1 -m set --match-set rkn dst -p tcp --syn -j REDIRECT --to-ports $PROXY_PORT

ip6tables -t nat -I PREROUTING 1 -m set --match-set rkn6 src -p tcp --syn -j REDIRECT --to-ports $PROXY_PORT
ip6tables -t nat -I PREROUTING 1 -m set --match-set rkn6 dst -p tcp --syn -j REDIRECT --to-ports $PROXY_PORT

iptables -t nat -I OUTPUT 1 -p tcp -m set --match-set rkn dst -j REDIRECT --to-port $PROXY_PORT
ip6tables -t nat -I OUTPUT 1 -p tcp -m set --match-set rkn6 dst -j REDIRECT --to-port $PROXY_PORT
```

```
opkg install tor-geoip
```

/etc/tor/torrc
```
User tor
TransPort 9040
DNSPort 9053
SOCKSPort 9050
ExitNodes {de}
DataDirectory /var/lib/tor
ExitPolicy reject *:*
ExitPolicy reject6 *:*
Log notice syslog
```

```
/etc/init.d/tor enable
/etc/init.d/tor start
```

