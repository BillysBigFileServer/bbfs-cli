package cmd

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/BillysBigFileServer/bfsp-go"
	"github.com/BillysBigFileServer/cli/client"
	"github.com/cenkalti/backoff/v4"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
	"lukechampine.com/blake3"
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
	fileInfo, err := file.Stat()
	if err != nil {
		panic(err)
	}

	fileId, err := uuid.NewRandom()
	if err != nil {
		panic(err)
	}

	fmt.Println("uploading file", args[0])

	bfspClient, err := client.NewFileServerClient()
	if err != nil {
		panic(err)
	}

	chunks := sync.Map{}
	totalSize := fileInfo.Size()
	totalUploaded := atomic.Uint64{}

	g := errgroup.Group{}
	g.SetLimit(100)
	for offset := 0; offset < int(fileInfo.Size()); offset += 1024 * 1024 {
		buf := make([]byte, 1024*1024)
		n, err := file.ReadAt(buf, int64(offset))
		if err != nil && !errors.Is(err, io.EOF) {
			panic(err)
		}
		buf = buf[:n]

		g.Go(func() error {
			chunkHash := blake3.Sum256(buf)
			chunkId, err := uuid.NewRandom()
			if err != nil {
				return err
			}
			chunkLen := uint32(len(buf))

			chunkNonce := make([]byte, 24)
			// random bytes for chunk nonce
			_, err = rand.Read(chunkNonce)
			if err != nil {
				return err
			}
			chunkMetadata := &bfsp.ChunkMetadata{
				Id:     chunkId.String(),
				Hash:   chunkHash[:],
				Size:   chunkLen,
				Indice: int64(offset / (1024 * 1024)),
				Nonce:  chunkNonce,
			}

			processecdChunk, err := bfsp.CompressEncryptChunk(buf, chunkMetadata, fileId.String(), masterKey)
			if err != nil {
				return err
			}

			b := backoff.NewExponentialBackOff(backoff.WithMaxElapsedTime(10 * time.Second))
			err = backoff.Retry(func() error {
				return bfsp.UploadChunk(bfspClient, chunkMetadata, fileId.String(), *processecdChunk, masterKey)
			}, b)

			chunks.Store(uint64(chunkMetadata.Indice), chunkMetadata.Id)
			totalUploaded := totalUploaded.Add(uint64(chunkLen))

			percentUploaded := float64(totalUploaded) / float64(totalSize) * 100
			fmt.Fprintf(os.Stderr, "%f%s uploaded\n", percentUploaded, "%")

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		panic(err)
	}

	chunksFileMetadata := map[uint64]string{}
	chunks.Range(func(key, value interface{}) bool {
		chunksFileMetadata[key.(uint64)] = value.(string)
		return true
	})

	currentUnixUTCTime := time.Now().UTC().Unix()
	fileMetadata := &bfsp.FileMetadata{
		Id:               fileId.String(),
		Chunks:           chunksFileMetadata,
		FileName:         fileInfo.Name(),
		FileType:         bfsp.FileType_UNKNOWN,
		FileSize:         uint64(fileInfo.Size()),
		Directory:        []string{},
		CreateTime:       currentUnixUTCTime,
		ModificationTime: currentUnixUTCTime,
	}
	err = bfsp.UploadFileMetadata(bfspClient, fileMetadata, masterKey)
	if err != nil {
		panic(err)
	}

}
