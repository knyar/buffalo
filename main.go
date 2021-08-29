package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
)

var dryRun = false

const tplMessage = "We tried %s with buffalo sauce. Rating: 10/10"
const tplURL = "https://twitter.com/GoodWithBuffalo/status/%d"

func tweet(text string) (int64, error) {
	if dryRun {
		id := rand.Int63()
		log.Printf("Dry-run: returning fake tweet id %d", id)
		return id, nil
	}
	apiKey := os.Getenv("API_KEY")
	apiSecret := os.Getenv("API_SECRET")
	accessToken := os.Getenv("ACCESS_TOKEN")
	accessTokenSecret := os.Getenv("ACCESS_TOKEN_SECRET")
	config := oauth1.NewConfig(apiKey, apiSecret)
	token := oauth1.NewToken(accessToken, accessTokenSecret)
	httpClient := config.Client(oauth1.NoContext, token)
	client := twitter.NewClient(httpClient)
	tweet, _, err := client.Statuses.Update(text, nil)
	if err != nil {
		return 0, err
	}
	return tweet.ID, nil
}

func post(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("expected POST"))
		return
	}

	if err := r.ParseForm(); err != nil {
		log.Printf("error parsing: %v", err)
		fmt.Fprintf(w, "error parsing: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	food := strings.ToLower(r.FormValue("food"))
	food = strings.TrimSpace(food)
	log.Printf("Food: %s", food)

	if len(food) > 35 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("food name too long"))
		return
	}

	if food == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

	id, err := tweet(fmt.Sprintf(tplMessage, food))
	if err != nil {
		log.Printf("error happened: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("error happened: %v", err)))
		return
	}

	url := fmt.Sprintf(tplURL, id)

	http.Redirect(w, r, url, http.StatusSeeOther)
}

func main() {
	if dry := os.Getenv("DRY_RUN"); dry != "" {
		dryRun = true
	}

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)
	http.HandleFunc("/post", post)

	log.Println("Listening on :8000")
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal(err)
	}
}
