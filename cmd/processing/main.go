package main

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"os"

	//"path/filepath"
	//"strconv"
	"strings"

	"github.com/nfnt/resize"
	amqp "github.com/rabbitmq/amqp091-go"
)

var (
	Uid    string
	format string
)

const (
	pathIn  = "D:/Go/Task/images/upload_"
	pathOut = "D:/Go/Task/images/upload_"
)

type InputArgs struct {
	OutputPath string /** Output directory */
	LocalPath  string /** Enter the directory or file path */
	Quality    int    /** Quality */
	Width      int    /** Width dimension, pixel unit */
}

var inputArgs *InputArgs

func NewInputArgs(OutputPath string, LocalPath string, Quality int, Width int) *InputArgs {
	return &InputArgs{
		OutputPath: OutputPath,
		LocalPath:  LocalPath,
		Quality:    Quality,
		Width:      Width,
	}
}

func connToRabbitMQ() {
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

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			sBody := strings.Split(string(d.Body), ",")
			fmt.Println(sBody)
			Uid = sBody[0]
			format = sBody[1]
			log.Printf("Received a message: %s", d.Body)
			inputArgs = NewInputArgs(pathOut+Uid+".", pathIn+Uid+"."+format, 70, 300)
			execute()
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}

func main() {
	connToRabbitMQ()
}

/** Is it a picture */
func isPictureFormat(path string) string {
	temp := strings.Split(path, ".")
	if len(temp) <= 1 {
		return ""
	}
	mapRule := make(map[string]int64)
	mapRule["jpg"] = 1
	mapRule["png"] = 1
	mapRule["jpeg"] = 1
	// fmt.Println(temp[1]+"---")
	/** Add other formats */
	if mapRule[temp[1]] == 1 {
		println(temp[1])
		return temp[1]
	} else {
		return ""
	}
}

func execute() {
	format := isPictureFormat(inputArgs.LocalPath)
	/** Single */
	/** If the input file, then it is single, allowing custom path */
	fmt.Println("Start single sheet compression...")
	if !exists("../../images") {
		err := os.Mkdir("../../images", 0777)
		if err != nil {
			panic(err)
		}
	}
	inputArgs.OutputPath = inputArgs.OutputPath + format
	fmt.Println("OutputPath", inputArgs.OutputPath)
	if !imageCompress(
		func() (io.Reader, error) {
			return os.Open(inputArgs.LocalPath)
		},
		func() (*os.File, error) {
			return os.Open(inputArgs.LocalPath)
		},
		inputArgs.OutputPath,
		inputArgs.Quality,
		inputArgs.Width, format) {
		fmt.Println("Failed to generate thumbnail")
	} else {
		fmt.Println("Thumbnail generated successfully " + inputArgs.OutputPath)
		return
	}
	//}
}

func imageCompress(
	getReadSizeFile func() (io.Reader, error),
	getDecodeFile func() (*os.File, error),
	to string,
	Quality,
	base int,
	format string) bool {
	/** Read file */
	file_origin, err := getDecodeFile()
	defer file_origin.Close()
	if err != nil {
		fmt.Println("os.Open(file) error")
		log.Fatal(err)
		return false
	}
	var origin image.Image
	var config image.Config
	var temp io.Reader
	/** Read size */
	temp, err = getReadSizeFile()
	if err != nil {
		fmt.Println("os.Open(temp)")
		log.Fatal(err)
		return false
	}
	var typeImage int64
	format = strings.ToLower(format)
	/** jpg format */
	if format == "jpg" || format == "jpeg" {
		typeImage = 1
		origin, err = jpeg.Decode(file_origin)
		if err != nil {
			fmt.Println("jpeg.Decode(file_origin)")
			log.Fatal(err)
			return false
		}
		temp, err = getReadSizeFile()
		if err != nil {
			fmt.Println("os.Open(temp)")
			log.Fatal(err)
			return false
		}
		config, err = jpeg.DecodeConfig(temp)
		if err != nil {
			fmt.Println("jpeg.DecodeConfig(temp)")
			return false
		}
	} else if format == "png" {
		typeImage = 0
		origin, err = png.Decode(file_origin)
		if err != nil {
			fmt.Println("png.Decode(file_origin)")
			log.Fatal(err)
			return false
		}
		temp, err = getReadSizeFile()
		if err != nil {
			fmt.Println("os.Open(temp)")
			log.Fatal(err)
			return false
		}
		config, err = png.DecodeConfig(temp)
		if err != nil {
			fmt.Println("png.DecodeConfig(temp)")
			return false
		}
	}
	/** Do proportional scaling */
	width := uint(base) /** Benchmark */
	height := uint(base * config.Height / config.Width)

	canvas := resize.Thumbnail(width, height, origin, resize.Lanczos3)
	file_out, err := os.Create(to)
	defer file_out.Close()
	if err != nil {
		log.Fatal(err)
		return false
	}
	if typeImage == 0 {
		err = png.Encode(file_out, canvas)
		if err != nil {
			fmt.Println("Failed to compress image")
			return false
		}
	} else {
		err = jpeg.Encode(file_out, canvas, &jpeg.Options{Quality: Quality})
		if err != nil {
			fmt.Println("Failed to compress image")
			return false
		}
	}

	return true
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
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
