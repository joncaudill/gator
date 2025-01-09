package main

import (
	"fmt"
	"internal/config"
)

func main() {
	cfg, err := config.Read()
	if err != nil {
		fmt.Println("error reading config file:", err)
		return
	}
	err = cfg.SetUser("joncaudill")
	if err != nil {
		fmt.Println("error setting user name:", err)
		return
	}
	cfg, err = config.Read()
	if err != nil {
		fmt.Println("error reading config file:", err)
		return
	}
	fmt.Printf("current config:%v\n", cfg)
}
