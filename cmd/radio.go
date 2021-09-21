package main

import "github.com/ImDevinC/go-radio/pkg/discord"

func main() {
	client, err := discord.NewClient("ODg5NzMzMTQ4ODg2MzMxNDAy.YUlikQ.bc_-iOJwPRhs_5NMK0X9hqLviXg")
	if err != nil {
		panic(err)
	}
	client.Run()
}
