package cmd

import (
	"fmt"
	"github.com/BillysBigFileServer/bfsp-go"
	"github.com/BillysBigFileServer/cli/client"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var shareCmd = &cobra.Command{
	Use:   "share",
	Short: "Share a file on BBFS.io",
	Run:   runShare,
}

func init() {
	rootCmd.AddCommand(shareCmd)
}

func runShare(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		fmt.Println("Usage: share <file>")
		return
	}

	masterKey, err := client.MasterKey()
	if err != nil {
		panic(err)
	}

	bfspClient, err := client.NewFileServerClient()
	if err != nil {
		panic(err)
	}

	files, err := bfsp.ListFileMetadata(bfspClient, []string{}, masterKey)
	if err != nil {
		panic(err)
	}

	var fileMeta *bfsp.FileMetadata
	for _, meta := range files {
		if fullName(meta) == args[0] {
			fileMeta = meta
			break
		}
	}

	shareURL, err := client.ShareFile(fileMeta)
	if err != nil {
		panic(err)
	}

	fmt.Println(shareURL)

}
