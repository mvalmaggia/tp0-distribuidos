package codec

import (
	"fmt"

	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/model"
)

// Serializes ClientBet into a key/value string separated by |
func EncodeBet(bet model.ClientBet) string {
	date := bet.Birthdate.Format("2006-01-02") // format date to YYYY-MM-DD
	return fmt.Sprintf(
		"agency:%s|document:%s|number:%d|first_name:%s|last_name:%s|birthdate:%s\n",
		bet.Agency,
		bet.ID,
		bet.Number,
		bet.Name,
		bet.Lastname,
		date,
	)
}
