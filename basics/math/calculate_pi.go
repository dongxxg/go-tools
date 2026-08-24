package main

import (
	"fmt"
	"github.com/spf13/cobra"

	//"github.com/spf13/cobra"
	"math"
	"math/big"
	"strconv"
	"time"
)

var precision uint = 0

func Pi(accuracy uint) *big.Float {
	k := 0
	pi := new(big.Float).SetPrec(precision).SetFloat64(0)
	k1k2k3 := new(big.Float).SetPrec(precision).SetFloat64(0)
	k4k5k6 := new(big.Float).SetPrec(precision).SetFloat64(0)
	temp := new(big.Float).SetPrec(precision).SetFloat64(0)
	minusOne := new(big.Float).SetPrec(precision).SetFloat64(-1)
	total := new(big.Float).SetPrec(precision).SetFloat64(0)

	two2Six := math.Pow(2, 6)
	two2SixBig := new(big.Float).SetPrec(precision).SetFloat64(two2Six)

	for {
		if k > int(accuracy) {
			break
		}
		t1 := float64(float64(1) / float64(10*k+9))
		k1 := new(big.Float).SetPrec(precision).SetFloat64(t1)
		t2 := float64(float64(64) / float64(10*k+3))
		k2 := new(big.Float).SetPrec(precision).SetFloat64(t2)
		t3 := float64(float64(32) / float64(4*k+1))
		k3 := new(big.Float).SetPrec(precision).SetFloat64(t3)
		k1k2k3.Sub(k1, k2)
		k1k2k3.Sub(k1k2k3, k3)

		t4 := float64(float64(4) / float64(10*k+5))
		k4 := new(big.Float).SetPrec(precision).SetFloat64(t4)
		t5 := float64(float64(4) / float64(10*k+7))
		k5 := new(big.Float).SetPrec(precision).SetFloat64(t5)
		t6 := float64(float64(1) / float64(4*k+3))
		k6 := new(big.Float).SetPrec(precision).SetFloat64(t6)
		k4k5k6.Add(k4, k5)
		k4k5k6.Add(k4k5k6, k6)
		k4k5k6 = k4k5k6.Mul(k4k5k6, minusOne)
		temp.Add(k1k2k3, k4k5k6)

		k7temp := new(big.Int).Exp(big.NewInt(-1), big.NewInt(int64(k)), nil)
		k8temp := new(big.Int).Exp(big.NewInt(1024), big.NewInt(int64(k)), nil)

		k7 := new(big.Float).SetPrec(precision).SetFloat64(0)
		k7.SetInt(k7temp)
		k8 := new(big.Float).SetPrec(precision).SetFloat64(0)
		k8.SetInt(k8temp)

		t9 := float64(256) / float64(10*k+1)
		k9 := new(big.Float).SetPrec(precision).SetFloat64(t9)
		k9.Add(k9, temp)
		total.Mul(k9, k7)
		total.Quo(total, k8)
		pi.Add(pi, total)

		k = k + 1
	}
	pi.Quo(pi, two2SixBig)
	return pi
}

func main() {
	var rootCmd = &cobra.Command{
		Use:   "myapp",
		Short: "A simple command-line application",
		Long:  "A simple command-line application created using Cobra",
		Run: func(cmd *cobra.Command, args []string) {
			precisionParam, _ := cmd.Flags().GetString("precision")
			newPrecision, _ := strconv.ParseUint(precisionParam, 10, 32)

			startTime := time.Now()

			// 计算 π 的值
			precision = uint(newPrecision)
			PI := Pi(precision)
			piString := PI.Text('f', int(precision))
			if len(piString) <= 100 {
				fmt.Println("π为：", piString)
			} else {

				fmt.Println("π前100位为：", piString[:100])
			}
			elapsedTime := time.Since(startTime)
			fmt.Printf("计算耗时：%s\n", elapsedTime)
		},
	}
	rootCmd.Flags().StringP("precision", "p", "", "Precision of the π")
	rootCmd.MarkFlagRequired("precision")
	rootCmd.Execute()

}
