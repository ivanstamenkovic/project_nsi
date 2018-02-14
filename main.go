package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/astaxie/beego"
	"github.com/ivanstamenkovic/project_nsi/controllers"
	_ "github.com/ivanstamenkovic/project_nsi/routers"

	"github.com/miekg/dns"
)

func parseQuery(m *dns.Msg) {
	for _, q := range m.Question {
		switch q.Qtype {
		case dns.TypeA:
			beego.Debug(q.Name[:(len(q.Name) - 1)])
			ipAddress, err := controllers.DnsClient.Get(q.Name[:(len(q.Name) - 1)]).Result()

			beego.Debug(err)
			if err == nil {
				rr, errRR := dns.NewRR(fmt.Sprintf("%s A %s", q.Name, ipAddress))
				if errRR == nil {
					m.Answer = append(m.Answer, rr)
				}
			}
		}
	}
}

func handleDnsRequest(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	m.Compress = false
	switch r.Opcode {
	case dns.OpcodeQuery:
		parseQuery(m)
	}

	w.WriteMsg(m)
}

func startDns() {

	dns.HandleFunc(".", handleDnsRequest)

	port := 53
	server := &dns.Server{Addr: ":" + strconv.Itoa(port), Net: "udp"}
	beego.Debug("DNS server running")
	server.ListenAndServe()
	defer server.Shutdown()
}

func main() {

	controllers.DnsClient.Set("test.idee.com", "127.0.0.1", 10*time.Minute)

	go startDns()

	beego.Run()

}
