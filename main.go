package main

import (
	"fmt"
	"github.com/antzucaro/matchr"
	"github.com/tealeg/xlsx"
	"github.com/xrash/smetrics"
	"os"
	"sort"
	"strconv"
	"strings"
)

func main() {
	args := os.Args[1:]
	xlsfile1, err := xlsx.OpenFile(args[0])
	if err != nil {
		panic(err)
	}
	xlsfile2, err := xlsx.OpenFile(args[2])
	if err != nil {
		panic(err)
	}

	var threshHold float64
	if len(args) >= 5 {
		threshHold, err = strconv.ParseFloat(args[4], 32)
		if err != nil {
			panic(err)
		}
	} else {
		threshHold = 0.0
	}

	file1Column, err := strconv.Atoi(args[1])
	if err != nil {
		panic(err)
	}
	file2Column, err := strconv.Atoi(args[3])
	if err != nil {
		panic(err)
	}
	dict1 := makeNameDictionary(xlsfile1, file1Column)
	dict2 := makeNameDictionary(xlsfile2, file2Column)

	matches := make([]Match, 0)
	for _, name := range dict1 {
		for _, otherName := range dict2 {
			name1 := strings.Join(name.NameParts, " ")
			name2 := strings.Join(otherName.NameParts, " ")
			level := smetrics.JaroWinkler(name1, name2, 0.7, 4)
			score := matchr.DamerauLevenshtein(name1, name2)
			smith := matchr.SmithWaterman(name1, name2)
			initialsMatchScore := initialMatch(name.Initials, otherName.Initials)
			match := Match{Name1: name, Name2: otherName, Score: []float64{level, float64(score), float64(initialsMatchScore), float64(smith)}}
			matches = append(matches, match)
		}
	}
	sort.Sort(Matches(matches))
	var perfectMatches []Name = make([]Name, 0)
	for _, match := range matches {
		if match.Score[0] > 0.999 && match.Score[1] < 1 && match.Score[2] < 1 {
			perfectMatches = append(perfectMatches, match.Name1)
		}
	}
	for _, match := range matches {
		if match.Score[0] < threshHold && match.Score[1] > 0 {
			continue
		}
		if match.Score[1] > 12 && match.Score[0] < 0.7 {
			continue
		}
		if match.Score[2] > 0.0 {
			continue
		}

		if match.Score[3] < 6 {
			continue
		}
		isPerfect := false
		for _, p := range perfectMatches {
			if (match.Name1.Original == p.Original || match.Name2.Original == p.Original) && (match.Name2.Original != match.Name1.Original) {
				isPerfect = true
			}
		}
		if isPerfect {
			continue
		}
		//		fmt.Printf("[%s] %s <==> [%s] %s \n", match.Name1.Initials, match.Name1.Original, match.Name2.Initials, match.Name2.Original)
		fmt.Printf("[%f] %f - [%s] %s <==> [%s] %s - [%f]\n", match.Score[3], match.Score[0], match.Name1.Initials, match.Name1.Original, match.Name2.Initials, match.Name2.Original, match.Score[1])
		//		fmt.Printf("%s\n%s\n", match.Name1.Actual, match.Name2.Actual)
	}
}

func initialMatch(s1, s2 string) int {
	bigger, smaller := s1, s2
	if len(s1) < len(s2) {
		bigger, smaller = s2, s1
	}
	lastIndex := -1
	for _, char := range smaller {
		found := strings.IndexRune(string(bigger[lastIndex+1:]), char)
		if found < 0 {
			return 1
		}
		lastIndex = found
	}
	return 0
}

type Match struct {
	Name1 Name
	Name2 Name
	Score []float64
}

type Matches []Match

func (m Matches) Len() int {
	return len(m)
}

func (m Matches) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}

func (m Matches) Less(i, j int) bool {
	return m[i].Score[0] > m[j].Score[0]
}

type Matcher interface {
	Match(s1, s2 string) int
}

type Name struct {
	Actual    string
	Original  string
	NameParts []string
	Initials  string
}
type NameDict []Name

func (a Name) Len() int {
	return len(a.NameParts)
}

func (a NameDict) Len() int {
	return len(a)
}

func (a NameDict) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a NameDict) Less(i, j int) bool {
	return a[i].NameParts[0][0] < a[j].NameParts[0][0]
}

func makeNameDictionary(file *xlsx.File, colNumber int) []Name {
	dict1 := make([]Name, 0)
	sheet1 := file.Sheets[0]
	for _, row := range sheet1.Rows {
		for i, cell := range row.Cells {
			if uint(i) != uint(colNumber) {
				//			if i != colNumber {
				continue
			}
			name := strings.Trim(cell.Value, "\n \t")
			if len(name) < 2 {
				continue
			}
			//			fmt.Printf("%d 	- %s\t", i, cell)
			name = strings.Replace(name, ",", "", -1)
			name = strings.Replace(name, ".", "", -1)
			name = strings.ToLower(name)
			nameParts := strings.Split(name, " ")
			for i, _ := range nameParts {
				nameParts[i] = strings.Trim(nameParts[i], " \t\n")
			}
			sort.Strings(nameParts)
			dict1 = append(dict1, Name{Actual: name, Original: strings.Join(nameParts, " "), NameParts: nameParts, Initials: getInitials(nameParts)})
		}
		//		fmt.Println()
	}
	sort.Sort(NameDict(dict1))
	return dict1
}

func getInitials(parts []string) string {
	var name []string = make([]string, len(parts))
	for i, p := range parts {
		name[i] = string(p[0])
	}
	return strings.Join(name, "")
}
