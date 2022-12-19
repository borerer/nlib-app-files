package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"os"

	"github.com/borerer/nlib-app-files/file"
	nlibgo "github.com/borerer/nlib-go"
)

var (
	minioClient *file.MinioClient
)

func mustString(in map[string]interface{}, key string) (string, error) {
	raw, ok := in[key]
	if !ok {
		return "", fmt.Errorf("missing %s", key)
	}
	str, ok := raw.(string)
	if !ok {
		return "", fmt.Errorf("invalid type %s", key)
	}
	return str, nil
}

func get(in map[string]interface{}) interface{} {
	filename, err := mustString(in, "filename")
	if err != nil {
		return err.Error()
	}
	reader, err := minioClient.GetFile(filename)
	if err != nil {
		return err.Error()
	}
	defer reader.Close()
	buf, err := io.ReadAll(reader)
	if err != nil {
		return err.Error()
	}
	return buf
}

func put(in map[string]interface{}) interface{} {
	filename, err := mustString(in, "filename")
	if err != nil {
		return err.Error()
	}
	content, err := mustString(in, "content")
	if err != nil {
		return err.Error()
	}
	buf, err := base64.StdEncoding.DecodeString(content)
	if err != nil {
		return err.Error()
	}
	reader := bytes.NewReader(buf)
	_, err = minioClient.PutFile(filename, true, reader)
	if err != nil {
		return err.Error()
	}
	return "ok"
}

func wait() {
	ch := make(chan bool)
	<-ch
}

func main() {
	minioClient = file.NewMinioClient(&file.MinioConfig{
		Endpoint:  os.Getenv("NLIB_MINIO_ENDPOINT"),
		AccessKey: os.Getenv("NLIB_MINIO_ACCESS_KEY"),
		SecretKey: os.Getenv("NLIB_MINIO_SECRET_KEY"),
		UseSSL:    os.Getenv("NLIB_MINIO_USE_SSL") == "true",
		Bucket:    os.Getenv("NLIB_MINIO_BUCKET"),
	})
	if err := minioClient.Start(); err != nil {
		println(err.Error())
		return
	}
	nlib := nlibgo.NewClient(os.Getenv("NLIB_SERVER"), "files")
	nlib.RegisterFunction("get", get)
	nlib.RegisterFunction("put", put)
	wait()
}
