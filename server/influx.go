package main

import (
	"encoding/json"
	"github.com/influxdata/influxdb/client/v2"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

type instance struct {
	Uuid      string
	Name      string
	MemTotal  string
	MemUnUsed string
	MemUsed   string
	CpuUsage  string
	Rx        string
	Tx        string
	BkTotal   string
	BkWr      string
	BkRd      string
	BkDev     string
}

type DB struct {
	Url      string
	Db       string
	Username string
	Password string
}

var Hostname string

func toInt(input string) int64 {
	r, err := strconv.ParseInt(input, 10, 64)
	CheckError(err)
	return r
}
func toFloat(input string) float64 {
	r, err := strconv.ParseFloat(input, 64)
	CheckError(err)
	return r
}

func (d *DB) Init() {
	//read config
	file, _ := os.Open("../json/db_conf.json")
	decoder := json.NewDecoder(file)
	err := decoder.Decode(d)
	CheckError(err)
	Hostname, err = os.Hostname()
	CheckError(err)
	log.Println("DB URL:", d.Url)
	log.Println("DB Name:", d.Db)
	log.Println("DB Username:", d.Username)
	log.Println("DB Password:", d.Password)
	file.Close()
}
func (d *DB) InitVmInfo(data []string) instance {
	VM := instance{}
	for _, d := range data {
		tmpData := strings.Split(d, ":")
		if tmpData[0] == "Uuid" {
			VM.Uuid = tmpData[1]
		} else if tmpData[0] == "Name" {
			VM.Name = tmpData[1]
		} else if tmpData[0] == "MemTotal" {
			VM.MemTotal = tmpData[1]
		} else if tmpData[0] == "MemUsed" {
			VM.MemUsed = tmpData[1]
		} else if tmpData[0] == "MemUnUsed" {
			VM.MemUnUsed = tmpData[1]
		} else if tmpData[0] == "CPU" {
			VM.CpuUsage = tmpData[1]
		} else if tmpData[0] == "Rx" {
			VM.Rx = tmpData[1]
		} else if tmpData[0] == "Tx" {
			VM.Tx = tmpData[1]
		} else if tmpData[0] == "BkDev" {
			VM.BkDev = tmpData[1]
		} else if tmpData[0] == "BkWr" {
			VM.BkWr = tmpData[1]
		} else if tmpData[0] == "BkRd" {
			VM.BkRd = tmpData[1]
		} else if tmpData[0] == "BkTotal" {
			VM.BkTotal = tmpData[1]
		}
	}
	return VM
}
func (d *DB) InsertVmInfo(data []string) {
	VM := d.InitVmInfo(data)
	// Create a new HTTPClient
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     d.Url,
		Username: d.Username,
		Password: d.Password,
	})
	CheckError(err)
	// Create a new point batch
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  d.Db,
		Precision: "s",
	})
	CheckError(err)
	// Create a point and add to batch
	tags := map[string]string{"uuid": VM.Uuid, "Name": VM.Name, "Hostname": Hostname, "BkDev": VM.BkDev}
	fields := map[string]interface{}{
		"Total":    toInt(VM.MemTotal),
		"Used":     toInt(VM.MemUsed),
		"UnUsed":   toInt(VM.MemUnUsed),
		"CpuUsage": toFloat(VM.CpuUsage),
		"Rx":       toInt(VM.Rx),
		"Tx":       toInt(VM.Tx),
		"BkTotal":  toInt(VM.BkTotal),
		"BkWr":     toInt(VM.BkWr),
		"BkRd":     toInt(VM.BkRd),
	}
	log.Println("Send VM information:", tags, fields)
	pt, err := client.NewPoint("vm_usage", tags, fields, time.Now())
	CheckError(err)
	bp.AddPoint(pt)

	// Write the batch
	if err := c.Write(bp); err != nil {
		log.Fatal(err)
	}
	c.Close()
}

func (d *DB) QueryInfo(id string, command string) []client.Result {
	var res []client.Result
	// Create a new HTTPClient
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     d.Url,
		Username: d.Username,
		Password: d.Password,
	})
	CheckError(err)
	q := client.Query{
		Command:  command + id,
		Database: d.Db,
	}
	if response, err := c.Query(q); err == nil {
		if response.Error() != nil {
			log.Println("err1:", response.Error())
		}
		res = response.Results
	} else {
		log.Println("err2", err)
	}
	log.Println("Success")
	c.Close()
	return res
}
