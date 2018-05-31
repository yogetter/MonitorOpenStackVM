package main

import (
	libvirt "github.com/libvirt/libvirt-go"
	"log"
	"time"
)

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

func UpdateVmInfo(conn *libvirt.Connect, VmQueue chan *instance) {
	for i, dom := range doms {
		go GetVmStats(&VMs[i], &dom, conn, VmQueue)
		VMs[i] = *<-VmQueue
		VMs[i].SetAllValue(tmp[i])
		VMs[i].GetValue()
		VMs[i].dom.Free()
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
		RefreshDomain(conn)
		log.Println("update VM's information:", len(doms))
		if InitVms != len(doms) {
			log.Println("Total VMs change, run again")
			continue
		}
		UpdateVmInfo(conn, VmQueue)
		log.Println("End of collect")
		conn.Close()
	}
}
