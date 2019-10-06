package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
)

func main() {
	webui()
}

func NodeStatus(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, Metrice())
}

func webui() {
	server := http.Server{
		Addr: "0.0.0.0:51866",
	}
	http.HandleFunc("/", NodeStatus)
	server.ListenAndServe()
}

func Metrice() string {
	node_status := make(map[string]interface{})
	// cpuload
	cmd_output, _ := exec.Command("/bin/bash", "-c", "uptime |awk -F',| +' '{print $(NF-4),$(NF-2),$NF}'").Output()
	cpu_load_status := make(map[string]interface{})
	v := strings.Split(string(cmd_output), " ")
	cpu_load_status["m1"] = strings.TrimRight(v[0], "\n")

	cpu_load_status["m5"] = strings.TrimRight(v[1], "\n")
	cpu_load_status["m15"] = strings.TrimRight(v[2], "\n")

	// cpuusage
	cmd_output, _ = exec.Command("/bin/bash", "-c", "vmstat 1 3|awk 'NR==5{print $(NF-4)}'").Output()
	res_f := strings.TrimRight(string(cmd_output), "\n")
	cpu_usage_status := make(map[string]interface{})
	res, _ := strconv.ParseFloat(res_f, 64)
	cpu_usage_status["cpu_usage"] = (100 - res) / 100

	// memory
	cmd_output, _ = exec.Command("/bin/bash", "-c", "free -m|awk '/Mem:/{print $2,$NF}'").Output()
	mem_status := make(map[string]interface{})
	v = strings.Split(string(cmd_output), " ")
	mem_status["mem_total"] = strings.TrimRight(v[0], "\n")
	mem_status["mem_available"] = strings.TrimRight(v[1], "\n")

	// disk
	check_decive, _ := exec.Command("/bin/bash", "-c", "lsblk |grep -E 'G|T'|grep -E '─[a-zA-Z]'|awk '!/SWAP/{print $7}'").Output()
	disk_status := make(map[string]interface{})
	mount_status := make(map[string]interface{})
	for _, v := range strings.Split(string(check_decive), "\n") {
		if v != "" {
			check_disk, _ := exec.Command("/bin/bash", "-c", "df |awk '$2 > 50000000 && $6 ~ /\\"+strings.TrimSuffix(v, " ")+"$/{print }'").Output()
			c := strings.Split(string(check_disk), " ")
			mount_status["Size"] = strings.TrimRight(c[1], "\n")
			mount_status["Used"] = strings.TrimRight(c[3], "\n")
			mount_status["Avail"] = strings.TrimRight(c[5], "\n")
			mount_status["UsePercent"] = strings.TrimRight(c[7], "\n")
			disk_status[v] = mount_status
			mount_status = make(map[string]interface{})
		}
	}
	node_status["cpu_load_status"] = cpu_load_status
	node_status["cpu_usage_status"] = cpu_usage_status
	node_status["mem_status"] = mem_status
	node_status["disk_status"] = disk_status

	//fmt.Println((node_status))
	nodejson, _ := json.Marshal(node_status)
	return string(nodejson)
}
