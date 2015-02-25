package main

import (
	"bufio"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"
)

func main() {
	// Be polite
	log.Println("-----| STREAM RECORDER |-----")

	// Load settings
	type ConfigStruct struct {
		Playlist              string        `json:"Playlist"`
		OutputFolderPath      string        `json:"OutputFolderPath"`
		OutputIntervalMinutes time.Duration `json:"OutputIntervalMinutes"`
		OutputLengthHours     time.Duration `json:"OutputLengthHours"`
		Timezone              string        `json:"Timezone"`
	}

	var config ConfigStruct

	configFile, _ := os.Open("./config.json")
	configDecoder := json.NewDecoder(configFile)

	if err := configDecoder.Decode(&config); err != nil {
		log.Println(err)
	}

	configFile.Close()

	log.Println("Playlist URL is:", config.Playlist)

	// Fetch the streaming URL from the playlist
	streamUrl := getStreamUrlFromPlaylist(config.Playlist)

	log.Println("Streaming URL found:", streamUrl)

	// Setting timer
	t0 := time.Now()

	// Create the output folder
	os.Mkdir(config.OutputFolderPath, 0777) // TO DO: Fix permission bits.

	log.Println("Output folder succesfully created.")

	outputFile := createNewOutputfile(t0, config.OutputFolderPath, config.Timezone)

	// Set up the buffers and streamings
	stream, _ := http.Get(streamUrl)
	defer stream.Body.Close()

	reader := bufio.NewReader(stream.Body)
	writer := bufio.NewWriter(outputFile)

	for {
		// Recordings are done in an regular basis
		if time.Since(t0) < time.Minute*config.OutputIntervalMinutes {
			saveStreamingBytes(reader, writer)
		} else {
			// When an hour has passed reset timer and writer.
			t0 = time.Now()
			writer = bufio.NewWriter(createNewOutputfile(t0, config.OutputFolderPath, config.Timezone))

			// Also delete recordings with more than 24 hours.
			eraseOldOutputs(config.OutputFolderPath, config.OutputLengthHours)
		}
	}

}

func eraseOldOutputs(folder string, length time.Duration) {
	outputFolder, _ := os.Open(folder)
	defer outputFolder.Close()

	files, _ := outputFolder.Readdir(0)

	for _, el := range files {
		if time.Since(el.ModTime()) > time.Hour*length {
			os.Remove(folder + "/" + el.Name())
		}
	}
}

func createNewOutputfile(t time.Time, folder string, tz string) *os.File {
	// Layout string: Mon Jan 2 15:04:05 -0700 MST 2006
	location, _ := time.LoadLocation(tz)
	t = t.In(location)

	fileName := folder + "/" + t.Format("20060102_150405") + ".mp3"
	outputFile, _ := os.Create(fileName)

	log.Println("File", fileName, "succesfully created.")

	return outputFile
}

func saveStreamingBytes(src *bufio.Reader, dst *bufio.Writer) {
	// Byte by byte streaming
	b, err := src.ReadByte()

	if err == nil {
		dst.WriteByte(b)
		dst.Flush()
	} else {
		log.Println(err.Error())
	}
}

func getStreamUrlFromPlaylist(url string) string {
	resp, err := http.Get(url)

	if err != nil {
		log.Println(err)
	}

	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	reachedUrl := false

	var streamUrl string

	for !reachedUrl && scanner.Scan() {
		current := scanner.Text()
		reachedUrl, _ := regexp.MatchString("^http://.*", current)

		if reachedUrl {
			streamUrl = current
		}
	}

	return streamUrl
}
