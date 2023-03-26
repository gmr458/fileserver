package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
)

var (
	staticDirName string
	port          int
)

func main() {
	flag.IntVar(&port, "port", 4000, "Server port")
	flag.StringVar(&staticDirName, "dir", "", "Directory to expose")
	flag.Parse()
	if staticDirName == "" {
		fmt.Println("Pass the -dir flag")
		os.Exit(1)
	}

	fs := http.FileServer(http.Dir(staticDirName))
	http.Handle("/", fs)
	http.HandleFunc("/downloadall", handlerDownloadAll)

	fmt.Printf("Local: http://localhost:%d/\n", port)

	ips := getIps()
	networkLinks := ""
	for k, ip := range ips {
		networkLinks += fmt.Sprintf("http://%s:%d/", ip, port)

		if k == len(ips)-1 {
			break
		}

		networkLinks += ", "
	}
	fmt.Printf("Network: %s\n", networkLinks)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}

func getIps() []string {
	interfaces, err := net.Interfaces()
	if err != nil {
		log.Fatal(err)
	}

	var ips []string

	for _, i := range interfaces {
		addrs, err := i.Addrs()
		if err != nil {
			log.Fatal(err)
		}

		for k, a := range addrs {
			if k == 0 {
				if !strings.HasPrefix(a.String(), "127") {
					ip := strings.Split(a.String(), "/")[0]
					ips = append(ips, ip)
				}
			}
		}
	}

	return ips
}

func handlerDownloadAll(w http.ResponseWriter, r *http.Request) {
	dirEntries, err := os.ReadDir(staticDirName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	ips := getIps()

	var text string

	for _, ip := range ips {
		text += "--------------------------------------\n"
		text += fmt.Sprintf("Links available for IP %s\n\n", ip)

		cmd := "wget"
		for k, v := range dirEntries {
			filename := v.Name()
			cmd += fmt.Sprintf(" \"http://%s:%d/%s\"", ip, port, filename)

			if k == len(dirEntries)-1 {
				break
			}

			cmd += " \\\n"
		}

		text += cmd
	}

	_, err = w.Write([]byte(text))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
