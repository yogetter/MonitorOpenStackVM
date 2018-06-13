# MonitorOpenStackVM #
This is a project for Monitor OpenStack VM with Grafana & influxdb.  
Support OpenStack version: pike.  
# Requirement #
* Have installed Grafana and set influxdb for db store.  
* Import **json/grafana_dashboard.json** in your Grafana dashboard.  
* Edit **json/db_conf.json**, **json/openstack_conf.json**, **json/rabbitmq.json**, **json/user_info.json** for your environment.  
* Install libvirt-dev package on compute node.
# Server #
Run on 1 controller node
1. git clone  https://github.com/yogetter/MonitorOpenStackVM/
2. cd server  
3. go get && go build  
# Client #
Run on each compute node
1. git clone  https://github.com/yogetter/MonitorOpenStackVM/
2. cd client  
2. go get && go build  
