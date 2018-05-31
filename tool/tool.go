package tool

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

func CheckError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

type OpenStackTool struct {
	OS_AUTH_URL   string
	NOVA_ENDPOINT string
	influx        DB
	InstancesInDB [][]interface{}
	InstancesLive []string
}

func (o *OpenStackTool) Init(influx *DB) {
	data, _ := os.Open("json/openstack_conf.json")
	decoder := json.NewDecoder(data)
	err := decoder.Decode(o)
	CheckError(err)
}

func (o *OpenStackTool) GetUrl(catalog interface{}) {
	for _, value := range catalog.([]interface{}) {
		if value.(map[string]interface{})["name"] == "nova" {
			for _, url := range value.(map[string]interface{})["endpoints"].([]interface{}) {
				if url.(map[string]interface{})["interface"] == "internal" {
					o.NOVA_ENDPOINT = url.(map[string]interface{})["url"].(string)
				}
			}
		}
	}
}

func (o *OpenStackTool) GetToken() string {
	var ResponseData interface{}
	data, _ := os.Open("json/user_info.json")
	client := &http.Client{}
	req, err := http.NewRequest("POST", o.OS_AUTH_URL+":5000/v3/auth/tokens", data)
	CheckError(err)
	req.Header.Set("Content-Type", "application/json")
	res, err := client.Do(req)
	defer res.Body.Close()
	o.IoRead(res, &ResponseData)
	catalog := ResponseData.(map[string]interface{})["token"].(map[string]interface{})["catalog"]
	o.GetUrl(catalog)
	token := res.Header.Get("X-Subject-Token")
	return token
}
func (o *OpenStackTool) IoRead(r *http.Response, f *interface{}) {
	body, err := ioutil.ReadAll(r.Body)
	dec := json.NewDecoder(strings.NewReader(string(body)))
	err = dec.Decode(f)
	CheckError(err)

}
func (o *OpenStackTool) InsertInstance(res_data []interface{}) []string {
	var tmp []string
	for _, value := range res_data {
		tmp = append(tmp, value.(map[string]interface{})["id"].(string))
	}
	return tmp
}
func (o *OpenStackTool) GetInstances() []interface{} {
	var tmp interface{}
	token := o.GetToken()
	client := &http.Client{}
	req, err := http.NewRequest("GET", o.NOVA_ENDPOINT+"/servers?all_tenants", nil)
	CheckError(err)
	req.Header.Set("X-Auth-Token", token)
	res, err := client.Do(req)
	defer res.Body.Close()
	o.IoRead(res, &tmp)
	res_data := tmp.(map[string]interface{})["servers"].([]interface{})
	return res_data
}

func (o *OpenStackTool) DeleteData() {
	// Found ID should be delete
	flag := true
	for _, data1 := range o.InstancesInDB {
		for _, data2 := range o.InstancesLive {
			if data1[1] == data2 {
				flag = false
				log.Println("same:", data1)
				break
			}
		}
		if flag {
			log.Println("Delete id:", data1)
			log.Println("influx Url:", o.influx.Url)
			o.influx.QueryInfo("'"+data1[1].(string)+"'", "drop series where uuid = ")
		}
		flag = true
	}
}
func (o *OpenStackTool) QueryData() {
	tmp_data := o.influx.QueryInfo("uuid", "show tag values from vm_usage with key = ")
	if tmp_data[0].Series != nil {
		//log.Println("Tmp_data:", tmp_data[0], "value:", tmp_data[0].Series)
		o.InstancesInDB = tmp_data[0].Series[0].Values
	}
}

func (o *OpenStackTool) CheckStart() {
	InstanceData := o.GetInstances()
	o.InstancesLive = o.InsertInstance(InstanceData)
	o.QueryData()
	log.Println("InDb:", o.InstancesInDB)
	log.Println("Live:", o.InstancesLive)
	o.DeleteData()
}
