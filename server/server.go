package main

import (
	//libvirt "github.com/libvirt/libvirt-go"
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
	file, _ := os.Open("json/rabbitmq.json")
	decoder := json.NewDecoder(file)
	err := decoder.Decode(s)
	CheckError(err)
	log.Println("DB URL:", s.Url)
	log.Println("DB Username:", s.Username)
	log.Println("DB Password:", s.Password)
	file.Close()

}

//func RefreshDomain(conn *libvirt.Connect) {
//	tmpDoms, err := conn.ListAllDomains(libvirt.CONNECT_LIST_DOMAINS_ACTIVE)
//	CheckError(err)
//	doms = tmpDoms
//}
//
func CheckError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

var influx DB
var tool OpenStackTool

//var doms []libvirt.Domain
//var VMs []instance
//var tmp []instance

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
	log.Println(InsertData)
}

func main() {
	log.Println(" DeBug1 ")
	influx = DB{}
	influx.Init()
	log.Println(" DeBug2 ")
	tool = OpenStackTool{}
	tool.Init(&influx)
	log.Println(" DeBug3 ")

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
		tool.CheckStart()
		log.Println("Received a mesage: ", string(d.Body))
		InsertToInflux(string(d.Body))
	}

}
