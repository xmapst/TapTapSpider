package main

import (
	"log"
	"os"
)

func init() {
	log.SetFlags(log.Ldate | log.Lmicroseconds)
}

func main() {
	_ = os.Remove("output")
	_ = os.MkdirAll("output", 0777)
	if err := NewSpider().Run(); err != nil {
		log.Fatalln(err)
	}
}
