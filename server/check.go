package main

import (
	"MonitorOpenStackVM/tool"
	"log"
)

type CheckInstance struct {
	influx        DB
	InstancesInDB [][]interface{}
	InstancesLive []string
}

func (c *CheckInstance) InsertInstance(res_data []interface{}) []string {
	var tmp []string
	for _, value := range res_data {
		tmp = append(tmp, value.(map[string]interface{})["id"].(string))
	}
	return tmp
}

func (c *CheckInstance) DeleteData() {
	// Found ID should be delete
	flag := true
	for _, data1 := range c.InstancesInDB {
		for _, data2 := range c.InstancesLive {
			if data1[1] == data2 {
				flag = false
				log.Println("same:", data1)
				break
			}
		}
		if flag {
			log.Println("Delete id:", data1)
			log.Println("influx Url:", c.influx.Url)
			c.influx.QueryInfo("'"+data1[1].(string)+"'", "drop series where uuid = ")
		}
		flag = true
	}
}
func (c *CheckInstance) QueryData() {
	tmp_data := c.influx.QueryInfo("uuid", "show tag values from vm_usage with key = ")
	if tmp_data[0].Series != nil {
		//log.Println("Tmp_data:", tmp_data[0], "value:", tmp_data[0].Series)
		c.InstancesInDB = tmp_data[0].Series[0].Values
	}
}

func (c *CheckInstance) CheckStart(o *tool.OpenStackTool, influx *DB) {
	c.influx = *influx
	InstanceData := o.GetInstances()
	c.InstancesLive = c.InsertInstance(InstanceData)
	c.QueryData()
	log.Println("InDb:", c.InstancesInDB)
	log.Println("Live:", c.InstancesLive)
	c.DeleteData()
}
