package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"path"
	"strconv"

	"github.com/PuerkitoBio/goquery"
	"github.com/nlopes/slack"
)

// settings of slack
const (
	qiitaURL   = "https://qiita.com/"
	webhookURL = ""
	userAgent  = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/80.0.3987.149 Safari/537.36"
)

// QiitaResponse struct
type QiitaResponse struct {
	Trend struct {
		Edges []Edge `json:"edges"`
	} `json:"trend"`
}

// Edge struct
type Edge struct {
	Node struct {
		LikesCount int    `json:"likesCount"`
		Title      string `json:"title"`
		UUID       string `json:"uuid"`
		Author     struct {
			ProfileImageURL string `json:"profileImageUrl"`
			URLName         string `json:"urlName"`
		} `json:"author"`
	} `json:"node"`
}

func getPage() {
	req, _ := http.NewRequest("GET", qiitaURL, nil)
	req.Header.Add("User-Agent", userAgent)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		log.Fatal(err)
	}

	value, exists := doc.Find(".p-home_main > div").Attr("data-hyperapp-props")
	if exists {
		var qiitaResponse QiitaResponse
		if err := json.Unmarshal([]byte(value), &qiitaResponse); err != nil {
			fmt.Println(err)
			return
		}
		send(qiitaResponse)
	}
}

func send(qiitaResponse QiitaResponse) {
	articles := qiitaResponse.Trend.Edges
	for _, article := range articles {
		postSlack(article)
	}
}

// postSlack function
func postSlack(article Edge) error {
	payload := &slack.WebhookMessage{
		Attachments: []slack.Attachment{
			{
				Blocks: []slack.Block{
					getArticleSectionBlock(article),
				},
			},
		},
	}

	err := slack.PostWebhook(webhookURL, payload)
	if err != nil {
		return err
	}

	return nil
}

// getArticleSectionBlock function
func getArticleSectionBlock(article Edge) slack.Block {
	articleURL, _ := url.Parse(qiitaURL)
	articleURL.Path = path.Join(
		articleURL.Path,
		article.Node.Author.URLName,
		"items",
		article.Node.UUID,
	)

	title := fmt.Sprintf(
		"<%s|%s>", articleURL.String(), article.Node.Title,
	)
	author := fmt.Sprintf(
		"Author: %s", article.Node.Author.URLName,
	)
	likesCount := fmt.Sprintf(
		"%s likes", strconv.Itoa(article.Node.LikesCount),
	)
	text := fmt.Sprintf("%s\n%s\n%s", title, author, likesCount)

	textBlockObject := slack.NewTextBlockObject(
		"mrkdwn",
		text,
		false,
		false,
	)
	imageElement := slack.NewImageBlockElement(
		article.Node.Author.ProfileImageURL,
		article.Node.Author.URLName,
	)
	section := slack.NewSectionBlock(
		textBlockObject,
		nil,
		slack.NewAccessory(imageElement),
	)
	return section
}

func main() {
	getPage()
}
