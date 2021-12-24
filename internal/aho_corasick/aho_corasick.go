package aho_corasick

type State struct {
	Id         int
	Char       rune
	NextStates map[rune]int
}

func (s *State) NewState(id int, char rune) {
	s.Id = id
	s.Char = char
	s.NextStates = make(map[rune]int)
}

type AhoCorasick struct {
	CurrState  int
	StateMap   map[int]*State
	OutputMap  map[int][]string
	FailureMap map[int]int
	DoneState  bool
}

func (ac *AhoCorasick) NewAhoCorasick() {
	ac.StateMap = make(map[int]*State)
	ac.OutputMap = make(map[int][]string)
	ac.CurrState = 0
	initState := new(State)
	initState.NewState(ac.CurrState, '\000')
	ac.StateMap[initState.Id] = initState
	ac.DoneState = false
}

func (ac *AhoCorasick) GoTo(sId int, a rune) (int, bool) {
	var s *State
	var ok bool

	if s, ok = ac.StateMap[sId]; ok == false {
		return -1, false
	}

	var n int
	if n, ok = s.NextStates[a]; ok == true {
		return n, true
	}

	if sId == 0 && ac.DoneState {
		return 0, true
	}

	return -1, false
}

func (ac *AhoCorasick) AddString(input string, output string) {
	cs := 0
	for _, i := range input {
		if _, ok := ac.GoTo(cs, i); ok != true {
			ac.CurrState++
			s := new(State)
			s.NewState(ac.CurrState, i)
			ac.StateMap[s.Id] = s
			ac.StateMap[cs].NextStates[i] = s.Id
		}
		cs = ac.StateMap[cs].NextStates[i]
	}
	ac.OutputMap[cs] = []string{output}
}

func (ac *AhoCorasick) Failure() {
	ac.DoneState = true
	ac.FailureMap = make(map[int]int)
	queue := []int{}

	for _, sId := range ac.StateMap[0].NextStates {
		ac.FailureMap[sId] = 0
		queue = append(queue, sId)
	}

	qId := 0
	for {
		if qId == len(queue) {
			break
		}
		sId := queue[qId]
		s := ac.StateMap[sId]
		for na, nId := range s.NextStates {
			queue = append(queue, nId)
			fsId := ac.FailureMap[sId]
			for {
				var nFsId int
				var ok bool
				if nFsId, ok = ac.GoTo(fsId, na); ok == true {
					ac.FailureMap[nId] = nFsId
					ac.OutputMap[nId] = append(ac.OutputMap[nId], ac.OutputMap[nFsId]...)
					break
				}
				fsId = ac.FailureMap[fsId]
			}
		}
		qId++
	}
}

func (ac *AhoCorasick) FirstMatch(str string) []string {
	sId := 0
	for _, a := range str {
		var ok bool
		for {
			if sId, ok = ac.GoTo(sId, a); ok == true {
				break
			}
			sId = ac.FailureMap[sId]
		}
		if o, ok := ac.OutputMap[sId]; ok == true && len(o) > 0 {
			return o
		}
	}
	return []string{}
}
