package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/aaronland/go-http-crumb"
	"log"
)

func main() {

	ttl := flag.Int("ttl", 3600, "Time to live (in seconds)")
	key := flag.String("key", "", "Optional key to use when generating crumb base")

	flag.Parse()

	ctx := context.Background()

	uri, err := crumb.NewRandomEncryptedCrumbURI(ctx, *ttl, *key)

	if err != nil {
		log.Fatalf("Failed to generate crumb URI, %v", err)
	}

	fmt.Println(uri)
}
