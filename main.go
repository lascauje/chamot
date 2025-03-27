package main

import (
	"chamot/cmd/bubbleterm"
	"chamot/cmd/hackernews"
	"chamot/cmd/ollama"
	"chamot/config"
	"log"
)

func main() {
	cfg, ok := config.LoadCfg()
	if !ok {
		log.Fatalf("error loading config file")
	}

	hn := hackernews.NewHackerNews(cfg.HNUrlStory, cfg.HNUrlItem, cfg.HNUrlWebItem, cfg.HNNumStory)
	ol := ollama.NewOllama(cfg.OllamaURL, cfg.OllamaModel, cfg.OllamaNumCtx)
	bt := bubbleterm.NewBubbleTerm(hn, ol)

	if err := bt.Run(); err != nil {
		log.Fatalf("error running app")
	}
}
