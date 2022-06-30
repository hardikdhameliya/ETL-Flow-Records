package fetch

import "time"

// Flow record
type Flow struct {
	Timestamp      time.Time `json:"timestamp" bson:"timestamp"`
	Hostname       string    `json:"hostname" bson:"hostname"`
	Id             string    `json:"id" bson:"id"`
	AppId          uint32    `json:"app_id" bson:"app_id"`
	AppName        string    `json:"app_name" bson:"app_name"`
	Username       string    `json:"username" bson:"username"`
	Service        string    `json:"service" bson:"service"`
	IpVersion      uint8     `json:"ip_version" bson:"ip_version"`
	SrcIp          string    `json:"src_ip" bson:"src_ip"`
	SrcPort        uint16    `json:"src_port" bson:"src_port"`
	DstIp          string    `json:"dst_ip" bson:"dst_ip"`
	DstPort        uint16    `json:"dst_port" bson:"dst_port"`
	Protocol       uint8     `json:"protocol" bson:"protocol"`
	StartTimestamp time.Time `json:"start_timestamp" bson:"start_timestamp"`
	EndTimestamp   time.Time `json:"end_timestamp" bson:"end_timestamp"`
	EndReason      uint8     `json:"end_reason" bson:"end_reason"`
	TcpFlags       uint16    `json:"tcp_flags" bson:"tcp_flags"`
	SrcMac         string    `json:"src_mac" bson:"src_mac"`
	DstMac         string    `json:"dst_mac" bson:"dst_mac"`
	Vlan           uint16    `json:"vlan" bson:"vlan"`
	SrcBytes       uint64    `json:"src_bytes" bson:"src_bytes"`
	SrcPackets     uint64    `json:"src_packets" bson:"src_packets"`
	DstBytes       uint64    `json:"dst_bytes" bson:"dst_bytes"`
	DstPackets     uint64    `json:"dst_packets" bson:"dst_packets"`
	SrcVendor      string    `json:"src_vendor,omitempty" bson:"src_vendor,omitempty"`
	DstVendor      string    `json:"dst_vendor,omitempty" bson:"dst_vendor,omitempty"`
}
