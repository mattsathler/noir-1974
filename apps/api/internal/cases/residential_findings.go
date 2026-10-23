package cases

import (
	"encoding/json"
	"fmt"
)

type residentialFinding struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Kind    string `json:"kind"`
}

type residentialFindingList struct {
	Items []residentialFinding `json:"items"`
}

type residentialCaseFindings struct {
	Killer    []residentialFinding `json:"killer"`
	Victim    []residentialFinding `json:"victim"`
	Colleague []residentialFinding `json:"colleague"`
	Associate []residentialFinding `json:"associate"`
}

func readResidentialFindingList(path string) ([]residentialFinding, error) {
	data, err := caseSchema.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var list residentialFindingList
	if err := json.Unmarshal(data, &list); err != nil {
		return nil, fmt.Errorf("decode residential findings %s: %w", path, err)
	}
	return list.Items, nil
}

func readResidentialCaseFindings(archetype string) (residentialCaseFindings, error) {
	data, err := caseSchema.ReadFile("residential_search/" + archetype + ".json")
	if err != nil {
		return residentialCaseFindings{}, err
	}
	var findings residentialCaseFindings
	if err := json.Unmarshal(data, &findings); err != nil {
		return residentialCaseFindings{}, fmt.Errorf("decode residential case findings %s: %w", archetype, err)
	}
	return findings, nil
}
