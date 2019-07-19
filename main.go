package main

import (
	"context"
	"fmt"
	"github.com/influxdata/influxdb-client-go"
	"github.com/spf13/viper"
	"github.com/whatupdave/mcping"
	"log"
	"net/http"
	"os"
	"runtime"
	"time"
)

const version = "0.0.1"

func DoMeasures(resp mcping.PingResponse, client influxdb.Client) error {
	// submit all the fields of the ping to the telegraf tcp line
	hostname, _ := os.Hostname()
	myMetrics := []influxdb.Metric{
		influxdb.NewRowMetric(
			map[string]interface{}{"online": 1}, "mcping",
			map[string]string{"hostname": hostname},
			time.Date(2018, 3, 4, 5, 6, 7, 8, time.UTC)),
	}

	write_err := client.Write(context.Background(), "mcping-go", "server_A", myMetrics...)
	if write_err != nil {
		return write_err
	}
	return nil
}

func main() {
	fmt.Printf("mcping-bin version %s\n", version)
	// ref https://github.com/pallets/click/blob/4da5e93cede17262424671208799bc6921dcfa36/click/utils.py#L368-L417
	viper.AddConfigPath("$HOME/.config/")
	viper.AddConfigPath("$HOME/Library/Application Support/")
	viper.AddConfigPath(os.Getenv("MCPING_CONF_DIR"))
	cwd, _ := os.Getwd()
	viper.AddConfigPath(cwd)
	if runtime.GOOS == "windows" {
		winAppData := os.Getenv("LOCALAPPDATA")
		viper.AddConfigPath(winAppData)
	}
	viper.SetConfigName("mcping")

	viper.SetDefault("telegraf_server", "localhost:8094")
	viper.SetDefault("minecraft_server", "localhost:25565")
	viper.ReadInConfig()
	myToken := viper.GetString("telegraf_token")
	grafServer := viper.GetString("telegraf_server")
	httpClient := &http.Client{}
	influx, influx_err := influxdb.New(httpClient, influxdb.WithAddress(grafServer), influxdb.WithToken(myToken))
	if influx_err != nil {
		log.Fatalf("influx fail: %s", influx_err)
	}
	log.Printf("mcping config file: %s", viper.ConfigFileUsed())
	defer influx.Close()

	mcServer := viper.GetString("minecraft_server")
	resp, mcErr := mcping.Ping(mcServer)
	if mcErr != nil {
		log.Printf("minecraft fail: %s", mcErr)
		log.Printf("minecraft host tried: %s", mcServer)
	}
	log.Println("Mineplex has", resp.Online, "players online")

	measure_err := DoMeasures(resp, *influx)
	if measure_err != nil {
		log.Fatalf("telegraf measure fail: %s", measure_err)
	}
}
