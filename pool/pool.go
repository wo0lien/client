package pool

import (
	"fmt"
	"github.com/wo0lien/client/imagetools"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
)

const bufferSize = 1024

/*
Pool start a pool of TCP client
*/
func Pool(nb, filter, port int, host, filePath string) {

	img, err := imagetools.Open(filePath)
	if err != nil {
		fmt.Printf("Error: %+v", err)
		return
	}

	fmt.Printf("test")
	for i := 0; i < nb; i++ {
		id := strconv.Itoa(i)
		createDirIfNotExist(id)

		err = imagetools.Export(img, "./"+id+"/img.png")

		if err != nil {
			fmt.Printf("Error: %+v", err)
			return
		}
	}

	if err != nil {
		fmt.Printf("Error: %+v", err)
		return
	}

	var wg sync.WaitGroup

	for i := 0; i < nb; i++ {
		wg.Add(1)
		go StartClient(i, filter, port, host, &wg)
	}

	wg.Wait()

}

/*
StartClient start a TCP client
*/
func StartClient(id, filter, port int, host string, wg *sync.WaitGroup) {
	// Connect to the server

	defer wg.Done()

	path := strconv.Itoa(id)

	connection, err := net.Dial("tcp", host+":"+strconv.Itoa(port))
	if err != nil {
		log.Fatal("tcp dial error", err)
	}

	fmt.Println("Connected to server")
	defer connection.Close()

	//send file
	sendFile("./"+path+"/img.png", filter, connection)

	//receive file back

	receiveFile(id, connection)

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

func receiveFile(id int, connection net.Conn) {

	fmt.Println("Start receiving the file back")
	bufferFileName := make([]byte, 64)
	bufferFileSize := make([]byte, 10)

	connection.Read(bufferFileSize)
	fileSize, _ := strconv.ParseInt(strings.Trim(string(bufferFileSize), ":"), 10, 64)

	connection.Read(bufferFileName)
	fileName := strings.Trim(string(bufferFileName), ":")

	newFile, err := os.Create("./" + strconv.Itoa(id) + "/" + fileName)

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

func createDirIfNotExist(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			panic(err)
		}
	}
}
