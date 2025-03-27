package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Cfg struct {
	HNUrlStory   string
	HNUrlItem    string
	HNUrlWebItem string
	HNNumStory   int
	OllamaURL    string
	OllamaModel  string
	OllamaNumCtx int
}

func LoadCfg() (*Cfg, bool) {
	var ok bool

	var err error

	cfg := &Cfg{
		HNUrlStory:   "",
		HNUrlItem:    "",
		HNUrlWebItem: "",
		HNNumStory:   0,
		OllamaURL:    "",
		OllamaModel:  "",
		OllamaNumCtx: 0,
	}

	if err := godotenv.Load(); err != nil {
		return nil, false
	}

	cfg.HNUrlStory, ok = os.LookupEnv("HN_URL_STORY")
	if !ok {
		return nil, false
	}

	cfg.HNUrlItem, ok = os.LookupEnv("HN_URL_ITEM")
	if !ok {
		return nil, false
	}

	cfg.HNUrlWebItem, ok = os.LookupEnv("HN_URL_WEB_ITEM")
	if !ok {
		return nil, false
	}

	hnNumStory, ok := os.LookupEnv("HN_NUM_STORY")
	if !ok {
		return nil, false
	}

	cfg.HNNumStory, err = strconv.Atoi(hnNumStory)
	if err != nil {
		return nil, false
	}

	cfg.OllamaURL, ok = os.LookupEnv("OLLAMA_URL")
	if !ok {
		return nil, false
	}

	cfg.OllamaModel, ok = os.LookupEnv("OLLAMA_MODEL")
	if !ok {
		return nil, false
	}

	ollamaNumCtx, ok := os.LookupEnv("OLLAMA_NUM_CTX")
	if !ok {
		return nil, false
	}

	cfg.OllamaNumCtx, err = strconv.Atoi(ollamaNumCtx)
	if err != nil {
		return nil, false
	}

	return cfg, true
}
