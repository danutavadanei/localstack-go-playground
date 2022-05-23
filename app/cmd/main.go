package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"
	"io"
	"localstack/internal/config"
	"localstack/internal/server"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

const (
	s3Bucket = "spike-test-bucket"
	s3Region = "us-west-2"
)

func downloadHandler(client *s3.S3) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		keys, _ := r.URL.Query()["key"]

		result, err := client.GetObject(&s3.GetObjectInput{
			Key:    aws.String(keys[0]),
			Bucket: aws.String(s3Bucket),
		})

		if err != nil {
			http.Error(w, fmt.Sprintf("Error getting file from s3 %s", err.Error()), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", keys[0]+".txt"))
		w.Header().Set("Cache-Control", "no-store")

		bytesWritten, copyErr := io.Copy(w, result.Body)

		if copyErr != nil {
			http.Error(w, fmt.Sprintf("Error copying file to the http response %s", copyErr.Error()), http.StatusInternalServerError)
			return
		}

		log.Printf("Download of \"%s\" complete. Wrote %s bytes", "my-file.csv", strconv.FormatInt(bytesWritten, 10))
	}
}

func uploadHandler(s3Uploader *s3manager.Uploader) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		key := uuid.NewString()

		s3Uploader.Upload(&s3manager.UploadInput{
			Body:   r.Body,
			Bucket: aws.String(s3Bucket),
			Key:    aws.String(key),
		})

		w.Write([]byte(key))
	}
}

func main() {
	v := viper.New()
	v.AutomaticEnv()
	sigChannel := make(chan os.Signal)
	signal.Notify(sigChannel, os.Interrupt, syscall.SIGTERM)

	cfg := config.NewAppConfig(v)
	awsCfg := aws.NewConfig().WithEndpoint(cfg.AWSConfig.Endpoint).WithS3ForcePathStyle(true).WithRegion(s3Region)

	sess := session.Must(session.NewSession(awsCfg))
	uploader := s3manager.NewUploader(sess)

	m := mux.NewRouter()
	m.HandleFunc("/download", downloadHandler(s3.New(sess)))
	m.HandleFunc("/upload", uploadHandler(uploader))
	m.HandleFunc("/s3/buckets", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("list"))
	})

	srvShutdown := make(chan bool)
	srv := server.StartHttpServer(cfg.HTTPServerConfig, m, srvShutdown)

	<-sigChannel
	go shutdown(srv)
	<-srvShutdown

}

func shutdown(server *http.Server) {
	ctxShutDown, _ := context.WithTimeout(context.Background(), 30)
	err := server.Shutdown(ctxShutDown)
	if err != nil {
		_ = fmt.Errorf("error shutting down server (%s): %v", server.Addr, err)
		err = server.Close()
		if err != nil {
			_ = fmt.Errorf("error closing server (%s): %v", server.Addr, err)
		}
	}
}