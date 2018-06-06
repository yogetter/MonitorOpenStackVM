# MonitorOpenStackVM #
This is a project for Monitor OpenStack VM with Grafana & influxdb.  
Support OpenStack version: pike.  
# Requirement #
* Have installed Grafana and set influxdb for db store.  
* Import **json/grafana_dashboard.json** in your Grafana dashboard.  
* Edit **json/db_conf.json**, **json/openstack_conf.json**, **json/rabbitmq.json**, **json/user_info.json** for your environment.  
# Server #
Run on 1 controller node
# Client #
Run on each compute node
