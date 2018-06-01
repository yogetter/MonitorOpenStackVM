package main

import (
	"MonitorOpenStackVM/tool"
	"encoding/json"
	"github.com/streadway/amqp"
	"log"
	"os"
	"strings"
)

type server struct {
	Url      string
	Username string
	Password string
}

func (s *server) init() {
	//read config
	file, _ := os.Open("../json/rabbitmq.json")
	decoder := json.NewDecoder(file)
	err := decoder.Decode(s)
	CheckError(err)
	log.Println("DB URL:", s.Url)
	log.Println("DB Username:", s.Username)
	log.Println("DB Password:", s.Password)
	file.Close()

}

func CheckError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func ChInit(ch *amqp.Channel) <-chan amqp.Delivery {
	err := ch.ExchangeDeclare(
		"VMusage", // name
		"fanout",  // type
		true,      // durable
		false,     // auto-deleted
		false,     // internal
		false,     // no-wait
		nil,       // arguments
	)
	CheckError(err)
	q, err := ch.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	CheckError(err)
	err = ch.QueueBind(
		q.Name,    // queue name
		"",        // routing key
		"VMusage", // exchange
		false,
		nil)

	CheckError(err)
	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	CheckError(err)
	return msgs
}
func DataSplit(VmUsage string) []string {
	return strings.Split(VmUsage, ";")

}

func InsertToInflux(VmUsage string) {
	InsertData := DataSplit(VmUsage)
	influx.InsertVmInfo(InsertData)
}

var influx DB
var t tool.OpenStackTool
var check CheckInstance

func main() {
	influx = DB{}
	influx.Init()
	t = tool.OpenStackTool{}
	t.Init()

	config := server{}
	config.init()
	conn, err := amqp.Dial("amqp://" + config.Username + ":" + config.Password + "@" + config.Url)
	CheckError(err)
	defer conn.Close()

	ch, err := conn.Channel()
	CheckError(err)
	defer conn.Close()
	defer ch.Close()
	msgs := ChInit(ch)
	log.Println(" [*] Waiting for messages. To exit press CTRL+C")
	for d := range msgs {
		check.CheckStart(&t, &influx)
		log.Println("Received a mesage: ", string(d.Body))
		InsertToInflux(string(d.Body))
	}

}
