package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"time"
)

func main() {
    cmd := &cobra.Command{
		Use:   "abc",
		Short: "abx",
		Long: "abc",
		Run: func(cmd *cobra.Command, args []string) {
            b, _ := cmd.Flags().GetBool("b")
            c, _ := cmd.Flags().GetBool("c")
            if b {
				fmt.Println("================b\n****************")
				time.Sleep(10 * time.Second)
				fmt.Println("================end\n****************")
			}
			if c {
				fmt.Println("================c\n****************")
				time.Sleep(10 * time.Second)
				fmt.Println("================end\n****************")
			}
		},
	}

	cmd.Flags().BoolP("b", "b", false, "b")
	cmd.Flags().BoolP("c", "c", false, "c")
    cmd.Execute()
}
