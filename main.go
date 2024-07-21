package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/knyar/buffalo/store"

	"github.com/michimani/gotwi"
	"github.com/michimani/gotwi/tweet/managetweet"
	"github.com/michimani/gotwi/tweet/managetweet/types"
)

var dryRun = false

const tplMessage = "We tried %s with buffalo sauce. Rating: 10/10"
const tplURL = "https://x.com/GoodWithBuffalo/status/%d"

// tweet posts a message to twitter and returns tweet id.
func tweet(text string) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if dryRun {
		id := rand.Int63()
		log.Printf("Dry-run: returning fake tweet id %d for %s", id, text)
		return id, nil
	}
	client, err := gotwi.NewClient(&gotwi.NewClientInput{
		AuthenticationMethod: gotwi.AuthenMethodOAuth1UserContext,
		OAuthToken:           os.Getenv("GOTWI_ACCESS_TOKEN"),
		OAuthTokenSecret:     os.Getenv("GOTWI_ACCESS_TOKEN_SECRET"),
	})
	if err != nil {
		return 0, err
	}
	req := &types.CreateInput{Text: gotwi.String(text)}
	resp, err := managetweet.Create(ctx, client, req)
	if err != nil {
		return 0, err
	}

	id, err := strconv.ParseInt(gotwi.StringValue(resp.Data.ID), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("could not convert tweet id %q to integer: %w", gotwi.StringValue(resp.Data.ID), err)
	}

	log.Printf("Posted tweet id %d for %s", id, text)
	return id, nil
}

// Trim common prefixes while storing food names in the database.
// This will make sure that 'a cheeseburger' points to the same tweet as 'cheeseburger'.
func trimPrefix(s string) string {
	s = strings.TrimPrefix(s, "a ")
	s = strings.TrimPrefix(s, "the ")
	return s
}

type server struct {
	store *store.Store
}

func (s *server) post(w http.ResponseWriter, r *http.Request) {
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

	if len(food) > 35 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("food name too long"))
		return
	}

	if food == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

	item, err := s.store.Get(trimPrefix(food))
	if err != nil {
		log.Printf("error happened: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("error happened: %v", err)))
		return
	}

	if item == nil {
		id, err := tweet(fmt.Sprintf(tplMessage, food))
		if err != nil {
			log.Printf("error happened: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("error happened: %v", err)))
			return
		}

		item, err = s.store.Put(trimPrefix(food), id)
		if err != nil {
			log.Printf("error happened: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("error happened: %v", err)))
			return
		}
	}

	url := fmt.Sprintf(tplURL, item.ID)
	log.Printf("Tweet for '%s' is %s", food, url)
	http.Redirect(w, r, url, http.StatusSeeOther)
}

func main() {
	if dry := os.Getenv("DRY_RUN"); dry != "" {
		dryRun = true
	}

	store, err := store.New()
	if err != nil {
		log.Fatal(err)
	}

	s := &server{store: store}

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)
	http.HandleFunc("/post", s.post)

	log.Println("Listening on :8000")
	log.Fatal(http.ListenAndServe(":8000", nil))
}
