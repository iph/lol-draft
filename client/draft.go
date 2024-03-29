package draft

import (
	"github.com/iph/lol-bot/ols"
)

type DraftPlayer struct {
	Player      *ols.Player
	bid         int
	bidTeamName string
}
type Draft struct {
	Current  DraftPlayer
	Herd     []DraftPlayer
	Assigned []DraftPlayer
	paused   bool
}

func InitDraft(players ols.Players) Draft {
	herd := []DraftPlayer{}
	for _, player := range players {
		herd = append(herd, DraftPlayer{Player: &player})
	}
	var current DraftPlayer
	current, herd = herd[len(herd)-1], herd[:len(herd)-1]
	assigned := []DraftPlayer{}
	draft := Draft{
		Current:  current,
		Herd:     herd,
		Assigned: assigned,
	}

	return draft
}

func (d *Draft) Pause() {
	d.paused = true
}

func (d *Draft) IsCurrentlyDrafting() bool {
	return d.paused
}

func (d *Draft) ArePlayersLeft() bool {
	return len(d.Herd) > 0
}

func (d *Draft) NextPlayer() {
	d.Assigned = append(d.Assigned, d.Current)
	d.Current, d.Herd = d.Herd[len(d.Herd)-1], d.Herd[:len(d.Herd)-1]
}

func (d *Draft) Bid(amt int, team string) bool {
	if d.Current.bid < amt {
		return false
	}
	d.Current.bid, d.Current.bidTeamName = amt, team
	return true
}
