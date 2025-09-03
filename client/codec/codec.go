package codec

import (
	"fmt"

	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/model"
)

// Serializes ClientBet into a key/value string separated by |
func EncodeBet(bet model.ClientBet) string {
	date := bet.Birthdate.Format("2006-01-02") // format date to YYYY-MM-DD
	return fmt.Sprintf(
		"dni:%d|numero:%d|nombre:%s|apellido:%s|fecha:%s\n",
		bet.ID,
		bet.Number,
		bet.Name,
		bet.Lastname,
		date,
	)
}
