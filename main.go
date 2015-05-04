package main

import (
	"fmt"
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

	threshHold, err := strconv.ParseFloat(args[4], 32)
	if err != nil {
		panic(err)
	}

	//	names1 := make([]string, 0)
	// xlsfile2 := xlsx.OpenFile(args[1])
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
	//	fmt.Printf("%s\n\n", dict1)
	//	fmt.Printf("%s", dict2)

	for _, name := range dict1 {
		for _, otherName := range dict2 {
			name1 := strings.Join(name, " ")
			name2 := strings.Join(otherName, " ")
			level := smetrics.JaroWinkler(name1, name2, 0.7, 4)
			if level > threshHold {
				fmt.Printf(" %f - %s <==> %s\n", level, name1, name2)
			}
		}
	}
}

type NameParts []string
type NameDict []NameParts

func (a NameParts) Len() int {
	return len(a)
}

func (a NameDict) Len() int {
	return len(a)
}

func (a NameDict) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a NameDict) Less(i, j int) bool {

	return a[i][0][0] < a[j][0][0]
}

func makeNameDictionary(file *xlsx.File, colNumber int) []NameParts {
	dict1 := make([]NameParts, 0)
	sheet1 := file.Sheets[0]
	for _, row := range sheet1.Rows {
		for i, cell := range row.Cells {
			if i != colNumber {
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
			dict1 = append(dict1, nameParts)
		}
		//		fmt.Println()
	}
	sort.Sort(NameDict(dict1))
	return dict1
}
