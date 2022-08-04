package main

import (
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"github.com/Aereum/aereum/core/crypto"
	"github.com/gobwas/ws/wsutil"
)

type NameToken struct {
	Name  string `json:"name"`
	Token string `json:"token"`
}

type OpenContentJSON struct {
	Author    NameToken `json:"author"`
	HTML      string    `json:"html"`
	TimeStamp time.Time `json:"timestamp"`
}

type OpenStageJSON struct {
	Onboarding string            `json:"onboardign"`
	Stage      NameToken         `json:"stage"`
	Owner      NameToken         `json:"owner"`
	Moderators []NameToken       `json:"moderators"`
	Submittors []NameToken       `json:"submittors"`
	Content    []OpenContentJSON `json:"content"`
}

func (t *Theatre) NameToken(token crypto.Token) NameToken {
	var nt NameToken
	if member, ok := t.members[token]; ok {
		nt.Name = member.Name
		nt.Token = token.Hex()
	}
	return nt
}

type OpenStage struct {
	Owner       crypto.Token
	Submittors  map[crypto.Token]struct{}
	Moderators  map[crypto.Token]struct{}
	Description string
	Content     []*OpenContent
}

func (s *OpenStage) Submittor(token crypto.Token) bool {
	s.Submittors[token] = struct{}{}
	return true
}

func (s *OpenStage) Moderator(token crypto.Token) bool {
	s.Moderators[token] = struct{}{}
	return true
}

func (t *Theatre) Publish(content *OpenContent, stageToken crypto.Token) bool {
	stage, ok := t.stages[stageToken]
	if !ok {
		return false
	}
	stageName := stage.Description
	stage.Content = append(stage.Content, content)
	actorName, _ := t.members[content.Author]
	if actorName == nil {
		return false
	}
	t.act <- StageAct{
		actor:     content.Author,
		actorName: actorName.Name,
		stage:     stageToken,
		stageName: stageName,
	}
	return true
}

type OpenContent struct {
	Author      crypto.Token
	ContentType string
	Content     []byte
	TimeStamp   time.Time
}

func (t *Theatre) HTML(content *OpenContent) string {
	html := ""
	if content.ContentType == "text" {
		if author, ok := t.members[content.Author]; ok {
			text := string(content.Content)
			rows := strings.Split(text, "\n")
			for _, row := range rows {
				html = fmt.Sprintf("%v<p>%v:%v</p>", html, author.Name, row)
			}

		}

	}
	return html
}

type Member struct {
	Name     string
	Redirect string
}

type StageAct struct {
	actor     crypto.Token
	actorName string
	stage     crypto.Token
	stageName string
}

type Theatre struct {
	stages  map[crypto.Token]*OpenStage
	members map[crypto.Token]*Member
	names   map[string]crypto.Token
	act     chan StageAct
}

func (t *Theatre) respondToken(ws net.Conn, tokenString string) {
	token := decodeToken(tokenString)
	stage, ok := t.stages[token]
	if ok {
		stagejson := OpenStageJSON{
			Stage:      NameToken{Name: stage.Description, Token: tokenString},
			Owner:      t.NameToken(stage.Owner),
			Moderators: make([]NameToken, 0),
			Submittors: make([]NameToken, 0),
			Content:    make([]OpenContentJSON, 0),
		}
		if stagejson.Owner.Name == "aereum-onboarding" {
			stagejson.Onboarding = "aereum-onboarding"
		}
		for moderator, _ := range stage.Moderators {
			stagejson.Moderators = append(stagejson.Moderators, t.NameToken(moderator))
		}
		for submittor, _ := range stage.Submittors {
			stagejson.Moderators = append(stagejson.Moderators, t.NameToken(submittor))
		}
		for _, post := range stage.Content {
			newContentJSON := OpenContentJSON{
				Author:    t.NameToken(post.Author),
				HTML:      string(post.Content),
				TimeStamp: post.TimeStamp,
			}
			stagejson.Content = append(stagejson.Content, newContentJSON)
		}
		if bytes, err := json.Marshal(stagejson); err == nil {
			wsutil.WriteServerText(ws, bytes)
		}
		return
	}
}

func decodeToken(hexToken string) crypto.Token {
	var token crypto.Token

	if bytes, err := hex.DecodeString(hexToken); err == nil {
		copy(token[:], bytes)
	}
	return token
}

func (t *Theatre) CreateOpenAudience(author crypto.Token, name, description string) (crypto.Token, bool) {
	pub, _ := crypto.RandomAsymetricKey()
	stage := OpenStage{
		Owner:       author,
		Submittors:  map[crypto.Token]struct{}{author: {}},
		Moderators:  map[crypto.Token]struct{}{author: {}},
		Description: description,
		Content:     make([]*OpenContent, 0),
	}
	t.stages[pub] = &stage
	return pub, true
}

func readPlays() *Theatre {
	theatre := &Theatre{
		members: make(map[crypto.Token]*Member),
		stages:  make(map[crypto.Token]*OpenStage),
		names:   make(map[string]crypto.Token),
		act:     make(chan StageAct),
	}
	file, err := os.Open("plays.csv")
	if err != nil {
		log.Fatalf("Could not open palys.csv")
		return nil
	}

	sheakespeare, _ := crypto.RandomAsymetricKey()
	theatre.names["Sheakespeare"] = sheakespeare
	theatre.members[sheakespeare] = &Member{Name: "Sheakespeare", Redirect: ""}

	playsCSV := csv.NewReader(file)
	playsCSV.Comma = '\t'
	playsCSV.FieldsPerRecord = 3
	plays := make(map[string][][2]string)
	stages := make(map[string]crypto.Token)
	for {
		if line, _ := playsCSV.Read(); line != nil {
			if _, ok := theatre.names[line[1]]; !ok {
				token, _ := crypto.RandomAsymetricKey()
				theatre.names[line[1]] = token
				theatre.members[token] = &Member{Name: line[1], Redirect: ""}
			}
			if _, ok := stages[line[0]]; !ok {
				stageToken, ok := theatre.CreateOpenAudience(sheakespeare, line[0], fmt.Sprintf("Play: %v", line[0]))
				if !ok {
					log.Fatal("could not stage new play.")
				}
				stages[line[0]] = stageToken
				plays[line[0]] = make([][2]string, 0)
			}
			if play, ok := plays[line[0]]; ok {
				plays[line[0]] = append(play, [2]string{line[1], line[2]})
			} else {
				log.Fatalf("could not fing stage for play %v", line[0])
			}
		} else {
			break
		}
	}
	go func() {
		countRow := make(map[string]int)
		for playName, _ := range plays {
			countRow[playName] = 0
		}
		for {
			for playName, text := range plays {
				row := countRow[playName]
				if row >= len(text) {
					row = 0
				}
				countRow[playName] = row + 1
				speach := text[row]
				author, authorExists := theatre.names[speach[0]]
				if !authorExists {
					log.Fatalf("could not find character %v", speach[0])
				}
				stageToken := stages[playName]
				newContent := &OpenContent{Author: author, ContentType: "text", Content: []byte(speach[1])}
				theatre.Publish(newContent, stageToken)
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()
	return theatre
}
