package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/influxdata/influxdb/client/v2"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
)

const (
	ADDR     = "http://vidc-testing-db.chinacloudapp.cn:8086"
	MyDB     = "bymonitor"
	username = "bymonitor"
	password = "byMonitor"
)

func GetAddr() string {
	resp, _ := http.Get("http://myexternalip.com/raw")
	body, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	return strings.TrimSpace(string(body))
}

func influxSendData(
	name string,
	tags map[string]string,
	fields map[string]interface{},
	t ...time.Time,
) {
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     ADDR,
		Username: username,
		Password: password,
	})
	if err != nil {
		fmt.Println("Error creating InfluxDB Client: ", err.Error())
		return
	}
	bp, _ := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  MyDB,
		Precision: "s",
	})
	pt, err := client.NewPoint(name, tags, fields, time.Now())
	if err != nil {
		fmt.Println("Error: ", err.Error())
		return
	}
	bp.AddPoint(pt)
	c.Write(bp)
	c.Close()
	runtime.GC()
}

func sendHostInfo(thisIp *string) string {
	c_info, _ := cpu.Info()
	h_info, _ := host.Info()
	m_virtualmemory, _ := mem.VirtualMemory()
	d_usage, _ := disk.Usage("/")
	influxSendData("host",
		map[string]string{"IP": *thisIp},
		map[string]interface{}{
			"hostname":             h_info.Hostname,
			"uptime":               int(h_info.Uptime),
			"bootTime":             int(h_info.BootTime),
			"procs":                int(h_info.Procs),
			"os":                   h_info.OS,
			"platform":             h_info.Platform,
			"platformFamily":       h_info.PlatformFamily,
			"platformVersion":      h_info.PlatformVersion,
			"kernelVersion":        h_info.KernelVersion,
			"virtualizationSystem": h_info.VirtualizationSystem,
			"virtualizationRole":   h_info.VirtualizationRole,
			"hostid":               h_info.HostID,
			"cpuNum":               len(c_info),
			"cpuModelName":         c_info[0].ModelName,
			"cpuCores":             c_info[0].Cores,
			"sysDiskSize":          int(d_usage.Total),
			"memSize":              int(m_virtualmemory.Total),
		}, time.Now())
	return "Send host info complete "
}

func sendCpuInfo(thisIp *string) string {
	c_percent, _ := cpu.Percent(0, true)
	c_times, _ := cpu.Times(false)
	for ct_k, _ := range c_times {
		ct_ks := "0"
		if ct_k != 0 {
			ct_ks = string(ct_k)
		}
		influxSendData("cpu",
			map[string]string{"IP": *thisIp, "NUM": ct_ks},
			map[string]interface{}{
				"useRatio":  c_percent[ct_k],
				"total":     c_times[ct_k].Total(),
				"idle":      c_times[ct_k].Idle,
				"system":    c_times[ct_k].System,
				"user":      c_times[ct_k].User,
				"nice":      c_times[ct_k].Nice,
				"iowait":    c_times[ct_k].Iowait,
				"irq":       c_times[ct_k].Irq,
				"softirq":   c_times[ct_k].Softirq,
				"steal":     c_times[ct_k].Steal,
				"guest":     c_times[ct_k].Guest,
				"guestNice": c_times[ct_k].GuestNice,
				"stolen":    c_times[ct_k].Stolen,
			}, time.Now())
	}
	return "Send cpu info complete "
}

func sendMemInfo(thisIp *string) string {
	m_virtualmemory, _ := mem.VirtualMemory()
	influxSendData("mem",
		map[string]string{"IP": *thisIp},
		map[string]interface{}{
			"available":    int(m_virtualmemory.Available),
			"total":        int(m_virtualmemory.Total),
			"used":         int(m_virtualmemory.Used),
			"usedPercent":  m_virtualmemory.UsedPercent,
			"free":         int(m_virtualmemory.Free),
			"active":       int(m_virtualmemory.Active),
			"inactive":     int(m_virtualmemory.Inactive),
			"wired":        int(m_virtualmemory.Wired),
			"buffers":      int(m_virtualmemory.Buffers),
			"cached":       int(m_virtualmemory.Cached),
			"writeback":    int(m_virtualmemory.Writeback),
			"dirty":        int(m_virtualmemory.Dirty),
			"writebackTmp": int(m_virtualmemory.WritebackTmp),
			"shared":       int(m_virtualmemory.Shared),
			"slab":         int(m_virtualmemory.Slab),
			"pageTables":   int(m_virtualmemory.PageTables),
			"swapCached":   int(m_virtualmemory.SwapCached),
		}, time.Now())
	runtime.GC()
	return "Send mem info complete "
}

