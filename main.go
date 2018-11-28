package main

import (
    //"github.com/davecgh/go-spew/spew"
    "net/http"
    "flag"
    "log"
    "log/syslog"
    "fmt"
    //"html"
    "os/exec"
    "time"
    "strings"
    "os"
    "bufio"
)

var listenIP, dnsmasqList, ipsetList, adminPage string

func init() {
    flag.StringVar(&listenIP, "ip", "192.168.1.1", "Listen IP")
    flag.StringVar(&dnsmasqList, "dnsmasq-config", "/etc/storage/dnsmasq/rkn-ipset-list.conf", "dnsmasq config file for ipsec list")
    flag.StringVar(&ipsetList, "ipset-list", "rkn", "IPSET list name")
    flag.StringVar(&adminPage, "admin-page", "http://192.168.1.1:8080", "Redirect to admin page")
    flag.Parse()

    if w, err := syslog.New(syslog.LOG_NOTICE, "rkn-redirect"); err == nil {
        log.SetOutput(w)
    } else {
        log.Printf("Can't switch log output to syslog: %s", err)
    }
}

func main() {
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        urlKey := r.URL.Query()["url"]

        if urlKey != nil && len(urlKey) > 0 {
            url := urlKey[0]
            host := hostFromUrl(url)

            if notExistInDnsList(dnsmasqList, host) {
                log.Printf("Add %s to %s", host, ipsetList)
                addToDnsList(dnsmasqList, hostFromUrl(url))
            }

            log.Println("Restart DNS server")
            restartDnsServer()

            time.Sleep(1 * time.Second)
            http.Redirect(w, r, "http://" + url, 302)
        } else {
            log.Printf("Just redirect to %s", adminPage)
            http.Redirect(w, r, adminPage, 302)
        }
    })

    log.Printf("Starting web server on %s:80", listenIP)
    log.Fatal(http.ListenAndServe(listenIP + ":80", nil))
}

func notExistInDnsList(list string, host string) bool {
    file, err := os.Open(list)
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        if scanner.Text() == fmt.Sprintf("ipset=/%s/%s\n", host, ipsetList) {
            return false
        }
    }

    if err := scanner.Err(); err != nil {
        log.Fatal(err)
    }

    return true
}

func addToDnsList(file string, host string) {
    if f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
        defer f.Close()
        if _, err = f.WriteString(fmt.Sprintf("ipset=/%s/%s\n", host, ipsetList)); err != nil {
            log.Printf("Error write to ipset list file: %s", err)
        }
    } else {
        log.Printf("Error open ipset list file: %s", err)
    }
}

func hostFromUrl(s string) string {
    return strings.Split(s, "/")[0]
}

func restartDnsServer() {
    if err := exec.Command("/sbin/restart_dhcpd").Run(); err != nil {
        log.Printf("Error restart DNS server: %s", err)
    }
}

