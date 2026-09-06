package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
)

func downloadFile(url, filepath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func initDeps() {
	// Check yt-dlp
	info, err := os.Stat("yt-dlp")
	if os.IsNotExist(err) || (err == nil && info.Size() < 1000000) {
		fmt.Println("Downloading yt-dlp...")
		err := downloadFile("https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp", "yt-dlp")
		if err == nil {
			os.Chmod("yt-dlp", 0755)
			fmt.Println("yt-dlp downloaded.")
		}
	}
	
	// Check ffmpeg
	info2, err2 := os.Stat("ffmpeg")
	if os.IsNotExist(err2) || (err2 == nil && info2.Size() < 1000000) {
		fmt.Println("Downloading ffmpeg...")
		err := downloadFile("https://johnvansickle.com/ffmpeg/releases/ffmpeg-release-amd64-static.tar.xz", "ffmpeg.tar.xz")
		if err == nil {
			fmt.Println("Extracting ffmpeg...")
			cmd := exec.Command("tar", "-xf", "ffmpeg.tar.xz")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			errTar := cmd.Run()
			if errTar == nil {
				cmdMv := exec.Command("sh", "-c", "mv ffmpeg-*-amd64-static/ffmpeg ffmpeg && mv ffmpeg-*-amd64-static/ffprobe ffprobe && rm -rf ffmpeg-*-amd64-static ffmpeg.tar.xz")
				cmdMv.Run()
				os.Chmod("ffmpeg", 0755)
				os.Chmod("ffprobe", 0755)
				fmt.Println("ffmpeg downloaded and extracted.")
			}
		}
	}

	// Download webpmux
	info3, err3 := os.Stat("webpmux")
	if os.IsNotExist(err3) || (err3 == nil && info3.Size() < 100000) {
		fmt.Println("Downloading webpmux...")
		cmd := exec.Command("curl", "-L", "-o", "libwebp.tar.gz", "https://storage.googleapis.com/downloads.webmproject.org/releases/webp/libwebp-1.3.2-linux-x86-64.tar.gz")
		cmd.Run()
		exec.Command("tar", "-xzf", "libwebp.tar.gz").Run()
		exec.Command("cp", "libwebp-1.3.2-linux-x86-64/bin/webpmux", ".").Run()
		os.Remove("libwebp.tar.gz")
		os.RemoveAll("libwebp-1.3.2-linux-x86-64")
		os.Chmod("webpmux", 0755)
		fmt.Println("webpmux downloaded.")
	}
}
