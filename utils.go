package main

import (
	"crypto/subtle"
	"crypto/tls"
	"encoding/binary"
	"errors"
	"github.com/wneessen/go-fileperm"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"path"
	"syscall"
	"time"
)

var stopSignals = [...]os.Signal{syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGABRT, syscall.SIGINT}

var (
	wrongMagicBytesErr = errors.New("wrong magic bytes received")
	wrongPassErr       = errors.New("wrong password received")
	passTooLongErr     = errors.New("too long password")
	successBytesErr    = errors.New("success bytes not received")
	stringTooLongErr   = errors.New("too long string")
	clientSideErr      = errors.New("error occurred on client side while receiving file")
)

const magicBytes = "\x12\x34\x13\x37\x13\x57\x90\xf7\x24\x68\xc2\x8e"
const successBytes = "\x45\x86\xb7\x34\x98\xf2\xff\xa7"
const maxPassLength = 256
const maxStringLength = 512

func getExitChanForServer() chan os.Signal {
	exitChan := make(chan os.Signal)
	signal.Notify(exitChan, stopSignals[:]...)
	return exitChan
}

func currentWorkingDirectoryIsWritable() (bool, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return false, err
	}
	up, err := fileperm.New(cwd)
	if err != nil {
		return false, err
	}
	return up.UserWritable(), nil
}

func fileIsReadable(name string) (bool, error) {
	up, err := fileperm.New(name)
	if err != nil {
		return false, err
	}
	return up.UserReadable(), nil
}

func getNewListenerWithTls(listenAddr string) (net.Listener, error) {
	config := &tls.Config{}
	config.Certificates = make([]tls.Certificate, 1)
	config.Certificates[0], _ = getRandomTLS(4096)
	l, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return nil, err
	}
	return tls.NewListener(l, config), nil
}

func writeFile(conn net.Conn, filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		_ = writeString(conn, "")
		return err
	}
	defer file.Close()
	fileStat, err := file.Stat()
	if err != nil {
		_ = writeString(conn, "")
		return err
	}
	err = writeString(conn, path.Base(file.Name()))
	if err != nil {
		return err
	}
	err = writeUint64(conn, uint64(fileStat.Size()))
	if err != nil {
		return err
	}
	_, err = io.Copy(conn, file)
	return err
}

func readFile(conn net.Conn) error {
	filename, err := readString(conn)
	if err != nil {
		return err
	}
	if filename == "" {
		return clientSideErr
	}
	fileSize, err := readUint64(conn)
	if err != nil {
		return err
	}
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.CopyN(file, conn, int64(fileSize))
	if err != nil {
		return err
	}
	log.Println("New file received:", filename)
	return nil
}

func readPass(conn net.Conn, pass string) error {
	err := conn.SetReadDeadline(time.Now().Add(time.Second * 5))
	if err != nil {
		return err
	}
	magicBytesReceived, err := readString(conn)
	if err != nil {
		return err
	}
	if magicBytesReceived != magicBytes {
		return wrongMagicBytesErr
	}
	passReceived, err := readString(conn)
	if subtle.ConstantTimeCompare([]byte(passReceived), []byte(pass)) == 0 {
		return wrongPassErr
	}
	err = conn.SetReadDeadline(time.Time{})
	if err != nil {
		return err
	}
	err = conn.SetWriteDeadline(time.Now().Add(time.Second * 5))
	if err != nil {
		return err
	}
	err = writeString(conn, successBytes)
	if err != nil {
		return err
	}
	return conn.SetWriteDeadline(time.Time{})
}

func writePass(conn net.Conn, pass string) error {
	if len(pass) > maxPassLength {
		return passTooLongErr
	}
	err := conn.SetWriteDeadline(time.Now().Add(time.Second * 5))
	if err != nil {
		return err
	}
	err = writeString(conn, magicBytes)
	if err != nil {
		return err
	}
	err = writeString(conn, pass)
	if err != nil {
		return err
	}
	err = conn.SetWriteDeadline(time.Time{})
	if err != nil {
		return err
	}
	err = conn.SetReadDeadline(time.Now().Add(time.Second * 5))
	if err != nil {
		return err
	}
	successMsg, err := readString(conn)
	if err != nil {
		return err
	}
	if successMsg != successBytes {
		return successBytesErr
	}
	return conn.SetReadDeadline(time.Time{})
}

func writeString(conn net.Conn, str string) error {
	if len(str) > maxStringLength {
		return stringTooLongErr
	}
	err := writeUint32(conn, uint32(len(str)))
	if err != nil {
		return err
	}
	_, err = conn.Write([]byte(str))
	return err
}

func readString(conn net.Conn) (string, error) {
	strLength, err := readUint32(conn)
	if err != nil {
		return "", err
	}
	if strLength > maxStringLength {
		return "", stringTooLongErr
	}
	strBuffer := make([]byte, strLength)
	_, err = io.ReadFull(conn, strBuffer)
	if err != nil {
		return "", err
	}
	return string(strBuffer), nil
}

func writeUint32(conn net.Conn, i uint32) error {
	var buffer [4]byte
	binary.LittleEndian.PutUint32(buffer[:], i)
	_, err := conn.Write(buffer[:])
	return err
}

func readUint32(conn net.Conn) (uint32, error) {
	var buffer [4]byte
	_, err := io.ReadFull(conn, buffer[:])
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(buffer[:]), nil
}

func writeUint64(conn net.Conn, i uint64) error {
	var buffer [8]byte
	binary.LittleEndian.PutUint64(buffer[:], i)
	_, err := conn.Write(buffer[:])
	return err
}

func readUint64(conn net.Conn) (uint64, error) {
	var buffer [8]byte
	_, err := io.ReadFull(conn, buffer[:])
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint64(buffer[:]), nil
}
