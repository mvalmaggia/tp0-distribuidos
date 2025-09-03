package model

import "time"

type ClientBet struct {
	Agency	  string
	Number    int
	Name      string
	Lastname  string
	ID        int
	Birthdate time.Time
}