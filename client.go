package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

const bufferSize = 1024

func main() {

	// Handle host:port, image and filter with flags
	host := flag.String("host", "127.0.0.1", "Nom d'hote du serveur")
	port := flag.Int("port", 8080, "Port du serveur")
	filter := flag.Int("filter", 1, "Filtre à utiliser : 1 = grayscale, 2 = edge")
	filePath := flag.String("path", "", "--REQUIRED-- Chemin relatif vers l'image")

	flag.Parse()

	if *filePath == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Connect to the server
	connection, err := net.Dial("tcp", *host+":"+strconv.Itoa(*port))
	if err != nil {
		log.Fatal("tcp dial error", err)
	}

	fmt.Println("Connected to server")
	defer connection.Close()

	//send file
	sendFile(*filePath, *filter, connection)

	//receive file back

	receiveFile(connection)

	fmt.Println("Closing the connection")

}

func sendFile(path string, filter int, connection net.Conn) {
	file, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
		return
	}
	fileInfo, err := file.Stat()
	if err != nil {
		fmt.Println(err)
		return
	}
	filterStr := fillString(strconv.FormatInt(int64(filter), 10), 10)
	fileSize := fillString(strconv.FormatInt(fileInfo.Size(), 10), 10)
	fileName := fillString(fileInfo.Name(), 64)
	fmt.Println("Sending filter, filename and filesize!")
	connection.Write([]byte(filterStr))
	connection.Write([]byte(fileSize))
	connection.Write([]byte(fileName))
	sendBuffer := make([]byte, bufferSize)
	fmt.Println("Start sending file!")
	for {
		_, err = file.Read(sendBuffer)
		if err == io.EOF {
			break
		}
		connection.Write(sendBuffer)
	}
	fmt.Println("File has been sent!")
}

func receiveFile(connection net.Conn) {

	fmt.Println("Start receiving the file back")
	bufferFileName := make([]byte, 64)
	bufferFileSize := make([]byte, 10)

	connection.Read(bufferFileSize)
	fileSize, _ := strconv.ParseInt(strings.Trim(string(bufferFileSize), ":"), 10, 64)

	connection.Read(bufferFileName)
	fileName := strings.Trim(string(bufferFileName), ":")

	newFile, err := os.Create(fileName)

	if err != nil {
		panic(err)
	}
	defer newFile.Close()
	var receivedBytes int64

	for {
		if (fileSize - receivedBytes) < bufferSize {
			io.CopyN(newFile, connection, (fileSize - receivedBytes))
			connection.Read(make([]byte, (receivedBytes+bufferSize)-fileSize))
			break
		}
		io.CopyN(newFile, connection, bufferSize)
		receivedBytes += bufferSize
	}
	fmt.Println("Received file " + fileName + " back")
}

func fillString(returnString string, toLength int) string {
	for {
		lengtString := len(returnString)
		if lengtString < toLength {
			returnString = returnString + ":"
			continue
		}
		break
	}
	return returnString
}
