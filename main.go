package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
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
	networkLinks := formatIPs(ips)
	if networkLinks == "" {
		fmt.Println("Network: Not connected to a local network")
	} else {
		fmt.Printf("Network: %s\n", networkLinks)
	}
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}

func formatIPs(ips []string) string {
	networkLinks := ""

	for k, ip := range ips {
		networkLinks += fmt.Sprintf("http://%s:%d/", ip, port)

		if k == len(ips)-1 {
			break
		}

		networkLinks += ", "
	}

	return networkLinks
}

func isIPV4(ip string) bool {
	if len(ip) < 7 || len(ip) > 15 {
		return false
	}

	if strings.Count(ip, ".") != 3 {
		return false
	}

	octets := strings.Split(ip, ".")

	if len(octets) != 4 {
		return false
	}

	first, err := strconv.Atoi(octets[0])
	if err != nil {
		return false
	}
	if first < 0 || first > 255 {
		return false
	}

	second, err := strconv.Atoi(octets[1])
	if err != nil {
		return false
	}
	if second < 0 || second > 255 {
		return false
	}

	third, err := strconv.Atoi(octets[2])
	if err != nil {
		return false
	}
	if third < 0 || third > 255 {
		return false
	}

	fourth, err := strconv.Atoi(octets[3])
	if err != nil {
		return false
	}
	if fourth < 0 || fourth > 255 {
		return false
	}

	return true
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

		for _, a := range addrs {
			ip := strings.Split(a.String(), "/")[0]
			flags := i.Flags.String()

			if isIPV4(ip) &&
				strings.Contains(flags, "up") &&
				strings.Contains(flags, "broadcast") &&
				strings.Contains(flags, "multicast") &&
				!strings.Contains(flags, "loopback") {
				ips = append(ips, ip)
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
