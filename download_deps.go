package main

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
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

func extractZip(zipFile, destDir string) error {
	r, err := zip.OpenReader(zipFile)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		path := filepath.Join(destDir, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(path, os.ModePerm)
			continue
		}
		
		os.MkdirAll(filepath.Dir(path), os.ModePerm)
		out, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}
		
		rc, err := f.Open()
		if err != nil {
			out.Close()
			return err
		}
		
		_, err = io.Copy(out, rc)
		out.Close()
		rc.Close()
		if err != nil {
			return err
		}
	}
	return nil
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
		err := downloadFile("https://github.com/vot/ffbinaries-precompiled/releases/download/v4.4.1/ffmpeg-4.4.1-linux-64.zip", "ffmpeg.zip")
		if err == nil {
			extractZip("ffmpeg.zip", ".")
			os.Remove("ffmpeg.zip")
			os.Chmod("ffmpeg", 0755)
			fmt.Println("ffmpeg downloaded.")
		}
	}
}
