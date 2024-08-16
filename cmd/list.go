/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"slices"
	"strings"

	"github.com/BillysBigFileServer/bfsp-go"
	"github.com/BillysBigFileServer/cli/client"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List the files you've uploaded",
	Run:   runList,
}

func init() {
	listCmd.Flags().StringP("directory", "d", "/", "")
	listCmd.Flags().BoolP("show-directories", "", true, "")
	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) {
	masterKey, err := client.MasterKey()
	if err != nil {
		panic(err)
	}

	bfspClient, err := client.NewFileServerClient()
	if err != nil {
		panic(err)
	}

	dir, err := cmd.Flags().GetString("directory")
	if err != nil {
		panic(err)
	}
	showDirs, err := cmd.Flags().GetBool("show-directories")
	if err != nil {
		panic(err)
	}

	dirSlice := directoryToSlice(dir)

	files, err := bfsp.ListFileMetadata(bfspClient, []string{}, masterKey)
	if err != nil {
		panic(err)
	}

	filesToShow := []*bfsp.FileMetadata{}
	uniqueDirectories := map[string]bool{}

	for _, meta := range files {
		if compareDirSlice(dirSlice, meta.Directory) {
			filesToShow = append(filesToShow, meta)
		}

		if strings.Contains(sliceToDirectory(meta.Directory), dir) {
			uniqueDirectories[sliceToDirectory(meta.Directory)] = true
		}
	}
	sortedUniqueDirectories := make([]string, len(uniqueDirectories))

	idx := 0
	for dir := range uniqueDirectories {
		sortedUniqueDirectories[idx] = dir
		idx += 1
	}
	slices.Sort(sortedUniqueDirectories)

	if showDirs {
		for _, dir := range sortedUniqueDirectories {
			fmt.Println(dir)
		}
	}
	for _, meta := range filesToShow {
		fmt.Println(fullName(meta))
	}
	return
}

func fullName(meta *bfsp.FileMetadata) string {
	fullName := ""
	// first, add the directories
	for _, dir := range meta.Directory {
		fullName = fullName + dir + "/"
	}
	// then add the file name
	fullName = fullName + meta.FileName
	return fullName
}

func directoryToSlice(dir string) []string {
	if dir == "/" {
		return []string{}
	}
	dir = strings.TrimPrefix(dir, "/")
	slice := strings.Split(dir, "/")
	for idx := range slice {
		if slice[idx] == " " {
			slice[idx] = ""
		}
	}
	return slice
}

func sliceToDirectory(dir []string) string {
	return "/" + strings.Join(dir, "/")
}

func compareDirSlice(dir1 []string, dir2 []string) bool {
	if len(dir1) != len(dir2) {
		return false
	}

	for idx := range len(dir1) {
		if dir1[idx] != dir2[idx] {
			return false
		}
	}

	return true
}
