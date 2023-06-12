package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/goburrow/serial"
)

func main() {
	/* ============================ */
	/* ========== Serial ========== */
	/* ============================ */
	baudrate, err := strconv.Atoi(os.Getenv("SERIAL_BAUDRATE"))
	if err != nil {
		log.Panicln(err)
	}

	port, err := serial.Open(&serial.Config{
		Address:  os.Getenv("SERIAL_PORT"),
		BaudRate: baudrate,
		DataBits: 8,
		StopBits: 1,
		Parity:   "N",
		Timeout:  time.Duration(3) * time.Second,
	})
	defer port.Close()
	if err != nil {
		log.Panicln(err)
	}

	/* ========================== */
	/* ========== MQTT ========== */
	/* ========================== */
	server := fmt.Sprintf("tcp://%s:%s", os.Getenv("MQTT_BROKER"), os.Getenv("MQTT_PORT"))
	opts := mqtt.
		NewClientOptions().
		AddBroker(server).
		SetClientID(os.Getenv("MQTT_CLIENTID")).
		// SetUsername("emqx").
		// SetPassword("public").
		SetKeepAlive(time.Duration(60) * time.Second).
		SetOnConnectHandler(func(client mqtt.Client) {
			log.Printf("Success to Connect %s\n", server)
		}).
		SetConnectionLostHandler(func(c mqtt.Client, err error) {
			log.Printf("Failed to Connect %s\n", server)
		}).
		SetConnectRetry(true).
		SetAutoReconnect(true).
		SetDefaultPublishHandler(func(client mqtt.Client, msg mqtt.Message) {
			log.Printf("From topic: %s Received message: %s\n", msg.Topic(), msg.Payload())
			if _, err := port.Write(msg.Payload()); err != nil {
				log.Panicln(err)
			}
			time.Sleep(1000 * time.Millisecond) // Wait Data
			data := make([]byte, 1024)
			if n, err := port.Read(data); err != nil {
				log.Panicln(err)
			} else {
				client.Publish(os.Getenv("MQTT_SEND_TOPIC"), 1, false, data[:n])
			}
		})
	client := mqtt.NewClient(opts)
	defer client.Disconnect(1000)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Panicln(token.Error())
	} else {
		/*
			| Pub_QoS | Sub_QoS | All_QoS |
			|    0    |    0    |    0    |
			|    0    |    1    |    0    |
			|    0    |    2    |    0    |
			|    1    |    0    |    0    |
			|    1    |    1    |    1    |
			|    1    |    2    |    1    |
			|    2    |    0    |    0    |
			|    2    |    1    |    1    |
			|    2    |    2    |    2    |

			All_QoS = min(Pub_QoS, Sub_QoS)
		*/
		client.Subscribe(os.Getenv("MQTT_RECEIVE_TOPIC"), 1, nil).Wait()
	}
	<-make(chan os.Signal, 1) // ref. https://stackoverflow.com/questions/48872360/golang-mqtt-publish-and-subscribe
}
