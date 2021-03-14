package main

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func main() {
	streams, err := fetchStreams()
	if err != nil {
		fmt.Println("Error fetching streams")
		fmt.Println(err)
		os.Exit(5)
	}

	printStreams(streams)

	fmt.Print("Which would you like to play?\n> ")

	var chosenStream *stream

	for chosenStream == nil {
		var i int
		_, err := fmt.Scanf("%d\n", &i)
		if err != nil {
			fmt.Println(err)
			continue
		}

		if i < len(streams) && i >= 0 {
			chosenStream = &streams[i]
		}
	}

	err = playStream(*chosenStream, "pls", "/usr/bin/mpv")
	if err != nil {
		fmt.Println("Error playing stream")
		fmt.Println(err)
		os.Exit(6)
	}
}

type stream struct {
	title string
	info  []string
	urls  map[string]string
}

func printStreams(streams []stream) {
	for i, stream := range streams {
		fmt.Printf("%d -> %s\n", i, stream.title)
		for _, info := range stream.info {
			fmt.Printf("     %s\n", info)
		}
		fmt.Println("")
	}
}

func playStream(stream stream, preferedFile string, player string) error {
	file := stream.urls[preferedFile]

	cmd := exec.Command(player, file)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin

	err := cmd.Run()
	return err
}

func fetchStreams() ([]stream, error) {
	res, err := http.Get("https://geheimezender.com/streamdata-desktop.php")
	if err != nil {
		return []stream{}, err
	}

	defer res.Body.Close()
	if res.StatusCode != 200 {
		return []stream{}, fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return []stream{}, err
	}

	streams := []stream{}

	// Find the review items
	doc.Find(".box").Each(func(i int, s *goquery.Selection) {
		streams = append(streams, parseStream(s))
	})

	return streams, nil
}

func parseStream(s *goquery.Selection) stream {
	title := s.Find("#aantalbezoekers").Text()

	infoEl := s.Find(".streamtitle")
	info, err := infoEl.Html()
	if err != nil {
		info = infoEl.Text()
	}

	infos := []string{}
	for _, info := range strings.Split(info, "<br/>") {
		cleaned := strings.TrimSpace(info)
		if len(cleaned) == 0 {
			continue
		}

		infos = append(infos, cleaned)
	}

	urls := parseURLs(s.Find(".images_online a"))

	return stream{title, infos, urls}
}

func parseURLs(s *goquery.Selection) map[string]string {
	urls := map[string]string{}

	s.Each(func(_ int, s *goquery.Selection) {
		url, ok := s.Attr("href")
		if !ok {
			return
		}

		if !strings.HasPrefix(url, "http") {
			return
		}

		split := strings.Split(url, ".")
		extension := split[len(split)-1]

		urls[extension] = url
	})

	return urls
}
