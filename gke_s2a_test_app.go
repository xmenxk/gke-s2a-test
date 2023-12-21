package main

import (
	apiv3 "cloud.google.com/go/translate/apiv3"
	"cloud.google.com/go/translate/apiv3/translatepb"
	"context"
	"fmt"
	"google.golang.org/api/option"
	"log"
	"net/http"
	"os"
	"sync"
)

// translate grpc client
var translateGrpcClient *apiv3.TranslationClient

// translate http client
var translateHttpClient *apiv3.TranslationClient
var ctx context.Context

func main() {
	os.Setenv("EXPERIMENTAL_GOOGLE_API_USE_S2A", "true")
	os.Setenv("GRPC_GO_LOG_SEVERITY_LEVEL", "info")
	os.Setenv("GRPC_GO_LOG_VERBOSITY_LEVEL", "99")
	var err error

	ctx = context.Background()

	// translate grpc client with connection pool size = 1
	translateGrpcClient, err = apiv3.NewTranslationClient(ctx, option.WithGRPCConnectionPool(1))
	if err != nil {
		log.Fatalf("can not create translate grpc client: %v", err)
	}
	defer translateGrpcClient.Close()

	// translate http client
	translateHttpClient, err = apiv3.NewTranslationRESTClient(ctx)
	if err != nil {
		log.Fatalf("can not create translate http client: %v", err)
	}
	defer translateHttpClient.Close()

	http.HandleFunc("/", indexHandler)

	// concurrently translates sentences, using the grpc client
	http.HandleFunc("/translatehttp", translateHttpHandler)

	// concurrently translates sentences, using the http client
	http.HandleFunc("/translategrpc", translateGrpcHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	log.Printf("Listening on port %s", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	w.Write([]byte("Hello and Welcome!"))
	return
}

func translateGrpcHandler(w http.ResponseWriter, r *http.Request) {

	var wg sync.WaitGroup
	texts := []string{
		"s2a is awesome",
		"zatar is great",
		"authentication is important",
		"mtls is a must",
		"google cloud is better than aws",
		"google cloud is better than azure",
		"how are you?",
		"good morning",
		"summer is the best",
		"I love Sunnyvale",
	}
	wg.Add(len(texts))
	for i := 0; i < len(texts); i++ {
		go func(t string) {
			doTranslate(t, "zh", ctx, w, translateGrpcClient)
			wg.Done()
		}(texts[i])
	}
	wg.Wait()
}

func translateHttpHandler(w http.ResponseWriter, r *http.Request) {
	var wg sync.WaitGroup
	texts := []string{
		"s2a is awesome",
		"zatar is great",
		"authentication is important",
		"mtls is a must",
		"google cloud is better than aws",
		"google cloud is better than azure",
		"how are you?",
		"good morning",
		"summer is the best",
		"I love Sunnyvale",
	}
	wg.Add(len(texts))
	for i := 0; i < len(texts); i++ {
		go func(t string) {
			doTranslate(t, "ar", ctx, w, translateHttpClient)
			wg.Done()
		}(texts[i])
	}
	wg.Wait()
}

func doTranslate(text string, targetLangCode string, ctx context.Context, w http.ResponseWriter, c *apiv3.TranslationClient) {
	log.Printf("sending translate text request ...")
	req := &translatepb.TranslateTextRequest{
		Parent:             fmt.Sprintf("projects/%s/locations/%s", "xmenxk-gke-dev", "us-central1"),
		TargetLanguageCode: targetLangCode,
		Contents:           []string{text},
		// See https://pkg.go.dev/cloud.google.com/go/translate/apiv3/translatepb#TranslateTextRequest.
	}
	resp, err := c.TranslateText(ctx, req)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("Translate: %v", err)))
		return
	}
	for _, translation := range resp.GetTranslations() {
		fmt.Fprintf(w, "Translated text: %v\n", translation.GetTranslatedText())
	}
}
