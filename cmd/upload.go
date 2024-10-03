package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/BillysBigFileServer/bbfs-cli/client"
	"github.com/BillysBigFileServer/bfsp-go"
	"github.com/spf13/cobra"
)

var uploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,

	Run: runUpload,
}

func init() {
	rootCmd.AddCommand(uploadCmd)
}

func runUpload(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		fmt.Println("Usage: upload <file>")
		return
	}

	masterKey, err := client.MasterKey()
	if err != nil {
		panic(err)
	}

	file, err := os.Open(args[0])
	if err != nil {
		panic(err)
	}

	fmt.Println("uploading file", args[0])

	bfspClient, err := client.NewFileServerClient()
	if err != nil {
		panic(err)
	}
	ctx := context.Background()
	ctx = bfsp.ContextWithClient(ctx, bfspClient)
	ctx = bfsp.ContextWithMasterKey(ctx, masterKey)

	err = bfsp.UploadFile(ctx, &bfsp.FileInfo{
		Name:   file.Name(),
		Reader: file,
	}, 100)
	if err != nil {
		panic(err)
	}

}
