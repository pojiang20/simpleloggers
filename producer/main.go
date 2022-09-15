package main

import (
	"fmt"
	"github.com/hashicorp/consul/api"
	"github.com/labstack/gommon/log"
	"io"
	"net/http"
	"strings"
	"time"
)

func main() {
	log.Printf("Starting producer\n")
	serviceKey := "service/distributed-logger/leader"

	config := api.DefaultConfig()
	config.Address = "consul:8500"

	client, err := api.NewClient(config)
	if err != nil {
		log.Fatalf("client err: %v", err)
	}

	msgID := 1
	for {
		kv, _, err := client.KV().Get(serviceKey, nil)
		if err != nil {
			log.Fatalf("kv acquire err: %v", err)
		}

		if kv != nil && kv.Session != "" {
			leaderHostname := string(kv.Value)
			sendMsg(leaderHostname, msgID)
			msgID++
		}

		time.Sleep(5 * time.Second)
	}
}

func sendMsg(hostname string, msgID int) {
	msg := fmt.Sprintf("Message: %v", msgID)
	log.Printf("Sending message %v\n", msgID)
	resp, err := http.Post(fmt.Sprintf("http://%v:3000/api/v1/log", hostname), "text/plain", strings.NewReader(msg))
	if err != nil {
		log.Printf("http post err: %v", err)
		return
	}

	defer resp.Body.Close()
	_, err = io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		log.Printf("status not OK: %v", resp.StatusCode)
		return
	}

	log.Printf("msg sent OK")
}
