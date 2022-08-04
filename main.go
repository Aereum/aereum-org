package main

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/Aereum/aereum/core/crypto"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

type OpenWebSockets struct {
	connections []net.Conn
	sync.Mutex
}

func (o *OpenWebSockets) Close(conn net.Conn) {
	o.Lock()
	defer o.Unlock()
	for n, open := range o.connections {
		if open == conn {
			o.connections = append(o.connections[0:n], o.connections[n+1:]...)
		}
	}
}

func (w *OpenWebSockets) Broadcast(json []byte) {
	for _, conn := range w.connections {
		wsutil.WriteServerText(conn, json)
	}
}

func main() {
	theatre := readPlays()
	if theatre == nil {
		log.Fatal("could not play sheakespeare")
	}

	webSockets := &OpenWebSockets{
		connections: make([]net.Conn, 0),
	}

	go contentStats(theatre.act, webSockets, 5)

	http.HandleFunc("/ws", http.HandlerFunc(serveWSAPI(theatre, webSockets)))
	http.Handle("/", http.FileServer(http.Dir("./static/")))
	err := http.ListenAndServe("127.0.0.1:7000", nil)
	if err != nil {
		log.Fatal(err)
	}

}

type tokenStat struct {
	token crypto.Token
	stat  int
}

type tokenStats []tokenStat

func (s tokenStats) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s tokenStats) Less(i, j int) bool {
	return s[i].stat > s[j].stat
}

func (s tokenStats) Len() int {
	return len(s)
}

func refreshStats(stages map[crypto.Token][]time.Time, N int) tokenStats {
	now := time.Now()
	ordered := make(tokenStats, 0)
	for token, hits := range stages {
		newStat := make([]time.Time, 0)
		for _, hit := range hits {
			if now.Sub(hit) < 60*time.Second {
				newStat = append(newStat, hit)
			}
		}
		if len(newStat) == 0 {
			delete(stages, token)
		} else {
			stages[token] = newStat
			ordered = append(ordered, tokenStat{token: token, stat: len(newStat)})
		}
	}
	sort.Sort(ordered)
	if len(ordered) >= N {
		return ordered[0:N]
	}
	return ordered
}

type Dispatcher interface {
	Broadcast(json []byte)
}

type Stats struct {
	Stages []NameToken
	People []NameToken
}

func contentStats(receiver chan StageAct, brodcaster Dispatcher, N int) {
	// events hit by stage and member
	stagesHit := make(map[crypto.Token][]time.Time)
	actorsHit := make(map[crypto.Token][]time.Time)
	// names for stages and members
	cast := make(map[crypto.Token]string)
	stages := make(map[crypto.Token]string)
	// stat refresh timmer
	refresh := time.After(time.Second)
	for {
		select {
		case newContent := <-receiver:
			now := time.Now()
			if stat, ok := stagesHit[newContent.stage]; ok {
				stagesHit[newContent.stage] = append(stat, now)
			} else {
				stagesHit[newContent.stage] = []time.Time{now}
			}
			if stat, ok := actorsHit[newContent.actor]; ok {
				actorsHit[newContent.actor] = append(stat, now)
			} else {
				actorsHit[newContent.actor] = []time.Time{now}
			}
			if _, ok := cast[newContent.actor]; !ok {
				cast[newContent.actor] = newContent.actorName
			}
			if _, ok := stages[newContent.stage]; !ok {
				stages[newContent.stage] = newContent.stageName
			}

		case <-refresh:
			stats := Stats{
				Stages: make([]NameToken, 0),
				People: make([]NameToken, 0),
			}
			statsStages := refreshStats(stagesHit, N)
			for _, stage := range statsStages {
				if name, ok := stages[stage.token]; ok {
					stats.Stages = append(stats.Stages, NameToken{Token: stage.token.Hex(), Name: name})
				}
			}
			statsPeople := refreshStats(actorsHit, N)
			for _, actor := range statsPeople {
				if name, ok := cast[actor.token]; ok {
					stats.People = append(stats.People, NameToken{Token: actor.token.Hex(), Name: name})
				}
			}
			if bytes, err := json.Marshal(stats); err == nil {
				brodcaster.Broadcast(bytes)
			}
			refresh = time.After(time.Second)
		}
	}
}

type GenericMessage struct {
	Token  string `json:"token"`
	Stage  string `json:"stage"`
	Member string `json:"member"`
}

func serveWSAPI(Theatre *Theatre, openWS *OpenWebSockets) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, _, _, err := ws.UpgradeHTTP(r, w)
		if err != nil {
			log.Fatal(err)
		}
		openWS.Lock()
		openWS.connections = append(openWS.connections, conn)
		openWS.Unlock()
		go func() {
			defer conn.Close()
			for {
				msg, code, err := wsutil.ReadClientData(conn)
				if err != nil {
					break
				}
				if code.IsData() {
					var genericMsg GenericMessage
					if err := json.Unmarshal(msg, &genericMsg); err != nil {
						log.Fatalf("json: %v, %v", err, string(msg))
					}
					if genericMsg.Token != "" {
						Theatre.respondToken(conn, genericMsg.Token)
					}
				}
			}
			openWS.Close(conn)
		}()
	}
}
