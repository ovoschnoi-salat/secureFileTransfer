package main

import (
	"errors"
	"io"
	"log"
	"net"
	"sync"
)

func listenForClients(listenAddr string) error {
	ln, err := getNewListenerWithTls(listenAddr)
	if err != nil {
		return err
	}
	log.Println("Server will start listening clients on", listenAddr)
	exitChan := getExitChanForServer()
	connChan := make(chan net.Conn, 1)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				if errors.Is(err, net.ErrClosed) {
					break
				}
				log.Println("Error accepting new connection:", err)
			}
			connChan <- conn
		}
		wg.Done()
	}()
loop:
	for {
		select {
		case <-exitChan:
			_ = ln.Close()
			break loop
		case conn := <-connChan:
			wg.Add(1)
			go handleConn(&wg, conn)
		}
	}
	wg.Wait()
	return nil
}

func handleConn(wg *sync.WaitGroup, conn net.Conn) {
	defer wg.Done()
	defer conn.Close()
	res := readPass(conn, passArg)
	if res != nil {
		log.Println("Error handling new connection:", res)
		return
	} else {
		log.Println("New stream initialized")
	}
	var controlByte [1]byte
	for {
		_, err := io.ReadFull(conn, controlByte[:])
		if err != nil {
			log.Println("Error reading control byte:", err)
			return
		}
		switch controlByte[0] {
		case 'f':
			err := readFile(conn)
			if err != nil {
				log.Println("Error receiving file:", err)
				return
			}
		case 'e':
			return
		default:
			log.Println("Error: unknown control byte received:", controlByte[0])
			return
		}
	}
}
