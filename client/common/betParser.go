package common

import (
    "encoding/csv"
    "fmt"
    "io"
    "os"
    "strconv"
    "time"

    "github.com/7574-sistemas-distribuidos/docker-compose-init/client/model"
)

type BetParser struct {
    clientID string
    file     *os.File
    reader   *csv.Reader
}

func NewBetParser(clientID string, filePath string) (*BetParser, error) {
    file, err := os.Open(filePath)
    if err != nil {
        return nil, fmt.Errorf("failed to open file %s | %v", filePath, err)
    }

    reader := csv.NewReader(file)
    reader.FieldsPerRecord = 0

    return &BetParser{
        clientID: clientID,
        file:     file,
        reader:   reader,
    }, nil
}

func (l *BetParser) NextBatch(size int) ([]model.ClientBet, error) {
    var bets []model.ClientBet

    for i := 0; i < size; i++ {
        record, err := l.reader.Read()
        if err != nil {
            if err == io.EOF {
                break
            }
            return nil, fmt.Errorf("failed to read from file | %v", err)
        }

        bet, err := parseBet(record, l.clientID)
        if err != nil {
            log.Warningf("Error parsing bet: %v", err)
            continue
        }

        bets = append(bets, bet)
    }

    return bets, nil
}

func (l *BetParser) Close() {
    _ = l.file.Close()
}

func parseBet(record []string, clientID string) (model.ClientBet, error) {
    // Format: FirstName LastName,LastName,ID,BirthDate,Number
    
    number, err := strconv.Atoi(record[4])
    if err != nil {
        return model.ClientBet{}, fmt.Errorf("invalid number: %w", err)
    }

    id, err := strconv.Atoi(record[2])
    if err != nil {
        return model.ClientBet{}, fmt.Errorf("invalid ID: %w", err)
    }

    birthdate, err := time.Parse("2006-01-02", record[3])
    if err != nil {
        return model.ClientBet{}, fmt.Errorf("invalid birthdate: %w", err)
    }

    return model.ClientBet{
        Agency:    clientID,
        Number:    number,
        Name:      record[0],
        Lastname:  record[1],
        ID:        id,        
        Birthdate: birthdate,
    }, nil
}