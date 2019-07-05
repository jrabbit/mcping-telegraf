package main

import (
	"fmt"
	"github.com/mdaffin/go-telegraf"
	"github.com/spf13/viper"
	"github.com/whatupdave/mcping"
	"log"
	"os"
	"runtime"
)

const version = "0.0.1"

func main() {
	fmt.Printf("mcping-bin version %s\n", version)
	// ref https://github.com/pallets/click/blob/4da5e93cede17262424671208799bc6921dcfa36/click/utils.py#L368-L417
	viper.AddConfigPath("$HOME/.config/")
	viper.AddConfigPath("$HOME/Library/Application Support/")
	cwd, _ := os.Getwd()
	viper.AddConfigPath(cwd)
	if runtime.GOOS == "windows" {
		winAppData := os.Getenv("LOCALAPPDATA")
		viper.AddConfigPath(winAppData)
	}
	viper.SetConfigName("mcping")

	viper.SetDefault("telegraf_server", "localhost:8094")
	viper.SetDefault("miceraft_server", "localhost:25565")
	viper.ReadInConfig()

	grafServer := viper.GetString("telegraf_server")
	log.Printf("mcping config file: %s", viper.ConfigFileUsed())
	client, tel_err := telegraf.NewTCP(grafServer)
	if tel_err != nil {
		log.Fatalf("telegraf fail: %s", tel_err)
	}
	defer client.Close()
	mcServer := viper.GetString("minecraft_server")
	resp, mcErr := mcping.Ping(mcServer)
	if mcErr != nil {
		log.Fatalf("minecraft fail: %s", mcErr)
	}
	m := telegraf.MeasureInt("mcping-go", "online", resp.Online)
	write_err := client.Write(m)
	if write_err != nil {
		log.Fatalf("telegraf write fail: %s", write_err)
	}
	log.Println("Mineplex has", resp.Online, "players online")
}
