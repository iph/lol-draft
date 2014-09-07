package ols

import (
	"encoding/json"
	"os"
)

type Player struct {
	Ign           string
	Id            string
	Name          string
	NormalizedIgn string
	Roles         []string
	Email         string
	Score         int
	Team          string
	Captain       bool
}

type Team struct {
	players []*Player
	wins    int
	losses  int
	name    string
}

// Sorting is a bit weird...
type Players []Player

func (p Players) Len() int {
	return len(p)
}

func (p Players) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (p Players) Less(i, j int) bool {
	return p[i].Score < p[j].Score
}

// Save players to a file.
func (p Players) save(file string) {

}

func NewPlayersFromJSONFile(file string) Players {
	json_league_data, err := os.Open(file)
	defer json_league_data.Close()

	if err != nil {
		panic("File problem:" + err.Error())
	}

	var json_blob Players
	json_reader := json.NewDecoder(json_league_data)
	json_reader.Decode(&json_blob)
	return json_blob
}

func (p *Players) Filter(filter func(Player) bool) Players {
	players := []Player{}
	for _, player := range *p {
		if filter(player) {
			players = append(players, player)
		}
	}

	return players
}
