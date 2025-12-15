package utils

import "strings"

func ProfaneFilter(body string) string {
	badWords := map[string]bool{
		"kerfuffle": true,
		"Kerfuffle": true,
		"sharbert":  true,
		"Sharbert":  true,
		"fornax":    true,
		"Fornax":    true,
	}

	bdy := strings.Split(body, " ")
	for idx, b := range bdy {
		badWord := badWords[b]
		if badWord {
			bdy[idx] = "****"
		}
	}

	return strings.Join(bdy, " ")
}
