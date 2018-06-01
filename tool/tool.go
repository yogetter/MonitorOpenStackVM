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
}

func (o *OpenStackTool) Init() {
	data, _ := os.Open("../json/openstack_conf.json")
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
	data, _ := os.Open("../json/user_info.json")
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
	ServerData := tmp.(map[string]interface{})["servers"].([]interface{})
	return ServerData
}
