package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"sync/atomic"

	"github.com/BillysBigFileServer/bfsp-go"
	"github.com/BillysBigFileServer/cli/client"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

// listCmd represents the list command
var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Downloads a file from BBFS.io to your local drive",
	Run:   runDownload,
}

func init() {
	downloadCmd.Flags().StringP("output-directory", "o", ".", "")
	downloadCmd.Flags().BoolP("show-progress", "P", true, "")
	rootCmd.AddCommand(downloadCmd)
}

func runDownload(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		fmt.Println("Usage: download <file>")
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

	fileMeta, err := getFileMetadata(bfspClient, args[0], masterKey)
	if err != nil {
		panic(err)
	}

	if fileMeta == nil {
		fmt.Println("file not found")
		return
	}

	fmt.Fprintln(os.Stderr, "Downloading", fullName(fileMeta))

	chunkIndices := []uint64{}

	for chunkIndice := range fileMeta.Chunks {
		chunkIndices = append(chunkIndices, chunkIndice)
	}
	sort.Slice(chunkIndices, func(i, j int) bool { return chunkIndices[i] < chunkIndices[j] })

	outputDir, err := cmd.Flags().GetString("output-directory")
	if err != nil {
		panic(err)
	}
	showProgress, err := cmd.Flags().GetBool("show-progress")
	fullPath := outputDir + "/" + fileMeta.FileName

	_, err = os.OpenFile(fullPath, 0, 0666)
	if !os.IsNotExist(err) {
		fmt.Println("file already exists")
		return
	}

	file, err := os.Create(fullPath + ".part")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	err = file.Truncate(int64(fileMeta.FileSize))
	if err != nil {
		panic(err)
	}

	totalUploaded := atomic.Uint64{}
	g := errgroup.Group{}
	g.SetLimit(100)
	for _, indice := range chunkIndices {
		indice := indice
		err := func() error {
			chunkId := fileMeta.Chunks[indice]
			chunk, err := bfsp.DownloadChunk(bfspClient, chunkId, fileMeta.Id, masterKey)
			if err != nil {
				return err
			}

			_, err = file.WriteAt(chunk, int64(indice)*1024*1024)
			if err != nil {
				return err
			}

			totalUploaded := totalUploaded.Add(uint64(len(chunk)))
			percentUploaded := float64(totalUploaded) / float64(fileMeta.FileSize) * 100
			if showProgress {
				fmt.Fprintf(os.Stderr, "\r\033[K%f%s downloaded", percentUploaded, "%")
				os.Stderr.Sync()
			}

			return nil
		}()
		if err != nil {
			panic(err)
		}
	}
	fmt.Fprintln(os.Stderr)
	if err := g.Wait(); err != nil {
		panic(err)
	}
	err = os.Rename(fullPath+".part", fullPath)
	if err != nil {
		panic(err)
	}

	return
}

func getFileMetadata(bfspClient bfsp.FileServerClient, file string, masterKey bfsp.MasterKey) (*bfsp.FileMetadata, error) {
	switch strings.HasPrefix(file, "https://") {
	case true:
		return fileMetadataFromURL(bfspClient, file, masterKey)
	default:
		return fileMetadataFromName(bfspClient, file, masterKey)
	}
}

func fileMetadataFromURL(bfspClient bfsp.FileServerClient, fileURL string, masterKey bfsp.MasterKey) (*bfsp.FileMetadata, error) {
	parts := strings.Split(fileURL, "https://bbfs.io/files/view_file#z:")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid url")
	}

	viewFileInfoB64 := parts[1]
	viewFileInfo, err := bfsp.DecodeViewFileInfoB64(viewFileInfoB64)
	if err != nil {
		return nil, err
	}

	fileMeta, err := bfsp.DownloadFileMetadata(bfspClient, viewFileInfo.Id, masterKey)
	if err != nil {
		return nil, err
	}

	return fileMeta, nil
}

func fileMetadataFromName(bfspClient bfsp.FileServerClient, fileName string, masterKey bfsp.MasterKey) (*bfsp.FileMetadata, error) {
	files, err := bfsp.ListFileMetadata(bfspClient, []string{}, masterKey)
	if err != nil {
		return nil, err
	}

	var fileMeta *bfsp.FileMetadata
	for _, meta := range files {
		if fullName(meta) == fileName {
			fileMeta = meta
			break
		}
	}

	return fileMeta, nil

}
