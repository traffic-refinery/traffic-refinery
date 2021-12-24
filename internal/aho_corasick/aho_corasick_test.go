package aho_corasick

import (
	"fmt"
	"reflect"
	"testing"
)

func TestAhoCorasickState(t *testing.T) {
	fmt.Printf("Running basic state test\n")

	ac := new(AhoCorasick)
	ac.NewAhoCorasick()

	output := make(map[int][]string)

	testStringI := "ASDF"
	testStringO := "asdf"
	ac.AddString(testStringI, testStringO)
	newStates := len(testStringI)
	output[newStates] = []string{testStringO}
	if !reflect.DeepEqual(output, ac.OutputMap) ||
		ac.CurrState != newStates {
		t.Error(fmt.Printf(`Adding new string failed
		  expected states %d, got %d
		  expected output %+v, got %+v`,
			newStates, ac.CurrState, output, ac.OutputMap))
	}

	ac.AddString(testStringI, testStringO)
	newStates = len(testStringI)
	if !reflect.DeepEqual(output, ac.OutputMap) ||
		ac.CurrState != newStates {
		t.Error(fmt.Printf(`Adding new string failed
		  expected states %d, got %d
		  expected output %+v, got %+v`,
			newStates, ac.CurrState, output, ac.OutputMap))
	}

	testStringI = "AVVF"
	testStringO = "avvf"
	ac.AddString(testStringI, testStringO)
	newStates += 3
	output[newStates] = []string{testStringO}
	if !reflect.DeepEqual(output, ac.OutputMap) ||
		ac.CurrState != newStates {
		t.Error(fmt.Printf(`Adding new string failed
		  expected states %d, got %d
		  expected output %+v, got %+v`,
			newStates, ac.CurrState, output, ac.OutputMap))
	}

	testStringI = "SVVF"
	testStringO = "svvf"
	newStates += 4
	ac.AddString(testStringI, testStringO)
	output[newStates] = []string{testStringO}
	if !reflect.DeepEqual(output, ac.OutputMap) ||
		ac.CurrState != newStates {
		t.Error(fmt.Printf(`Adding new string failed
		  expected states %d, got %d
		  expected output %+v, got %+v`,
			newStates, ac.CurrState, output, ac.OutputMap))
	}
}

func TestAhoCorasickBasicMatch(t *testing.T) {
	fmt.Printf("Running basic match test\n")

	ac := new(AhoCorasick)
	ac.NewAhoCorasick()
	for _, i := range []string{"he", "she", "his", "hers"} {
		ac.AddString(i, i)
	}
	ac.Failure()
	failure := map[int]int{1: 0, 2: 0, 3: 0, 4: 1, 5: 2, 6: 0, 7: 3, 8: 0, 9: 3}
	if !reflect.DeepEqual(failure, ac.FailureMap) {
		t.Error(fmt.Printf(`bad failure:
		  expected %+v
		  got      %+v
		  states   %+v`, failure, ac.FailureMap, ac.StateMap))
	}

	fm := ac.FirstMatch("ushers")
	efm := []string{"she", "he"}
	if !reflect.DeepEqual(fm, efm) {
		t.Error(fmt.Printf(`bad failure:
		  expected %+v
		  got      %+v`, fm, efm))
	}
	fm = ac.FirstMatch("asdf")
	if !reflect.DeepEqual(fm, []string{}) {
		t.Error(fmt.Printf(`bad failure:
		  expected %+v
		  got      %+v`, fm, []string{}))
	}
}
