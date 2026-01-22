package main

import (
	"fmt"
	"regexp"
)

func main() {
	a := "Jun 26 21:15:47 222.92.224.73 JS-SZ-CX-BAS-9.MAN.M6000E" +
		" alarm-log :[An alarm 612643 ID 17297350 level 5 occurred at 21:11:26 06-26-2023 sent by " +
		"JS-SZ-CX-BAS-9.MAN.M6000E PFU-0/7/0 %PORT% Output Optical Power " +
		"(dBm) threshold crossed  (rack = NA, shelf = 0, slot = 7, subslot = 1, port = 5, lane = 0, port_type = 132) <241-description> 15min low alarm occur. Low thres is: -10.0000, Current value is: -10.3763. ]"

	a = "Feb 23 10:46:31 4.13.205.50 alarm-log::[An alarm 612615 ID 132 level 1 occurred at 10:47:06 02-23-2022 sent by SX-JZ-LSXiaMSFZ-A-1.MCN.6150 PFU-1/2/0 %ETH-PORT% Missing Laser Module!  (rack = NA, shelf = 1, slot = 2, cpu = 0, port = 3, channel = 0, port_type = 131)]"
	r := regexp.MustCompile(`rack\s*=\s*(\S+)\s*,\s*shelf\s*=\s*(\d+)\s*,\s*slot\s*=\s*(\d+)\s*,.*?\s*port\s*=\s*(\d+)\s*`)
	lst := r.FindStringSubmatch(a)
	fmt.Println(lst)
}
