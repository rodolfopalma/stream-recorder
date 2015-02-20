package main

import (
	"bufio"
	"log"
	"net/http"
	"os"
	"regexp"
)

const (
	PlaylistURL = "http://www.radioreference.com/scripts/playlists/1/13315/0-5442179616.m3u"
	FilePath    = "./recording.mp3"
)

func main() {
	// Fetch the streaming URL from the playlist
	streamUrl := getStreamUrlFromPlaylist(PlaylistURL)

	log.Println("Streaming URL:", streamUrl)

	// Create the output file
	recording, _ := os.Create(FilePath)
	defer recording.Close()

	log.Println("Output file succesfully created.")

	// Read byte by byte and write the output file
	stream, _ := http.Get(streamUrl)
	reader := bufio.NewReader(stream.Body)
	writer := bufio.NewWriter(recording)

	saveStreamingBytes(reader, writer)

}

func saveStreamingBytes(src *bufio.Reader, dst *bufio.Writer) {
	for {
		b, err := src.ReadByte()

		if err == nil {
			dst.WriteByte(b)
			dst.Flush()
		} else {
			log.Println(err.Error())
		}
	}
}

func getStreamUrlFromPlaylist(url string) string {
	resp, _ := http.Get(url)
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
