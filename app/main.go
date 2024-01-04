package main

import (
	"context"
	"fmt"
)

func test(sl []string) []string {
	return sl[1:]
}

func main() {
	fmt.Println("starting")
	fmt.Println(getServer().Run(context.Background()))
}
