package main

import (
	"code.google.com/p/go.net/websocket"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"github.com/iph/lol-draft/draft"
	"github.com/iph/lol-draft/net"
	"github.com/iph/lol-draft/ols"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

var OlsPlayerFile string = "json/ols_players.json"
var BiddersFile string = "json/bidders.json"

type Bidder struct {
	Player          ols.Player
	RemainingPoints int
	Token           string
}

type Bidders []Bidder

type Application struct {
	Draft   draft.Draft
	Bidders Bidders
	Admin   Bidder
	Server  *net.Server
	Ticker  *time.Ticker
}

// TODO: Inefficient. Already done somewhere else..
func BiddersFromJSONFile(file string) Bidders {
	json_league_data, err := os.Open(file)
	defer json_league_data.Close()

	if err != nil {
		panic("File problem:" + err.Error())
	}

	var json_blob Bidders
	json_reader := json.NewDecoder(json_league_data)
	json_reader.Decode(&json_blob)
	return json_blob

}

// Should only be called once to initialize the
func BidderInit() {
	players := ols.NewPlayersFromJSONFile(OlsPlayerFile)
	captain_filter := func(player ols.Player) bool { return player.Captain }
	captains := players.Filter(captain_filter)
	bidders := []Bidder{}
	for _, captain := range captains {
		bidder := Bidder{
			Player:          captain,
			RemainingPoints: 100,
			Token:           GenerateSHA512Hash(time.Now().String()),
		}
		bidders = append(bidders, bidder)
	}
	var bidz Bidders
	bidz = bidders

	file, _ := os.Create(BiddersFile)
	defer file.Close()

	data, _ := json.MarshalIndent(bidz, "", "    ")
	file.Write(data)

}

func GenerateSHA512Hash(str string) (returnStr string) {
	h := sha512.New()
	h.Write([]byte(str))
	returnStr = hex.EncodeToString(h.Sum(nil))
	return returnStr
}

func initApplication() Application {
	players := ols.NewPlayersFromJSONFile(OlsPlayerFile)

	app := Application{
		Draft:   draft.InitDraft(players),
		Bidders: BiddersFromJSONFile(BiddersFile),
		Server:  net.NewServer(),
	}
	app.Server.Receive = ApplicationReceiveHandler(&app)
	app.Server.NewConnection = ApplicationNewConnectionHandler(&app)
	return app
}

func (a *Application) HandleMessage(message net.Message) {
	log.Printf("Bid received: %v\n", message)
	if a.Draft.IsCurrentlyDrafting() {
		// Get player.
		log.Printf("Getting bidder...\n")
		bidder := a.GetBidder(message.Token)
		if bidder == nil {
			log.Printf("Bid invalid, wrong token\n")
			return
		}

		amt, err := strconv.Atoi(message.Payload.(string))

		if err != nil {
			log.Printf("Error converting: %s\n", err.Error())
			return
		}
		validPoints := bidder.RemainingPoints - amt
		if a.Draft.IsCurrentlyDrafting() && validPoints >= 0 && a.Draft.Bid(amt, bidder.Player.Team) {
			log.Printf("Bid accepted. Broadcasting..\n")
			payload := net.BidUpdatePayload{Bid: strconv.Itoa(amt), Team: bidder.Player.Team}
			message := net.Message{Type: "BID_UPDATE", Payload: payload}
			a.Server.Broadcast(message, nil)
		}
	}

}

func (a *Application) StartBid() {
	a.Draft.Start()
	go func() {
		secondsGone := 0
		currentBidTeam := a.Draft.Current.BidTeamName

		ticker := time.NewTicker(time.Second)
		for now := range ticker.C {
			var _ = now // Wow really?
			updatedBidTeam := a.Draft.Current.BidTeamName
			log.Printf("team: %s\n", updatedBidTeam)
			if updatedBidTeam != "" && updatedBidTeam == currentBidTeam {
				secondsGone += 1
			} else {
				secondsGone = 0
			}
			currentBidTeam = updatedBidTeam
			payload := net.Message{Type: "COUNTDOWN", Payload: strconv.Itoa(secondsGone)}
			a.Server.Broadcast(payload, nil)
			// Done
			if secondsGone == 10 {
				ticker.Stop()
				a.Draft.Pause()
				a.UpdateBidderWin()
				break
			}
		}
	}()
}

func (a *Application) UpdateBidderWin() {
	winnerTeam := a.Draft.Current.BidTeamName
	// find winner team
	var winnerBidder Bidder
	for _, bidder := range a.Bidders {
		if bidder.Player.Team == winnerTeam {
			winnerBidder = bidder
			break
		}
	}

	log.Printf("Winner: %s\n", winnerBidder.Player.Ign)
	winnerBidder.RemainingPoints -= a.Draft.Current.BidAmt
	payload := net.Message{Type: "WINNER", Payload: net.BidderInfoPayload{a.Draft.Current.BidAmt, winnerTeam}}
	a.Server.Broadcast(payload, nil)

}

func (a *Application) GetBidder(token string) *Bidder {
	for _, bidder := range a.Bidders {
		if bidder.Token == token {
			return &bidder
		}
	}
	return nil
}

func (a *Application) Run() {
	a.Server.Run()

}

func (a *Application) SendBidderInfo(packet *net.Packet) {

	bidderToken := packet.Message.Token
	bidder := a.GetBidder(bidderToken)
	a.Server.Send(net.Message{
		Type: "BIDDER",
		Payload: net.BidderInfoPayload{
			Points: bidder.RemainingPoints,
			Team:   bidder.Player.Team,
		},
	}, packet.From)
}

func ApplicationReceiveHandler(app *Application) func(*net.Packet) {
	return func(packet *net.Packet) {
		if packet.Message.Type == "BID" {
			app.HandleMessage(packet.Message)
		}

		if packet.Message.Type == "START" {
			app.StartBid()
		}

		if packet.Message.Type == "BIDDER" {
			app.SendBidderInfo(packet)
		}

		if packet.Message.Type == "NEXT" {
			if app.Draft.ArePlayersLeft() {
				app.Draft.NextPlayer()
				log.Printf("Next player: %v", app.Draft.Current)
				log.Printf("Assigned: %v\n", app.Draft.Assigned)
				log.Printf("Herd: %v\n", app.Draft.Herd)
				message := app.GetPlayerInfo()
				app.Server.Broadcast(message, nil)
			} else {
				log.Printf("all done")
			}
		}

	}
}

func ApplicationNewConnectionHandler(app *Application) func(*net.Connection) {
	return func(conn *net.Connection) {
		message := app.GetPlayerInfo()
		app.Server.Send(message, conn)

	}
}

func (a *Application) GetPlayerInfo() net.Message {
	message := net.Message{
		Type:    "PLAYER_INFO",
		Payload: net.PlayerPayload{a.Draft.Current.Player},
	}
	return message
}

func serverHandleWebSockets(server *net.Server) func(*websocket.Conn) {
	return func(websock *websocket.Conn) {
		connection := server.AddConnection(websock)
		// Listen on the socket forever..
		connection.Spin(server)
	}
}

func main() {
	//BidderInit()
	app := initApplication()
	app.Run()
	http.Handle("/draft/", websocket.Handler(serverHandleWebSockets(app.Server)))

	err := http.ListenAndServe(":9001", nil)

	if err != nil {
		panic(err)
	}

}

// Acceptable message {"Type":"BID", "Payload":"10", "Token":"nvrfgt"}
