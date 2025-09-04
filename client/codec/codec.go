package codec

import (
	"fmt"
	"strings"

	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/model"
)

// Serializes ClientBet into a key/value string separated by |
func EncodeBet(bet model.ClientBet) string {
	date := bet.Birthdate.Format("2006-01-02") // format date to YYYY-MM-DD
	return fmt.Sprintf(
		"agency:%s|dni:%d|number:%d|first_name:%s|last_name:%s|birthdate:%s\n",
		bet.Agency,
		bet.ID,
		bet.Number,
		bet.Name,
		bet.Lastname,
		date,
	)
}

// EncodeBetBatch encodes a batch of bets to a single string
func EncodeBetBatch(bets []model.ClientBet) string {
    encodedBets := make([]string, len(bets))
    
    for i, bet := range bets {
        encodedBets[i] = EncodeBet(bet)
    }
    
    // Join all encoded bets with a semicolon separator
    return strings.Join(encodedBets, ";")
}

func DecodeWinners(encoded string) ([]int, error) {
    if strings.TrimSpace(encoded) == "" {
        return []int{}, nil
    }

    parts := strings.Split(encoded, ";")
    dnis := make([]int, len(parts))

    for i, part := range parts {
        fmt.Sscanf(strings.TrimSpace(part), "%d", &dnis[i])
    }
    return dnis, nil
}