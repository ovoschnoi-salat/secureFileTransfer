package main

import (
	"flag"
	"fmt"
	"log"
	"os/signal"
	"syscall"
)

const defaultPort = ":5000" // todo add support for changing port

var passArg string

func main() {
	signal.Ignore(syscall.SIGHUP)
	flag.Usage = func() {
		fmt.Print("Usage of secureFileTransfer as a server:\n" +
			"secureFileTransfer [--pass <value>]\n\n" +
			"Usage of secureFileTransfer as a client:\n" +
			"secureFileTransfer --server <value> [--pass <value>] <filename> [<filename> ...]\n")
	}
	var serverArg string
	flag.StringVar(&serverArg, "server", "", "")
	flag.StringVar(&passArg, "pass", "", "")
	flag.Parse()
	filenames := flag.Args()
	if (serverArg != "") != (len(filenames) > 0) {
		log.Fatalln("Error: server address and filenames should be passed at the same time")
	}
	if serverArg != "" {
		err := connectAndSendFilesToServer(serverArg, passArg, filenames)
		if err != nil {
			log.Fatalln("Client error:", err)
		} else {
			log.Println("Success")
		}
	} else {
		res, err := currentWorkingDirectoryIsWritable()
		if err != nil {
			log.Fatalln("Error checking cwd write permissions:", err)
		}
		if !res {
			log.Fatalln("Error: no rights to write files to cwd")
		}
		err = listenForClients(defaultPort)
		if err != nil {
			log.Fatalln("Server error:", err)
		}
	}
}
