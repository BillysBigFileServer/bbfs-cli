package client

import (
	"encoding/base64"
	"os"

	"github.com/BillysBigFileServer/bbfs-cli/config"
	"github.com/BillysBigFileServer/bfsp-go"
)

func FileServerBaseURL() string {
	if baseURL := os.Getenv("FILE_SERVER_BASE_URL"); baseURL != "" {
		return baseURL
	} else {
		return "big-file-server.fly.dev:9998"
	}
}

func FileServerHTTPS() bool {
	switch os.Getenv("FILE_SERVER_HTTPS") {
	case "true", "1":
		return true
	case "false", "0":
		return false
	default:
		return true
	}
}

func BigCentralBaseURL() string {
	if baseURL := os.Getenv("BIG_CENTRAL_BASE_URL"); baseURL != "" {
		return baseURL
	} else {
		return "https://bbfs.io"
	}
}

func NewFileServerClient() (bfsp.FileServerClient, error) {
	configFile, err := config.OpenDefaultConfigFile()
	if err != nil {
		return nil, err
	}
	bfspConfig, err := config.ReadConfig(configFile)
	if err != nil {
		return nil, err
	}

	bfspClient, err := bfsp.NewHTTPFileServerClient(bfspConfig.Token, FileServerBaseURL(), FileServerHTTPS())
	if err != nil {
		return nil, err
	}

	return bfspClient, nil
}

func MasterKey() (bfsp.MasterKey, error) {
	configFile, err := config.OpenDefaultConfigFile()
	if err != nil {
		return nil, err
	}
	bfspConfig, err := config.ReadConfig(configFile)
	if err != nil {
		return nil, err
	}

	masterKeyString := bfspConfig.EncryptionKey
	masterKey, err := base64.StdEncoding.DecodeString(masterKeyString)
	if err != nil {
		return nil, err
	}

	return masterKey, err
}

func ShareFile(fileMeta *bfsp.FileMetadata) (string, error) {
	configFile, err := config.OpenDefaultConfigFile()
	if err != nil {
		return "", err
	}
	bfspConfig, err := config.ReadConfig(configFile)
	if err != nil {
		return "", err
	}

	masterKey, err := base64.StdEncoding.DecodeString(bfspConfig.EncryptionKey)
	if err != nil {
		return "", err
	}

	viewInfo, err := bfsp.ShareFile(fileMeta, bfspConfig.Token, masterKey)
	if err != nil {
		return "", err
	}
	viewInfoB64, err := bfsp.EncodeViewFileInfo(viewInfo)
	if err != nil {
		return "", err
	}

	return BigCentralBaseURL() + "/files/view_file#z:" + viewInfoB64, nil
}
