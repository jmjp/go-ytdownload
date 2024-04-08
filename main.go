package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/kkdai/youtube/v2"
	ffmpeg_go "github.com/u2takey/ffmpeg-go"
)

const path_save = "./outputs"

type JobQueue struct {
	Id      int
	VideoId string
}

func main() {

	//read links from file
	links := []string{}
	file, err := os.Open("links.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		links = append(links, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	var videos []JobQueue
	for i := 0; i < len(links); i++ {
		if len(links[i]) > 18 {
			links[i] = "https://www.youtube.com/watch?v=" + links[i][17:]
		}
		videos = append(videos, JobQueue{Id: i, VideoId: links[i]})
	}

	if _, err := os.Stat("./tmp"); os.IsNotExist(err) {
		os.Mkdir("./tmp", 0777)
	}

	if _, err := os.Stat(path_save); os.IsNotExist(err) {
		os.Mkdir(path_save, 0777)
	}

	jobs := make(chan int, len(videos))
	results := make(chan int, len(videos))

	for w := 1; w <= 2; w++ {
		go worker(videos, jobs, results)
	}

	for j := 0; j < len(videos); j++ {
		jobs <- j
	}

	close(jobs)

	for a := 1; a <= len(videos); a++ {
		<-results
	}

	close(results)

	fmt.Println("All jobs are done!")

	os.RemoveAll("./tmp")

	fmt.Println("Done!")
}

func worker(queue []JobQueue, jobs <-chan int, results chan<- int) {
	client := youtube.Client{}
	for job := range jobs {
		video, err := client.GetVideo(queue[job].VideoId)
		if err != nil {
			panic(err)
		}
		downloadVideo(client, video)
		convertToMp3("./tmp/"+video.Title, video.Title)
		results <- job
	}
}

func downloadVideo(client youtube.Client, video *youtube.Video) {
	path := "./tmp/" + video.Title
	fmt.Printf("Downloading %s!\n", video.Title)
	formats := video.Formats.WithAudioChannels()
	stream, _, err := client.GetStream(video, &formats[0])
	if err != nil {
		// handle error
		panic(err)
	}

	file, err := os.Create(path + ".mp4")
	if err != nil {
		panic(err)
	}
	_, err = io.Copy(file, stream)
	if err != nil {
		panic(err)
	}
	stream.Close()
	file.Close()
}

func convertToMp3(path string, title string) {
	ffmpeg_go.Input(path + ".mp4").Output("./outputs/" + title + ".mp3").ErrorToStdOut().Run()
}
