package cmd

import (
	"context"
	"fmt"
	"strconv"

	"github.com/BillysBigFileServer/bbfs-cli/client"
	"github.com/BillysBigFileServer/bfsp-go"
	"github.com/BillysBigFileServer/bfsp-go/usage"
	"github.com/spf13/cobra"
)

var usageCmd = &cobra.Command{
	Use:   "usage",
	Short: "Shows how much data stored in your account",
	Run:   runUsage,
}

func init() {
	rootCmd.AddCommand(usageCmd)
}

func runUsage(cmd *cobra.Command, args []string) {
	bfspClient, err := client.NewFileServerClient()
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	ctx = bfsp.ContextWithClient(ctx, bfspClient)
	usage, err := usage.GetUsage(ctx)
	if err != nil {
		panic(err)
	}
	fmt.Println(humanSize(usage.TotalUsage), "/", humanSize(usage.StorageCap))

}

func humanSize(size uint64) string {
	fileSizeFloat := float64(size)
	switch {
	case size < 1024.0:
		return strconv.FormatFloat(fileSizeFloat, 'f', 2, 64) + "B"
	case size < 1024.0*1024.0:
		return strconv.FormatFloat(fileSizeFloat/1024.0, 'f', 2, 64) + "KiB"
	case size < 1024.0*1024.0*1024.0:
		return strconv.FormatFloat(fileSizeFloat/(1024.0*1024.0), 'f', 2, 64) + "MiB"
	default:
		return strconv.FormatFloat(fileSizeFloat/(1024.0*1024.0*1024.0), 'f', 2, 64) + "GiB"
	}

}
