package cmd

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"os"

	"github.com/BillysBigFileServer/bfsp-go"
	"github.com/BillysBigFileServer/bfsp-go/config"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to BBFS.io",
	Run:   runLogin,
}

func init() {
	rootCmd.AddCommand(loginCmd)
}

func runLogin(cmd *cobra.Command, args []string) {
	dlToken, err := uuid.NewRandom()
	if err != nil {
		panic(err)
	}
	bigCentralURL := config.BigCentralBaseURL()
	privKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		panic(err)
	}

	tempPubEncKeyBytes := x509.MarshalPKCS1PublicKey(&privKey.PublicKey)
	pubKeyString := base64.URLEncoding.EncodeToString(tempPubEncKeyBytes)

	dlTokenUrl := bigCentralURL + "/auth?dl_token=" + dlToken.String() + "#" + pubKeyString
	fmt.Fprintln(os.Stderr, "Visit", dlTokenUrl, "to sign in...")
	fmt.Fprintln(os.Stderr)

	tokenInfo, err := bfsp.GetToken(bigCentralURL, dlToken.String(), privKey)
	if err != nil {
		panic(err)
	}

	configFile, err := config.OpenDefaultConfigFile()
	if err != nil {
		panic(err)
	}
	defer configFile.Close()

	cliConfig, err := config.ReadConfig(configFile)
	if err != nil {
		fmt.Println("invalid config file; resetting")
		cliConfig = &config.Config{}
	}

	masterKey := tokenInfo.MasterKey
	masterKeyString := base64.StdEncoding.EncodeToString(masterKey)

	cliConfig.EncryptionKey = masterKeyString
	cliConfig.Token = tokenInfo.Token

	err = config.WriteConfigToFile(configFile, cliConfig)
	if err != nil {
		panic(err)
	}

	fmt.Println("Successfully signed in")
}
