package pcap

import (
	"fmt"
	"log"
	"testing"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

func TestPcap(t *testing.T) {
	// 修改为你自己文件路径
	pcapFile := "Eth0.pcap"

	// 打开 pcap 文件
	handle, err := pcap.OpenOffline(pcapFile)
	if err != nil {
		log.Fatal("打开 pcap 文件失败:", err)
	}
	defer handle.Close()

	// 创建 packet source
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())

	// 遍历每个数据包
	for packet := range packetSource.Packets() {
		fmt.Println("========== 新数据包 ==========")

		ethernetLayer := packet.Layer(layers.LayerTypeEthernet)
		if ethernetLayer != nil {
			ethernetPacket, _ := ethernetLayer.(*layers.Ethernet)
			fmt.Println("以太网层:")
			fmt.Printf("来源MAC: %s, 目标MAC: %s\n", ethernetPacket.SrcMAC, ethernetPacket.DstMAC)
		}

		ipLayer := packet.Layer(layers.LayerTypeIPv4)
		if ipLayer != nil {
			ip, _ := ipLayer.(*layers.IPv4)
			fmt.Println("IP层:")
			fmt.Printf("Source IP: %s, Destination IP: %s\n", ip.SrcIP, ip.DstIP)
		}

		tcpLayer := packet.Layer(layers.LayerTypeTCP)
		if tcpLayer != nil {
			tcp, _ := tcpLayer.(*layers.TCP)
			fmt.Println("TCP层:")
			fmt.Printf("Source Port: %d, Destination Port: %d\n", tcp.SrcPort, tcp.DstPort)
		}

		udpLayer := packet.Layer(layers.LayerTypeUDP)
		if udpLayer != nil {
			udp, _ := udpLayer.(*layers.UDP)
			fmt.Println("UDP层:")
			fmt.Printf("Source Port: %d, Destination Port: %d\n", udp.SrcPort, udp.DstPort)
		}
	}
}
