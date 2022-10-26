package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	const nameDir = "static"
	err := os.Mkdir(nameDir, os.ModePerm)

	if err != nil && err.Error() != fmt.Sprintf("mkdir %s: file exists", nameDir) {
		log.Fatal(err)
	}

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	const ADDR = ":3000"

	log.Printf("listening on http://localhost%s", ADDR)
	log.Fatal(http.ListenAndServe(ADDR, nil))
}
