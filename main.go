package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/joho/godotenv"
	"github.com/nexidian/gocliselect"
	"github.com/sashabaranov/go-openai"
)

type Matches struct {
	Index int
	Title string
	Link  string
}

type Batsman struct {
	Name  string
	Runs  string
	Balls string
}

type Bowler struct {
	Name    string
	Runs    string
	Wickets string
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAIRequest struct {
	Model       string                         `json:"model"`
	Messages    []openai.ChatCompletionMessage `json:"messages"`
	MaxTokens   int                            `json:"max_tokens"`
	Temperature float64                        `json:"temperature"`
}

type OpenAIResponse struct {
	Choices []openai.ChatCompletionChoice `json:"choices"`
}

func ParseTitle(title string) string {
	title = strings.Split(title, ",")[0]
	return title
}

func GetLiveScore(link string) (string, []Batsman, []Bowler) {
	score := ""
	currentBats := []Batsman{}
	flag := 0
	currentBowls := []Bowler{}
	c := colly.NewCollector()

	c.OnHTML("div", func(e *colly.HTMLElement) {
		if e.Attr("class") == "cb-col-100 cb-col cb-col-scores" || e.Attr("class") == "cb-col cb-col-100 cb-col-scores" {
			score = e.Text
		}
	})

	c.OnHTML("div.cb-min-inf.cb-col-100", func(e *colly.HTMLElement) {
		if flag == 0 {
			e.ForEach("div.cb-col.cb-col-100.cb-min-itm-rw", func(i int, e *colly.HTMLElement) {
				batsman := Batsman{
					Name:  e.ChildText("a"),
					Runs:  e.ChildText("div:nth-of-type(2)"),
					Balls: e.ChildText("div:nth-of-type(3)"),
				}
				currentBats = append(currentBats, batsman)
			})
			flag++
		} else {
			e.ForEach("div.cb-col.cb-col-100.cb-min-itm-rw", func(i int, e *colly.HTMLElement) {
				bowler := Bowler{
					Name:    e.ChildText("a"),
					Runs:    e.ChildText("div:nth-of-type(4)"),
					Wickets: e.ChildText("div:nth-of-type(5)"),
				}
				currentBowls = append(currentBowls, bowler)
			})
		}
	})

	c.Visit("https://www.cricbuzz.com" + link)
	return score, currentBats, currentBowls
}

func callOpenAI(request OpenAIRequest, apiKey string) (string, error) {
	client := openai.NewClient(apiKey)

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:     request.Model,
			Messages:  request.Messages,
			MaxTokens: request.MaxTokens,
		},
	)

	if err != nil {
		return "", err
	}

	return resp.Choices[0].Message.Content, nil
}

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY not found in .env file")
	}

	c := colly.NewCollector()
	matches := []Matches{}

	c.OnHTML("div", func(e *colly.HTMLElement) {
		if e.Attr("class") == "cb-col-100 cb-col cb-schdl cb-billing-plans-text" {
			match := Matches{
				Index: len(matches),
				Title: ParseTitle(e.Text),
				Link:  e.ChildAttr("a", "href"),
			}
			matches = append(matches, match)
		}
	})

	c.Visit("https://www.cricbuzz.com/cricket-match/live-scores")

	menu := gocliselect.NewMenu("Choose the match")

	for _, match := range matches {
		menu.AddItem(match.Title, match.Link)
	}

	choice := menu.Display()
	fmt.Println()

	var useOpenAI string
	fmt.Println("Do want to enable AI generated meta Comentary?: ")
	fmt.Scan(&useOpenAI)

	var mu sync.Mutex
	openAICallCounter := 0
	openAICallThreshold := 2
	openAICallResetTime := 1 * time.Minute

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	done := make(chan bool)

	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				score, currentBats, currentBowls := GetLiveScore(choice)
				fmt.Print("\033[2J")

				if strings.ToLower(useOpenAI) == "y" {

					mu.Lock()
					openAICallCounter++
					mu.Unlock()

					if openAICallCounter <= openAICallThreshold {
						// Call OpenAI API
						batsmenStr := ""
						for _, batsman := range currentBats {
							batsmenStr += fmt.Sprintf("%s (Runs: %s, Balls: %s), ", batsman.Name, batsman.Runs, batsman.Balls)
						}
						batsmenStr = strings.TrimSuffix(batsmenStr, ", ")

						bowlersStr := ""
						for _, bowler := range currentBowls {
							bowlersStr += fmt.Sprintf("%s (Runs: %s, Wickets: %s), ", bowler.Name, bowler.Runs, bowler.Wickets)
						}
						bowlersStr = strings.TrimSuffix(bowlersStr, ", ")

						// Construct the prompt
						input := fmt.Sprintf("You are a comedian, Based on the data given please summarize and predict the following cricket match data into a single sentence. Include the current score, batsmen performances (runs and balls), and bowlers performances (runs conceded and wickets):\nCurrent score: %s\nBatsmen: %s\nBowlers: %s", score, batsmenStr, bowlersStr)

						request := OpenAIRequest{
							Model: openai.GPT3Dot5Turbo,
							Messages: []openai.ChatCompletionMessage{
								{
									Role:    openai.ChatMessageRoleUser,
									Content: input,
								},
							},
							MaxTokens:   1024,
							Temperature: 0.8,
						}

						condensedScore, err := callOpenAI(request, apiKey)
						if err != nil {
							fmt.Println("Error: ", err)
							continue
						}

						fmt.Printf("\r%s\n", condensedScore)

					} else {
						fmt.Printf("\rScore %s\n", score)
					}
					// Reset the counter if necessary
					time.AfterFunc(openAICallResetTime, func() {
						mu.Lock()
						openAICallCounter = 0
						mu.Unlock()
					})

				} else {
					fmt.Printf("\rScore %s\n", score)
				}

			}
		}
	}()
	<-make(chan struct{})
}
