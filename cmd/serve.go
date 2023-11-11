/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/ophum/prometheus-http-sd-sakuracloud/internal/handler"
	"github.com/sacloud/iaas-api-go"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type serveConfig struct {
	Token  string
	Secret string
	Zone   string
}

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: serveRun,
}

func init() {
	rootCmd.AddCommand(serveCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// serveCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// serveCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func serveRun(cmd *cobra.Command, args []string) {
	config := serveConfig{}
	if err := viper.Unmarshal(&config); err != nil {
		log.Fatal(err)
	}

	sacloudClient := iaas.NewClient(config.Token, config.Secret)
	h := handler.NewServiceDiscovery(sacloudClient, config.Zone)

	r := gin.Default()
	r.Use(handler.ErrorMiddleware())

	r.GET("discovery/server", h.DiscoveryServer())
	r.GET("discovery/loadbalancer", h.DiscoveryLoadbalancer())

	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
