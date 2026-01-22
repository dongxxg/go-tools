package alarm

import (
	"fmt"
	"testing"
	"time"
	"unitechs.com/unios-dice/uni-base/lib/stores/redis"
)

func TestSerialNo(t *testing.T) {

	rdis := redis.NewRedis("localhost:6379", "node", "wiken")

	for j := 0; j < 3; j++ {
		go func(j int) {
			no := NewNumberGenerator(100, 1, rdis, "serial-no")
			for i := 0; i < 38; i++ {
				ss, err := no.Next()
				if err != nil {
					fmt.Println(err)
					continue
				}
				fmt.Printf("%d no: %s\n", j, ss)
			}
		}(j)
	}
	select {}
}

func TestSerialNoConcurrent(t *testing.T) {
	rdis := redis.NewRedis("localhost:6379", "node", "wiken")
	no := NewNumberGenerator(10000001200, 1, rdis, "serial-no")
	for i := 0; i < 1010; i++ {
		go func(i int) {
			n := time.Now()
			ss, err := no.next()
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Printf("time: %d  %d  %d \n", time.Now().Sub(n).Milliseconds(), i, ss)
		}(i)

		//fmt.Printf("%d no: %s\n", j, ss)
	}
	select {}
}
