package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"os"
)

var configPath string

var rootCmd = &cobra.Command{
	Use:     "mycommand",
	Short:   "root command's short message",
	Version: "1.0.0",
	//Run: func(cmd *cobra.Command, args []string) {
	//	log.Println("Execute RootCommand")
	//	log.Println("> Config path:", configPath)
	//},
}

func init() {
	log.Println("main's init is called")
	rootCmd.PersistentFlags().StringVar(&configPath, "config", "", "config file path")
	cobra.OnInitialize(onInit)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
		return
	}
	os.Exit(0)
}

func onInit() {
	log.Println("onInit is called")
}