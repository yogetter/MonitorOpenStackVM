package main

import (
	"encoding/json"
	"github.com/streadway/amqp"
	"github.com/yogetter/libvirt-go"
	"log"
	"os"
	"time"
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

func RefreshDomain(conn *libvirt.Connect) {
	tmpDoms, err := conn.ListAllDomains(libvirt.CONNECT_LIST_DOMAINS_ACTIVE)
	CheckError(err)
	doms = tmpDoms
}

func CheckError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func GetVmStats(VM *instance, dom *libvirt.Domain, conn *libvirt.Connect, VmQueue chan *instance) {
	domsPoint := make([]*libvirt.Domain, 1)
	domsPoint[0] = dom
	domStats, err := conn.GetAllDomainStats(domsPoint, 0, 0)
	CheckError(err)
	VM.InitAllValue(dom, domStats)
	VmQueue <- VM
	domStats[0].Domain.Free()
}

func InitVmInfo(conn *libvirt.Connect, VmQueue chan *instance) {
	VMs = make([]instance, len(doms))
	tmp = make([]instance, len(doms))
	for i, dom := range doms {
		go GetVmStats(&VMs[i], &dom, conn, VmQueue)
		tmp[i] = *<-VmQueue
		VMs[i].dom.Free()
	}
}

func UpdateVmInfo(conn *libvirt.Connect, VmQueue chan *instance) bool {
	for i, dom := range doms {
		go GetVmStats(&VMs[i], &dom, conn, VmQueue)
		VMs[i] = *<-VmQueue
		//Ensure VM's ID is same
		if VMs[i].Id == tmp[i].Id {
			VMs[i].SetAllValue(tmp[i])
		} else {
			break
			return true
		}
		//log.Println(VMs[i].GetValue())
		VMs[i].dom.Free()
	}
	return false
}

func SendDataToServer() {
	config := server{}
	config.init()
	conn, err := amqp.Dial("amqp://" + config.Username + ":" + config.Password + "@" + config.Url)
	CheckError(err)
	defer conn.Close()
	ch, err := conn.Channel()
	CheckError(err)
	defer ch.Close()
	err = ch.ExchangeDeclare(
		"VMusage", // name
		"fanout",  // type
		true,      // durable
		false,     // auto-deleted
		false,     // internal
		false,     // no-wait
		nil,       // arguments
	)
	CheckError(err)
	for _, VM := range VMs {
		data := VM.GetValue()
		for _, body := range data {
			err = ch.Publish(
				"VMusage", // exchange
				"",        // routing key
				false,     // mandatory
				false,     // immediate
				amqp.Publishing{
					ContentType: "text/plain",
					Body:        []byte(body),
				})
			CheckError(err)
			log.Print(" [x] Sent ", body)
		}
	}
}

var doms []libvirt.Domain
var VMs []instance
var tmp []instance

func main() {
	for {
		VmQueue := make(chan *instance)
		conn, err := libvirt.NewConnect("qemu:///system")
		CheckError(err)
		log.Println("Start collect VM's information")
		RefreshDomain(conn)
		InitVms := len(doms)
		log.Println("init VM's information:", InitVms)
		InitVmInfo(conn, VmQueue)
		time.Sleep(60 * time.Second)
		log.Println("Wait 60s for VM's information update")
		RefreshDomain(conn)
		log.Println("update VM's information:", len(doms))
		if InitVms != len(doms) {
			log.Println("Total number of VM have been change, run again")
			continue
		}
		if UpdateVmInfo(conn, VmQueue) {
			log.Println("VM id have been change, run again")
		}
		SendDataToServer()
		log.Println("End of collect")
		conn.Close()
	}
}
