package main

import (
	"github.com/spf13/cobra"
	"log"
)

var action1Config string

var sub1Cmd = &cobra.Command{
	Use:   "sub1",
	Short: "sub1 commands short",
}

var sub1Action1Cmd = &cobra.Command{
	Use: "action1",
	Short: "action1 commands short",
	Args: cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		log.Printf("Sub1-Action1 is called. args: %v", args)
	},
}

func init() {
	log.Println("sub1's init is called")
	sub1Action1Cmd.PersistentFlags().StringVarP(&action1Config, "conf", "c", "", "action1 configs")
	sub1Cmd.AddCommand(sub1Action1Cmd)
	rootCmd.AddCommand(sub1Cmd)
}
