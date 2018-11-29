package main

import (
    "os/exec"
    "net/http"
    "flag"
    "log"
    "log/syslog"
    //"html"
    "strings"
    "net"
    "fmt"
)

var listenIP, listenPort, dnsmasqList, ipsetList, ipset6List, adminPage string
var sysLog bool

func init() {
    flag.StringVar(&listenIP, "ip", "192.168.1.1", "Listen IP")
    flag.StringVar(&listenPort, "port", "80", "Listen port")
    flag.StringVar(&ipsetList, "ipset-list-v4", "rkn", "IPSET IPv4 list name")
    flag.StringVar(&ipset6List, "ipset-list-v6", "rkn6", "IPSET IPv6 list name")
    flag.StringVar(&adminPage, "admin-page", "http://192.168.1.1:8080", "Redirect to admin page")
    flag.BoolVar(&sysLog, "syslog", false, "Log to syslog")
    flag.Parse()

    if sysLog {
        if w, err := syslog.New(syslog.LOG_NOTICE, "rkn-redirect"); err == nil {
            log.SetOutput(w)
        } else {
            log.Printf("Can't switch log output to syslog: %s", err)
        }
    }
}

func main() {
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        urlKey := r.URL.Query()["url"]

        if urlKey != nil && len(urlKey) > 0 {
            url := urlKey[0]
            host := hostFromUrl(url)

            addToIpsetList(host)

            http.Redirect(w, r, "http://" + url, 302)
        } else {
            log.Printf("Just redirect to %s", adminPage)
            http.Redirect(w, r, adminPage, 302)
        }
    })

    log.Printf("Starting web server on %s:%s", listenIP, listenPort)
    log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%s", listenIP, listenPort), nil))
}

func addToIpsetList(host string) {
    if ips, err := net.LookupHost(host); err == nil {
        for _, ip := range ips {
            var list string

            if isIPv4(ip) {
                list = ipsetList
            } else {
                list = ipset6List
            }

            log.Printf("Add %s to ipset list %s", ip, list)

            err := exec.Command("ipset", "-A", list, ip).Run()

            if err != nil {
                log.Printf("Can't add to ipset list %s: %s", list, err)
            }
        }
    } else {
        log.Printf("Can't resolv %s: %s", host, err)
    }
}

func isIPv4(ip string) bool {
    return net.ParseIP(ip).To4() != nil
}

func hostFromUrl(s string) string {
    return strings.Split(s, "/")[0]
}

