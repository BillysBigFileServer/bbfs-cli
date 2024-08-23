package cmd

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/BillysBigFileServer/bbfs-cli/client"
	"github.com/BillysBigFileServer/bbfs-cli/config"
	"github.com/BillysBigFileServer/bfsp-go"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"golang.org/x/term"
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
	bigCentralURL := client.BigCentralBaseURL()
	dlTokenUrl := bigCentralURL + "/auth?dl_token=" + dlToken.String()

	fmt.Fprintln(os.Stderr, "Visit", dlTokenUrl, "to sign in...")
	fmt.Fprintln(os.Stderr)

	token, err := bfsp.GetDLToken(bigCentralURL, dlToken.String())
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

	fmt.Print("Enter your password: ")
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		panic(err)
	}
	password := strings.TrimSpace(string(bytePassword))

	masterKey, err := bfsp.CreateMasterEncKey(password)
	if err != nil {
		panic(err)
	}

	masterKeyString := base64.StdEncoding.EncodeToString(masterKey)

	cliConfig.EncryptionKey = masterKeyString
	cliConfig.Token = token

	err = config.WriteConfigToFile(configFile, cliConfig)
	if err != nil {
		panic(err)
	}

	fmt.Println("Successfully signed in")
}
