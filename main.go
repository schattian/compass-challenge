package main

import (
	"encoding/csv"
	"errors"
	"io"
	"log"
	"os"
	"strconv"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("you must provide an input filepath as argument, as in: `go run main.go contacts.csv`")
	}
	f, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	dedup, err := NewDeduplicator(f)
	if err != nil {
		log.Fatal(err)
	}
	err = dedup.CreateReport(os.Stdout, ACCURACY_NULL, true)
	if err != nil {
		log.Fatal(err)
	}
}

type Deduplicator struct {
	contacts map[int]*Contact
}

func NewDeduplicator(r io.Reader) (*Deduplicator, error) {
	reader := csv.NewReader(r)
	header, err := reader.Read()
	if err != nil {
		return nil, err
	}

	colIndex := make(map[string]int)
	for i, col := range header {
		colIndex[col] = i
	}

	contacts := make(map[int]*Contact)
	for i := 0; ; i++ {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		contact := &Contact{
			FirstName: record[colIndex["name"]],
			LastName:  record[colIndex["name1"]],
			Email:     record[colIndex["email"]],
			ZipCode:   record[colIndex["postalZip"]],
			Address:   record[colIndex["address"]],
		}
		contacts[i] = contact
	}

	return &Deduplicator{contacts: contacts}, nil
}

const (
	ACCURACY_HIGH  = 0.7
	ACCURACY_MED   = 0.4
	ACCURACY_LOW   = 0.1
	ACCURACY_NULL  = 0
	ACCURACY_NLOW  = -ACCURACY_LOW
	ACCURACY_NMED  = -ACCURACY_MED
	ACCURACY_NHIGH = -ACCURACY_HIGH

	LABEL_ACCURACY_HIGH  = "High"
	LABEL_ACCURACY_MED   = "Medium"
	LABEL_ACCURACY_LOW   = "Low"
	LABEL_ACCURACY_NULL  = "Zero"
	LABEL_ACCURACY_NLOW  = "Negative Low"
	LABEL_ACCURACY_NMED  = "Negative Medium"
	LABEL_ACCURACY_NHIGH = "Negative High"
)

func labelScore(score float64) string {
	switch {
	case score >= ACCURACY_HIGH:
		return LABEL_ACCURACY_HIGH
	case score >= ACCURACY_MED:
		return LABEL_ACCURACY_MED
	case score > ACCURACY_NULL:
		return LABEL_ACCURACY_LOW
	case score == ACCURACY_NULL:
		return LABEL_ACCURACY_NULL
	case score > ACCURACY_NMED:
		return LABEL_ACCURACY_NLOW
	case score >= ACCURACY_NMED:
		return LABEL_ACCURACY_NMED
	default:
		return LABEL_ACCURACY_NHIGH
	}
}

// CreateReport creates a report in csv format, scoring each contact as a potential duplicate with each other
// threshold indicates which scores to hide from the report (ie: score must be >= threshold to be reported)
// labelScores is used to output a label instead of the numeric value of the score
func (d *Deduplicator) CreateReport(w io.Writer, threshold float64, labelScores bool) error {
	_, err := w.Write([]byte("ContactID Source,ContactID Match,Accuracy\n"))
	if err != nil {
		return err
	}
	for idSource := range d.contacts {
		scores, err := d.score(idSource)
		if err != nil {
			return err
		}
		for idMatch, score := range scores {
			if score < threshold {
				continue
			}
			acc := strconv.FormatFloat(score, 'f', 2, 64)
			if labelScores {
				acc = labelScore(score)
			}
			_, err := w.Write([]byte(strconv.Itoa(idSource) + "," + strconv.Itoa(idMatch) + "," + acc + "\n"))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (d *Deduplicator) score(ID int) (map[int]float64, error) {
	c, ok := d.contacts[ID]
	if !ok {
		return nil, errors.New("contact not found")
	}
	score := make(map[int]float64)
	for id, contact := range d.contacts {
		if id <= ID { // to avoid duplicated comparisons. ie: from O(n^2) to O(n^2 / 2)
			continue
		}
		score[id] = c.getSimilarity(contact)
	}
	return score, nil
}

const (
	valMatch        = 1.0
	valPartialMatch = 0.75
	valMismatch     = -1.0
	valMatchUnknown = 0.0

	// arbitrary values that weigh the similarity scores
	// the ideal weight would be the proportion of people sharing the same attribute
	weightFirstName = 0.3
	weightLastName  = 0.7
	weightFullName  = 0.4

	weightZipCode     = 0.1
	weightAddress     = 0.9
	weightFullAddress = 0.6

	// the same person could use different email or have different addresses throughout its life
	// the following list defines minimum values for similarity scores among the attributes

	// we consider the amount of people changing name negligible
	minSimName = -1.0

	// same person could have a work, personal and an old email addresses
	minSimEmail = -0.2

	// one person could have a work address, a home address and potentially move
	// creating a third address in a relatively short period of time
	minSimAddress = -(1.0 / 3)
)

type Contact struct {
	FirstName string
	LastName  string
	Email     string
	ZipCode   string
	Address   string
}

func (c *Contact) getSimilarity(target *Contact) float64 {
	// we consider email the special case to directly mark as duplicate
	simEmail := max(minSimEmail, getSimilarity(c.Email, target.Email))
	if simEmail == valMatch {
		return valMatch
	}

	simFirstName := weightFirstName * max(minSimName, getNameSimilarity(c.FirstName, target.FirstName))
	simLastName := weightLastName * max(minSimName, getNameSimilarity(c.LastName, target.LastName))
	simFullName := weightFullName * (simFirstName + simLastName)
	if simFirstName < 0 || simLastName < 0 { // if any of the names doesnt match, then treat full name as full mismatch
		simFullName = weightFullName * (minSimName)
	}

	simZipCode := weightZipCode * max(minSimAddress, getSimilarity(c.ZipCode, target.ZipCode))
	simAddress := weightAddress * max(minSimAddress, getSimilarity(c.Address, target.Address))
	simFullAddress := weightFullAddress * (simZipCode + simAddress)
	return simFullName + simFullAddress + simEmail
}

func getSimilarity(s1, s2 string) float64 {
	if s1 == "" || s2 == "" {
		return valMatchUnknown
	}
	if s1 == s2 {
		return valMatch
	}
	return valMismatch
}

func getNameSimilarity(s1, s2 string) float64 {
	sim := getSimilarity(s1, s2)
	if sim == 0 {
		return sim
	}
	if len(s1) == 1 || len(s2) == 1 { // if any is an initial, just compare the initials
		if s1[0] == s2[0] {
			return valPartialMatch
		}
		return valMismatch
	}
	return sim
}
