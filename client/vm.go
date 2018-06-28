package main

import (
	"fmt"
	"github.com/yogetter/libvirt-go"
	"os"
	"strconv"
	"strings"
)

type instance struct {
	Id         string
	Name       string
	MemTotal   int64
	MemUnUsed  int64
	MemUsed    int64
	CpuUsage   float64
	CpuTime    uint64
	VcpuNumber int
	NetStats   []libvirt.DomainStatsNet
	BlockStats []libvirt.DomainStatsBlock
	dom        *libvirt.Domain
	Hostname   string
}

func (s *instance) GetName() {
	xml, err := s.dom.GetXMLDesc(1)
	CheckError(err)
	tmp := strings.SplitAfter(xml, "<nova:name>")[1]
	s.Name = strings.Split(tmp, "</nova:name>")[0]
}

func (s *instance) SetBlockStats(Block []libvirt.DomainStatsBlock) {
	s.BlockStats = make([]libvirt.DomainStatsBlock, len(Block))
	s.BlockStats = Block
}

func (s *instance) SetVcpuNumber() {
	Vcpu, err := s.dom.GetVcpus()
	CheckError(err)
	s.VcpuNumber = len(Vcpu)
}

func (s *instance) SetCpuValue(CpuTime uint64) {
	s.CpuTime = CpuTime
}

func (s *instance) SetMemValue() {
	id, err := s.dom.GetUUIDString()
	CheckError(err)
	mem, err := s.dom.MemoryStats(10, 0)
	s.Id = id
	CheckError(err)
	for _, stat := range mem {
		if stat.Tag == 4 {
			s.MemUnUsed = int64(stat.Val * 1024)
		} else if stat.Tag == 6 {
			s.MemTotal = int64(stat.Val * 1024)
		}
		s.MemUsed = s.MemTotal - s.MemUnUsed
	}
}

func (s *instance) SetInterfaceValue(Net []libvirt.DomainStatsNet) {
	if len(Net) == 0 {
		s.NetStats = make([]libvirt.DomainStatsNet, 1)
	} else {
		s.NetStats = make([]libvirt.DomainStatsNet, len(Net))
		s.NetStats = Net
	}
}

func (s instance) GetValue() []string {
	var data []string
	for _, Block := range s.BlockStats {
		tmpData := "Hostname:" + s.Hostname + ";" + "Uuid:" + s.Id + ";" + "Name:" + s.Name + ";" + "MemTotal:" + strconv.FormatInt(s.MemTotal, 10) +
			";" + "MemUsed:" + strconv.FormatInt(s.MemUsed, 10) + ";" + "MemUnUsed:" + strconv.FormatInt(s.MemUnUsed, 10) +
			";" + "CPU:" + strconv.FormatFloat(s.CpuUsage, 'f', -1, 64) + ";" + "Rx:" + strconv.FormatUint(s.NetStats[0].RxBytes, 10) +
			";" + "Tx:" + strconv.FormatUint(s.NetStats[0].TxBytes, 10) + ";" + "BkDev:" + Block.Name +
			";" + "BkWr:" + strconv.FormatUint(Block.WrBytes, 10) + ";" + "BkRd:" + strconv.FormatUint(Block.RdBytes, 10) +
			";" + "BkTotal:" + strconv.FormatUint(Block.Capacity, 10)
		data = append(data, tmpData)
	}
	return data

}

func (s instance) PrintValue() {
	fmt.Println("VM:")
	fmt.Println("Uuid: ", s.Id)
	fmt.Println("Name: ", s.Name)
	fmt.Println("MemTotal: ", strconv.FormatInt(s.MemTotal, 10))
	fmt.Println("MemUsed: ", strconv.FormatInt(s.MemUsed, 10))
	fmt.Println("MemUnUsed: ", strconv.FormatInt(s.MemUnUsed, 10))
	fmt.Println("CPU: ", strconv.FormatFloat(s.CpuUsage, 'f', -1, 64))
	fmt.Println("CPU: ", s.CpuUsage)
	fmt.Println("VcpuNumber: ", s.VcpuNumber)
	fmt.Println("BlockStats: ", s.BlockStats)
	fmt.Println("NetStats: ", s.NetStats)

}
func (s *instance) SetAllValue(tmp instance) {
	usedTime := (s.CpuTime - tmp.CpuTime) / 1000
	s.CpuUsage = float64(usedTime) / float64((60 * 1000000 * s.VcpuNumber))
	s.CpuUsage *= 100
	s.NetStats[0].RxBytes = (s.NetStats[0].RxBytes - tmp.NetStats[0].RxBytes) / 60
	s.NetStats[0].TxBytes = (s.NetStats[0].TxBytes - tmp.NetStats[0].TxBytes) / 60
	for i := 0; i < len(s.BlockStats); i++ {
		s.BlockStats[i].WrBytes = s.BlockStats[i].WrBytes - tmp.BlockStats[i].WrBytes
		s.BlockStats[i].RdBytes = s.BlockStats[i].RdBytes - tmp.BlockStats[i].RdBytes
	}
}

func (s *instance) InitAllValue(dom *libvirt.Domain, domStats []libvirt.DomainStats) {
	DomInfo, err := dom.GetInfo()
	CheckError(err)
	s.Hostname, err = os.Hostname()
	CheckError(err)
	s.dom = dom
	s.SetVcpuNumber()
	s.GetName()
	s.SetMemValue()
	s.SetCpuValue(DomInfo.CpuTime)
	s.SetBlockStats(domStats[0].Block)
	s.SetInterfaceValue(domStats[0].Net)
}
