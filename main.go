package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func main() {
	var (
		player   = flag.String("player", "/usr/bin/mpv", "full filesystem path to the prefered player")
		fileType = flag.String("filetype", "pls", "file type of the stream to play with the player")
	)

	flag.Parse()

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

	err = playStream(*chosenStream, *fileType, *player)
	if err != nil {
		fmt.Println("Error playing stream")
		fmt.Println(err)
		os.Exit(6)
	}
}

type stream struct {
	title string
	freq  string
	phone string

	location   string
	streamInfo string

	urls map[string]string
}

func printStreams(streams []stream) {
	for i, stream := range streams {
		fmt.Printf("%d -> %s\n", i, stream.title)
		fmt.Printf("     %-15s %s\n", stream.location, stream.freq)
		fmt.Printf("     %s\n\n", stream.phone)
	}
}

func playStream(stream stream, preferedFile string, player string) error {
	file, ok := stream.urls[preferedFile]
	if !ok {
		return fmt.Errorf("stream has no file with type: %s", preferedFile)
	}

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

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return []stream{}, err
	}

	streams := []stream{}

	doc.Find(".box").Each(func(i int, s *goquery.Selection) {
		streams = append(streams, parseStream(s))
	})

	return streams, nil
}

func parseStream(s *goquery.Selection) stream {
	streamInfo := s.Find("#aantalbezoekers").Text()
	streamInfos := strings.SplitN(streamInfo, " ", 2)
	streamInfos = trimAndExtend(streamInfos, 2)

	infoEl := s.Find(".streamtitle")
	info, err := infoEl.Html()
	if err != nil {
		info = infoEl.Text()
	}

	infos := strings.Split(info, "<br/>")
	infos = trimAndExtend(infos, 3)

	urls := parseURLs(s.Find(".images_online a"))

	return stream{
		strings.Replace(strings.TrimSpace(infos[1]), "&amp;", "&", -1),
		strings.TrimSpace(infos[0]),
		strings.TrimSpace(infos[2]),
		streamInfos[0],
		streamInfos[1],
		urls}
}

func trimAndExtend(list []string, length int) []string {
	output := make([]string, length)

	for i, item := range list {
		if i >= length {
			break
		}

		output[i] = strings.TrimSpace(item)
	}

	return output
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
