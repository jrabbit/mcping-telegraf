package main

import (
    "github.com/whatupdave/mcping"
    "fmt"
)

func main() {
    resp, _ := mcping.Ping("brightsight.jumpingcrab.com:25565")
    fmt.Println("Mineplex has", resp.Online, "players online")
}