package main

import (
	"fmt"
	"github.com/mdaffin/go-telegraf"
	"github.com/spf13/viper"
	"github.com/whatupdave/mcping"
	"log"
)

const version = "0.0.1"

func main() {
	fmt.Printf("mcping-bin version %s\n", version)
	viper.AddConfigPath("$HOME/.config/")
	viper.SetConfigName("mcping")
	viper.SetDefault("telegraf_server", "localhost:8094")
	viper.SetDefault("miceraft_server", "localhost:25565")
	viper.ReadInConfig()
	var grafServer string
	grafServer = viper.GetString("telegraf_server")
	log.Printf("mcping config file: %s", viper.ConfigFileUsed())
	client, tel_err := telegraf.NewTCP(grafServer)
	if tel_err != nil {
		log.Fatalf("telegraf fail: %s", tel_err)
	}
	defer client.Close()
	mcServer := viper.GetString("minecraft_server")
	resp, mcErr := mcping.Ping(mcServer)
	if mcErr != nil {
		log.Fatalf("minecraft fail: %s", err)
	}
	m := telegraf.MeasureInt("mcping-go", "online", resp.Online)
	write_err := client.Write(m)
	if write_err != nil {
		log.Fatalf("telegraf write fail: %s", write_err)
	}
	log.Println("Mineplex has", resp.Online, "players online")
}
