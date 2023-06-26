package main

import (
	"crypto/tls"
	"errors"
	"log"
)

func connectAndSendFilesToServer(serverAddr string, pass string, files []string) error {
	conn, err := tls.Dial("tcp", serverAddr, &tls.Config{InsecureSkipVerify: true})
	if err != nil {
		return err
	}
	defer conn.Close()
	err = writePass(conn, pass)
	if err != nil {
		return err
	}
	for _, file := range files {
		res, err := fileIsReadable(file)
		if err != nil {
			_, _ = conn.Write([]byte("e"))
			return err
		}
		if !res {
			_, _ = conn.Write([]byte("e"))
			return errors.New("file " + file + " is not readable")
		}
		_, err = conn.Write([]byte("f"))
		if err != nil {
			return err
		}
		err = writeFile(conn, file)
		if err != nil {
			_, _ = conn.Write([]byte("e"))
			return err
		}
		log.Println("File successfully sent:", file)
	}
	_, err = conn.Write([]byte("e"))
	return err
}
