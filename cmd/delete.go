/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/BillysBigFileServer/bfsp-go"
	"github.com/BillysBigFileServer/cli/client"
	"github.com/spf13/cobra"
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a file on BBFS.io",
	Run:   runDelete,
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}

func runDelete(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		fmt.Println("Usage: delete <file>")
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

	if fileMeta == nil {
		fmt.Println("file not found")
		return
	}

	err = bfsp.DeleteFileMetadata(bfspClient, fileMeta.Id)
	if err != nil {
		panic(err)
	}

	chunkIDs := []string{}
	for _, chunkID := range fileMeta.Chunks {
		chunkIDs = append(chunkIDs, chunkID)
	}

	err = bfsp.DeleteChunks(bfspClient, chunkIDs)
	if err != nil {
		panic(err)
	}

}
