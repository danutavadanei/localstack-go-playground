package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
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

func main() {
	v := viper.New()
	v.AutomaticEnv()
	sigChannel := make(chan os.Signal)
	signal.Notify(sigChannel, os.Interrupt, syscall.SIGTERM)

	cfg := config.NewAppConfig(v)

	sess := session.Must(session.NewSession(cfg.AWSConfig))
	s3uploader := s3manager.NewUploader(sess)
	s3client := s3.New(sess)

	m := mux.NewRouter()

	m.HandleFunc("/s3/buckets", func(w http.ResponseWriter, r *http.Request) {
		req, resp := s3client.ListBucketsRequest(&s3.ListBucketsInput{})

		err := req.Send()

		if err != nil {
			log.Printf("error listing s3 buckets: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		bytes, err := json.Marshal(resp)

		if err != nil {
			log.Printf("error encoding s3 response: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(bytes)
	}).Methods("GET").Name("listBuckets")

	m.HandleFunc("/s3/buckets/{bucket}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		bucket := vars["bucket"]

		resp, err := s3client.ListObjects(&s3.ListObjectsInput{Bucket: &bucket})

		if err != nil {
			log.Printf("error listing s3 bucket (%s): %v", bucket, err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		bytes, err := json.Marshal(resp)

		if err != nil {
			log.Printf("error encoding s3 response: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(bytes)
	}).Methods("GET").Name("listBucketObjects")

	m.HandleFunc("/s3/buckets/{bucket}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		bucket := vars["bucket"]

		err := r.ParseMultipartForm(32 << 20) // maxMemory 32MB

		if err != nil {
			log.Printf("error parsing request: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		f, h, err := r.FormFile("file")

		if err != nil {
			log.Printf("error parsing request: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		o, err := s3uploader.Upload(&s3manager.UploadInput{
			Body:   f,
			Bucket: aws.String(bucket),
			Key:    aws.String(h.Filename),
		})

		if err != nil {
			log.Printf("error uploading file to s3 bucket(%s): %v", bucket, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		bytes, err := json.Marshal(o)

		if err != nil {
			log.Printf("error encoding s3 response: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		_, _ = w.Write(bytes)
	}).Methods("POST", "PUT").Name("putBucketObject")

	m.HandleFunc("/s3/buckets/{bucket}/{key}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		bucket, key := vars["bucket"], vars["key"]

		o, err := s3client.GetObject(&s3.GetObjectInput{
			Key:    aws.String(key),
			Bucket: aws.String(bucket),
		})

		if err != nil {
			log.Printf("error getting file with key: %s from s3 bucket (%s): %v", key, bucket, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", key))
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Content-Type", *o.ContentType)

		bytesWritten, err := io.Copy(w, o.Body)

		if err != nil {
			log.Printf("error copying file with key: %s to the http response from s3 bucket (%s): %v", key, bucket, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		log.Printf("Download of \"%s\" complete. Wrote %s bytes", key, strconv.FormatInt(bytesWritten, 10))
	}).Methods("GET").Name("getBucketObject")

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
		log.Printf("error shutting down server (%s): %v", server.Addr, err)
		err = server.Close()
		if err != nil {
			log.Printf("error closing server (%s): %v", server.Addr, err)
		}
	}
}
