package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	guuid "github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	path = "D:/Go/Task/images"
)

func main() {
	setupRoutes()
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func connAndSendDataToRabbitMQ(id, format string) {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"Uid", // name
		false, // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	failOnError(err, "Failed to declare a queue")

	body := id + "," + format
	err = ch.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(body),
		})
	failOnError(err, "Failed to publish a message")
	log.Printf(" [x] Sent %s", body)
}

func uploadFile(w http.ResponseWriter, r *http.Request) {
	var fileName []string
	uId := ""
	fmt.Println("File Upload Endpoint Hit")

	// Parse our multipart form, 10 << 20 specifies a maximum
	// upload of 10 MB files.
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		fmt.Println(err, "Not parse MultipartForm")
		return
	}
	// FormFile returns the first file for the given key `myFile`
	// it also returns the FileHeader so we can get the Filename,
	// the Header and the size of the file
	file, handler, err := r.FormFile("myFile")
	if err != nil {
		fmt.Println("Error Retrieving the File")
		fmt.Println(err)
		return
	}
	defer file.Close()
	fmt.Printf("Uploaded File: %+v\n", handler.Filename)
	fmt.Printf("File Size: %+v\n", handler.Size)
	fmt.Printf("MIME Header: %+v\n", handler.Header)

	fileName = strings.Split(handler.Filename, ".")
	// Create a temporary file within our temp-images directory that follows
	// a particular naming pattern
	uId = genUUID()
	if !exists(path) {
		err := os.Mkdir(path, 0777)
		if err != nil {
			panic(err)
		}
	}
	newFileName := "upload_" + uId + "." + fileName[1]
	f, err := os.Create(filepath.Join(path, newFileName))
	if err != nil {
		panic(err)
	}

	if err != nil {
		fmt.Println(err)
	}
	defer f.Close()

	// read all of the contents of our uploaded file into a
	// byte array
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Println(err)
	}
	// write this byte array to our temporary file
	_, err = f.Write(fileBytes)
	if err != nil {
		fmt.Println(err, "Not write byte array to temporary file")
		return
	}
	// return that we have successfully uploaded our file!
	fmt.Fprintf(w, "Successfully Uploaded File\n")

	if strings.ToLower(fileName[1]) == "png" {
		connAndSendDataToRabbitMQ(uId, "0")
	} else if strings.ToLower(fileName[1]) == "jpg" || strings.ToLower(fileName[1]) == "jpeg" {
		connAndSendDataToRabbitMQ(uId, "1")
	} else {
		fmt.Println("Unkown type for sending")
		return
	}

}

func setupRoutes() {
	http.HandleFunc("/upload", uploadFile)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func genUUID() string {
	id := guuid.New()
	sid := id.String()
	return sid
}

// exists returns whether the given file or directory exists
func exists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	if err != nil {
		fmt.Println(err)
	}
	return false
}