func sendDiskInfo(thisIp *string) string {
	d_partitions, _ := disk.Partitions(false)
	for _, di_v := range d_partitions {
		d_usage, _ := disk.Usage(di_v.Mountpoint)
		influxSendData("disk",
			map[string]string{"IP": *thisIp, "DEVICE": di_v.Device},
			map[string]interface{}{
				"mountpoint":        di_v.Mountpoint,
				"fstype":            di_v.Fstype,
				"opts":              di_v.Opts,
				"total":             int(d_usage.Total),
				"free":              int(d_usage.Free),
				"used":              int(d_usage.Used),
				"usedPercent":       d_usage.UsedPercent,
				"inodesTotal":       int(d_usage.InodesTotal),
				"inodesUsed":        int(d_usage.InodesUsed),
				"inodesFree":        int(d_usage.InodesFree),
				"inodesUsedPercent": d_usage.InodesUsedPercent,
			}, time.Now())
	}
	return "Send disk info complete "
}

func sendNetInfo(thisIp *string) string {
	n_iocounters, _ := net.IOCounters(true)
	for _, ni_v := range n_iocounters {
		influxSendData("net",
			map[string]string{"IP": *thisIp, "NAME": ni_v.Name},
			map[string]interface{}{
				"bytesSent":   int(ni_v.BytesSent),
				"bytesRecv":   int(ni_v.BytesRecv),
				"packetsSent": int(ni_v.PacketsSent),
				"packetsRecv": int(ni_v.PacketsRecv),
				"errin":       int(ni_v.Errin),
				"errout":      int(ni_v.Errout),
				"dropin":      int(ni_v.Dropin),
				"dropout":     int(ni_v.Dropout),
				"fifoin":      int(ni_v.Fifoin),
				"fifoout":     int(ni_v.Fifoout),
			}, time.Now())
	}
	return "Send net info complete "
}

/**
 * 判断文件是否存在  存在返回 true 不存在返回false
 */
func checkFileIsExist(filename string) bool {
	var exist = true
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		exist = false
	}
	return exist
}

func main() {
	fmt.Println(time.Now())
	thisIp := GetAddr()
	//	sendHostInfo(&thisIp)
	d := time.Now().Day()
	h := time.Now().Hour()
	filepath := "log/" + fmt.Sprintf("%d", time.Now().Year()) + "/" + fmt.Sprintf("%d", time.Now().Month()) + "/"
	os.MkdirAll(filepath, 0777)
	fileName := filepath + fmt.Sprintf("%d", d) + ".mtor"
	logFile, err := os.OpenFile(fileName, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalln("Create file error !")
	}
	defer logFile.Close()
	// 创建一个日志对象
	debugLog := log.New(logFile, "[BM-Timer-Log]", log.LstdFlags)
	debugLog.Println("BY-Monitor Start ")

	ticker := time.NewTicker(time.Second * 5)
	for _ = range ticker.C {
		if d != time.Now().Day() {
			d = time.Now().Day()
			logFile.Close()
			//创建每日新日志文件
			filepath = "log/" + fmt.Sprintf("%d", time.Now().Year()) + "/" + fmt.Sprintf("%d", time.Now().Month()) + "/"
			os.MkdirAll(filepath, 0777)
			fileName = filepath + fmt.Sprintf("%d", d) + ".mtor"
			logFile, err := os.OpenFile(fileName, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
			if err != nil {
				log.Fatalln("Create file error !")
			}
			debugLog = log.New(logFile, "[BM-Timer-Log]", log.LstdFlags)
		}
		if h != time.Now().Hour() {
			debugLog.Println(sendHostInfo(&thisIp))
			h = time.Now().Hour()
		}
		debugLog.Println(sendCpuInfo(&thisIp))
		debugLog.Println(sendMemInfo(&thisIp))
		debugLog.Println(sendDiskInfo(&thisIp))
		debugLog.Println(sendNetInfo(&thisIp))
	}

}
