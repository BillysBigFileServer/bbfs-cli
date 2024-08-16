package client

import (
	"encoding/base64"

	"github.com/BillysBigFileServer/bfsp-go"
	"github.com/BillysBigFileServer/cli/config"
)

func NewFileServerClient() (bfsp.FileServerClient, error) {
	configFile, err := config.OpenDefaultConfigFile()
	if err != nil {
		return nil, err
	}
	bfspConfig, err := config.ReadConfig(configFile)
	if err != nil {
		return nil, err
	}

	//bfspClient, err := bfsp.NewHTTPFileServerClient(bfspConfig.Token, "localhost:9998", false)
	bfspClient, err := bfsp.NewHTTPFileServerClient(bfspConfig.Token, "big-file-server.fly.dev:9998", true)
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

	return "https://bbfs.io/files/view_file#z:" + viewInfoB64, nil
}
