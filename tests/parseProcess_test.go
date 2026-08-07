/**
  @Author: dongxx
  @Since: 2026/3/19 10:18
  @Desc: //TODO
**/

package tests

import (
	"bufio"
	"bytes"
	"regexp"
	"strconv"
	"testing"
)

var REG_PID = regexp.MustCompile(`^(\d+)\s+(.*)`)

func TestParse(t *testing.T) {
	output := []byte("    PID COMMAND\n      1 /usr/lib/systemd/systemd --switched-root --system --deserialize 16\n      2 [kthreadd]\n      3 [rcu_gp]\n      4 [rcu_par_gp]\n      6 [kworker/0:0H-kblockd]\n      8 [mm_percpu_wq]\n      9 [ksoftirqd/0]\n     10 [rcu_sched]\n     11 [rcu_bh]\n     12 [migration/0]\n     13 [cpuhp/0]\n     14 [cpuhp/1]\n     15 [migration/1]\n     16 [ksoftirqd/1]\n     18 [kworker/1:0H-kblockd]\n     19 [cpuhp/2]\n     20 [migration/2]\n     21 [ksoftirqd/2]\n     23 [kworker/2:0H-kblockd]\n     24 [cpuhp/3]\n     25 [migration/3]\n     26 [ksoftirqd/3]\n     28 [kworker/3:0H-kblockd]\n     30 [kdevtmpfs]\n     31 [netns]\n     32 [kauditd]\n     35 [khungtaskd]\n     36 [oom_reaper]\n     37 [writeback]\n     38 [kcompactd0]\n     39 [ksmd]\n     40 [khugepaged]\n     41 [crypto]\n     42 [kintegrityd]\n     43 [kblockd]\n     44 [md]\n     45 [edac-poller]\n     46 [watchdogd]\n     59 [kswapd0]\n    111 [kthrotld]\n    112 [acpi_thermal_pm]\n    113 [kmpath_rdacd]\n    114 [kaluad]\n    116 [ipv6_addrconf]\n    117 [kstrp]\n    385 [ata_sff]\n    392 [kworker/3:1H-kblockd]\n    393 [scsi_eh_0]\n    394 [scsi_tmf_0]\n    395 [scsi_eh_1]\n    396 [scsi_tmf_1]\n    409 [kworker/2:1H-events_highpri]\n    425 [kworker/1:1H-kblockd]\n    464 [kdmflush]\n    471 [kdmflush]\n    489 [jbd2/dm-0-8]\n    490 [ext4-rsv-conver]\n    502 [kworker/0:1H-kblockd]\n    597 /usr/lib/systemd/systemd-journald\n    616 /usr/lib/systemd/systemd-udevd\n    735 [jbd2/vda1-8]\n    737 [ext4-rsv-conver]\n    745 [ttm_swap]\n    754 /usr/lib/systemd/systemd-timesyncd\n    759 /sbin/auditd\n    782 /usr/bin/dbus-daemon --system --address=systemd: --nofork --nopidfile --systemd-activation --syslog-only\n    787 /usr/lib/polkit-1/polkitd --no-debug\n    790 /usr/bin/qemu-ga --method=virtio-serial --path=/dev/virtio-ports/org.qemu.guest_agent.0 --blacklist=guest-file-open,guest-file-close,guest-file-read,guest-file-write,guest-file-seek,guest-file-flush,guest-exec,guest-exec-status -F/etc/qemu-ga/fsfreeze-hook\n    792 /usr/sbin/chronyd\n    794 /usr/sbin/irqbalance --pid=/var/run/irqbalance.pid\n    795 /sbin/rngd -f\n    818 /usr/lib/systemd/systemd-logind\n    822 /usr/sbin/restorecond\n    844 /usr/bin/python3 /usr/sbin/firewalld --nofork --nopid\n    912 /usr/sbin/NetworkManager --no-daemon\n    913 /usr/lib/systemd/systemd-networkd\n   1388 /usr/sbin/rsyslogd -n -iNONE\n   1389 sshd: /usr/sbin/sshd -D [listener] 0 of 10-100 startups\n   1392 /usr/bin/python3 -Es /usr/sbin/tuned -l -P\n   1399 /usr/sbin/crond -n\n   1403 /sbin/agetty -o -p -- \\u --noclear tty1 linux\n   1965 /home/ctff_agent_yunos_64/jre/bin/java -Djava.library.path=/home/ctff_agent_yunos_64/bin -Xms128m -Xmx256m -XX:PermSize=128m -XX:MaxPermSize=256m -jar /home/ctff_agent_yunos_64/bin/CTFFMonitor.jar\n1080005 /home/ctff_agent_yunos_64/jre/bin/java -Djava.library.path=/home/ctff_agent_yunos_64/bin -Xms128m -Xmx256m -XX:PermSize=128m -XX:MaxPermSize=256m -jar /home/ctff_agent_yunos_64/bin/CTFFAgent.jar\n1634396 /usr/lib/systemd/systemd --user\n1634399 (sd-pam)\n1842971 [kworker/0:1-mm_percpu_wq]\n1843087 [kworker/2:0-events]\n1847180 [kworker/u8:1-events_unbound]\n1847396 [kworker/2:2-mm_percpu_wq]\n1858153 [kworker/3:0-events_power_efficient]\n1864911 pppd call dsl-provider vxlan0.c.112 user 057192635126@bc password 3050.9 mru 1300 debug maxfail 0 logfile /var/log/ppp-057192635126@bc-ns_vxlan300801.log\n1864945 pppd call dsl-provider vxlan0.c.112 user 057192635126@bc password 3050.9 mru 1300 debug maxfail 0 logfile /var/log/ppp-057192635126@bc-ns_vxlan300101.log\n1864992 [kworker/1:0-mm_percpu_wq]\n1864999 [kworker/0:0]\n1865030 sshd: root [priv]\n1865034 sshd: root@pts/0\n1865035 -bash\n1865082 sshd: root [priv]\n1865086 sshd: root@notty\n1865087 /usr/libexec/openssh/sftp-server -l INFO -f AUTH\n1865131 [kworker/u8:0-events_unbound]\n1865136 [kworker/3:2-cgroup_destroy]\n1865138 [kworker/1:1-ata_sff]\n1865183 [kworker/1:2-ata_sff]\n1865213 ps -ewwo pid,args\n")
	// 读取输出并解析
	processName := "pppd call dsl-provider vxlan0.c.112 user 057192635126@bc password 3050.9 mru 1300 debug maxfail 0 logfile /var/log/ppp-057192635126@bc-ns_vxlan300101.log"
	var pids []int
	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		var command, pidStr string
		if matchs := REG_PID.FindStringSubmatch(line); len(matchs) > 0 {
			pidStr = matchs[1]
			command = matchs[2]
			// 检查命令是否包含指定名称
			if command == processName {
				pid, err := strconv.Atoi(pidStr)
				if err != nil {
					continue
				}
				pids = append(pids, pid)
			}
		}
	}

	t.Logf("pids %v", pids)
	if err := scanner.Err(); err != nil {
		t.Logf("err %s", err.Error())
	}
}