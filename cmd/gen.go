/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log"

	"github.com/hantsaniala/hStream/pkg/gen"
	"github.com/hantsaniala/hStream/pkg/utils"
	"github.com/spf13/cobra"
)

// genCmd represents the gen command
var genCmd = &cobra.Command{
	Use:   "gen",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("gen called")
	},
}

var genKeyCmd = &cobra.Command{
	Use:   "key",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := gen.GenKey(utils.GetEnv("KEYMASTER_FOLDER"))
		if err != nil {
			log.Fatal(err.Error())
		}
	},
}

var genKeyPairCmd = &cobra.Command{
	Use:   "keypair",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		gen.GenKeyPair(utils.GetEnv("KEYMASTER_FOLDER"))
	},
}

var genKeyInfoCmd = &cobra.Command{
	Use:   "keyinfo",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		gen.GenKeyinfo(utils.GetEnv("KEYMASTER_FOLDER"), fmt.Sprintf("%s/api/v1/key/%s/%s", utils.GetEnv("HOST"), "master", utils.GetEnv("KEY")))
	},
}

func init() {
	rootCmd.AddCommand(genCmd)
	genCmd.AddCommand(genKeyCmd)
	genCmd.AddCommand(genKeyInfoCmd)
	genCmd.AddCommand(genKeyPairCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// genCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// genCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
