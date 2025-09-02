package model

import "time"

type ClientBet struct {
	Number    int
	Name      string
	Lastname  string
	ID        int
	Birthdate time.Time
}