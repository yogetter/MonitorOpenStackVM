package main

import (
	"encoding/json"
	"github.com/influxdata/influxdb/client/v2"
	"log"
	"os"
	"time"
)

type DB struct {
	Url      string
	Db       string
	Username string
	Password string
}

var Hostname string

func (d *DB) Init() {
	//read config
	file, _ := os.Open("json/db_conf.json")
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

func (d *DB) InsertVmInfo(VM instance) {
	// Create a new HTTPClient
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     d.Url,
		Username: d.Username,
		Password: d.Password,
	})
	CheckError(err)
	// Create a new point batch
	for _, Block := range VM.BlockStats {
		bp, err := client.NewBatchPoints(client.BatchPointsConfig{
			Database:  d.Db,
			Precision: "s",
		})
		CheckError(err)
		// Create a point and add to batch
		tags := map[string]string{"uuid": VM.Id, "Name": VM.Name, "Hostname": Hostname, "BkDev": Block.Name}
		fields := map[string]interface{}{
			"Total":    VM.MemTotal,
			"Used":     VM.MemUsed,
			"UnUsed":   VM.MemUnUsed,
			"CpuUsage": VM.CpuUsage,
			"Rx":       int64(VM.NetStats[0].RxBytes),
			"Tx":       int64(VM.NetStats[0].TxBytes),
			"BkTotal":  int64(Block.Capacity),
			"BkWr":     int64(Block.WrBytes),
			"BkRd":     int64(Block.RdBytes),
		}
		log.Println("Send VM information:", tags, fields)
		pt, err := client.NewPoint("vm_usage", tags, fields, time.Now())
		CheckError(err)
		bp.AddPoint(pt)

		// Write the batch
		if err := c.Write(bp); err != nil {
			log.Fatal(err)
		}
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
	//log.Printf("%T", res[0].Series[0].Values)
	log.Println("Success")
	c.Close()
	return res //[0].Series[0].Values
}
