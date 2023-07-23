package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/log"
	"net/http"
	"time"
)

func index(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Hello world info")
	log.Warn().Err(errors.New("testaja")).Msg("test aja kok")

	log.Warn().Err(errors.New("testaja")).Msg("test aja dong")
	loc, _ := time.LoadLocation("Asia/Jakarta")
	t := time.Now().In(loc)
	hour := t.Format("15:04")
	type Test struct {
		Test string `json:"test"`
	}

	test123 := TestLife{}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(test123)
	fmt.Println(hour)
}

type TestLife struct {
	TestAja string `json:"test_aja,omitempty"`
}

func check(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "<h1>Health check</h1>")
}

func main() {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)

	//compressor := middleware.NewCompressor(5, "/*")
	//compressor.SetEncoder("br", func(w io.Writer, level int) io.Writer {
	//	params := enc.NewBrotliParams()
	//	params.SetQuality(level)
	//	return enc.NewBrotliWriter(params, w)
	//})
	//r.Use(compressor.Handler)
	r.Use(middleware.Compress(5))
	r.Get("/test", index)
	r.Get("/health_check", check)

	fmt.Println("Server starting...")
	http.ListenAndServe(":8080", r)
}
