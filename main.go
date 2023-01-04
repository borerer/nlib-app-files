package main

import (
	"bytes"
	"encoding/base64"
	"io"
	"os"

	"github.com/borerer/nlib-app-files/file"
	nlib "github.com/borerer/nlib-go"
	"github.com/borerer/nlib-go/har"
)

var (
	minioClient    *file.MinioClient
	EncodingBase64 = "base64"
)

func toBase64(buf []byte) string {
	return base64.StdEncoding.EncodeToString(buf)
}

func fromBase64(s string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(s)
}

func get(req *nlib.FunctionIn) (*nlib.FunctionOut, error) {
	filename := har.GetQuery(req, "file")
	stat, err := minioClient.HeadFile(filename)
	if err != nil {
		return nil, err
	}
	reader, err := minioClient.GetFile(filename)
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	buf, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	b64Str := toBase64(buf)
	res := har.Text(b64Str)
	res.Headers = append(res.Headers, har.Header{Name: "Content-Type", Value: stat.ContentType})
	res.Content.Encoding = &EncodingBase64
	return res, nil
}

func put(req *nlib.FunctionIn) (*nlib.FunctionOut, error) {
	filename := har.GetQuery(req, "file")
	contentType := har.GetHeader(req, "Content-Type")
	buf, err := fromBase64(*req.PostData.Text)
	if err != nil {
		return nil, err
	}
	reader := bytes.NewReader(buf)
	_, err = minioClient.PutFile(filename, contentType, true, reader)
	if err != nil {
		return nil, err
	}
	return har.Text("ok"), nil
}

func main() {
	minioClient = file.NewMinioClient(&file.MinioConfig{
		Endpoint:  os.Getenv("NLIB_MINIO_ENDPOINT"),
		AccessKey: os.Getenv("NLIB_MINIO_ACCESS_KEY"),
		SecretKey: os.Getenv("NLIB_MINIO_SECRET_KEY"),
		UseSSL:    os.Getenv("NLIB_MINIO_USE_SSL") == "true",
		Bucket:    os.Getenv("NLIB_MINIO_BUCKET"),
	})
	nlib.Must(minioClient.Start())

	nlib.SetEndpoint(os.Getenv("NLIB_SERVER"))
	nlib.SetAppID("files")
	nlib.Must(nlib.Connect())

	nlib.RegisterFunction("get", get)
	nlib.RegisterFunction("put", put)
	nlib.Wait()
}
