// Package parser provides NFA construction and analysis utilities for regex patterns.
package parser

import (
	"fmt"
	"regexp/syntax"
)

// NFA represents a Non-deterministic Finite Automaton constructed from a regex.
type NFA struct {
	Start       *State
	Accept      *State
	States      []*State
	StateCount  int
	Transitions map[*State][]*Transition
}

// State represents a state in the NFA.
type State struct {
	ID          int
	IsAccept    bool
	Transitions []*Transition
	EpsilonTo   []*State // States reachable via epsilon transitions
}

// Transition represents a transition between states.
type Transition struct {
	From      *State
	To        *State
	Label     TransitionLabel
	IsEpsilon bool
}

// TransitionLabel represents what causes a transition.
type TransitionLabel struct {
	Type  TransitionType
	Runes []rune     // For literal characters
	Class *CharClass // For character classes
	Op    syntax.Op  // For special operations
}

// TransitionType indicates the type of transition.
type TransitionType int

const (
	TransitionLiteral TransitionType = iota // Match specific character(s)
	TransitionClass                         // Match character class
	TransitionAny                           // Match any character (.)
	TransitionEpsilon                       // Epsilon (no input consumed)
	TransitionAnchor                        // Anchor (^, $)
)

// CharClass represents a set of characters that can be matched.
type CharClass struct {
	Ranges []RuneRange // Inclusive ranges
	Negate bool        // True if this is a negated class
}

// RuneRange represents an inclusive range of runes.
type RuneRange struct {
	Lo rune
	Hi rune
}

// NewNFA creates a new empty NFA.
func NewNFA() *NFA {
	return &NFA{
		States:      make([]*State, 0),
		Transitions: make(map[*State][]*Transition),
		StateCount:  0,
	}
}

// NewState creates a new state and adds it to the NFA.
func (nfa *NFA) NewState() *State {
	state := &State{
		ID:          nfa.StateCount,
		IsAccept:    false,
		Transitions: make([]*Transition, 0),
		EpsilonTo:   make([]*State, 0),
	}
	nfa.States = append(nfa.States, state)
	nfa.StateCount++
	return state
}

// AddTransition adds a transition between two states.
func (nfa *NFA) AddTransition(from, to *State, label TransitionLabel) *Transition {
	trans := &Transition{
		From:      from,
		To:        to,
		Label:     label,
		IsEpsilon: label.Type == TransitionEpsilon,
	}

	from.Transitions = append(from.Transitions, trans)
	nfa.Transitions[from] = append(nfa.Transitions[from], trans)

	if trans.IsEpsilon {
		from.EpsilonTo = append(from.EpsilonTo, to)
	}

	return trans
}

// AddEpsilonTransition adds an epsilon transition (no input consumed).
func (nfa *NFA) AddEpsilonTransition(from, to *State) *Transition {
	return nfa.AddTransition(from, to, TransitionLabel{Type: TransitionEpsilon})
}

// BuildNFA constructs an NFA from a parsed regex AST.
func BuildNFA(re *syntax.Regexp) (*NFA, error) {
	nfa := NewNFA()

	// Create start and accept states
	start := nfa.NewState()
	accept := nfa.NewState()
	accept.IsAccept = true

	nfa.Start = start
	nfa.Accept = accept

	// Build NFA from regex
	if err := buildNFAFromRegexp(nfa, re, start, accept); err != nil {
		return nil, err
	}

	return nfa, nil
}

// buildNFAFromRegexp recursively builds NFA from regex AST.
func buildNFAFromRegexp(nfa *NFA, re *syntax.Regexp, start, accept *State) error {
	switch re.Op {
	case syntax.OpLiteral:
		return buildLiteral(nfa, re, start, accept)

	case syntax.OpCharClass:
		return buildCharClass(nfa, re, start, accept)

	case syntax.OpAnyChar, syntax.OpAnyCharNotNL:
		return buildAnyChar(nfa, re, start, accept)

	case syntax.OpConcat:
		return buildConcat(nfa, re, start, accept)

	case syntax.OpAlternate:
		return buildAlternate(nfa, re, start, accept)

	case syntax.OpStar:
		return buildStar(nfa, re, start, accept)

	case syntax.OpPlus:
		return buildPlus(nfa, re, start, accept)

	case syntax.OpQuest:
		return buildQuest(nfa, re, start, accept)

	case syntax.OpRepeat:
		return buildRepeat(nfa, re, start, accept)

	case syntax.OpCapture:
		// Treat capture groups as their contained expression
		if len(re.Sub) > 0 {
			return buildNFAFromRegexp(nfa, re.Sub[0], start, accept)
		}
		nfa.AddEpsilonTransition(start, accept)
		return nil

	case syntax.OpEmptyMatch:
		// Empty match: just epsilon transition
		nfa.AddEpsilonTransition(start, accept)
		return nil

	case syntax.OpBeginLine, syntax.OpEndLine, syntax.OpBeginText, syntax.OpEndText:
		// Anchors: treat as epsilon with special semantics
		nfa.AddTransition(start, accept, TransitionLabel{
			Type: TransitionAnchor,
			Op:   re.Op,
		})
		return nil

	default:
		// For unsupported operations, add epsilon transition
		nfa.AddEpsilonTransition(start, accept)
		return nil
	}
}

// buildLiteral builds NFA for literal string.
func buildLiteral(nfa *NFA, re *syntax.Regexp, start, accept *State) error {
	current := start

	for i, r := range re.Rune {
		var next *State
		if i == len(re.Rune)-1 {
			next = accept
		} else {
			next = nfa.NewState()
		}

		nfa.AddTransition(current, next, TransitionLabel{
			Type:  TransitionLiteral,
			Runes: []rune{r},
		})

		current = next
	}

	return nil
}

// buildCharClass builds NFA for character class [a-z].
func buildCharClass(nfa *NFA, re *syntax.Regexp, start, accept *State) error {
	class := &CharClass{
		Ranges: make([]RuneRange, 0),
		Negate: (re.Flags&syntax.ClassNL != 0),
	}

	// Convert rune pairs to ranges
	for i := 0; i < len(re.Rune); i += 2 {
		class.Ranges = append(class.Ranges, RuneRange{
			Lo: re.Rune[i],
			Hi: re.Rune[i+1],
		})
	}

	nfa.AddTransition(start, accept, TransitionLabel{
		Type:  TransitionClass,
		Class: class,
	})

	return nil
}

// buildAnyChar builds NFA for . (any character).
func buildAnyChar(nfa *NFA, re *syntax.Regexp, start, accept *State) error {
	nfa.AddTransition(start, accept, TransitionLabel{
		Type: TransitionAny,
		Op:   re.Op,
	})
	return nil
}

// buildConcat builds NFA for concatenation (ab).
func buildConcat(nfa *NFA, re *syntax.Regexp, start, accept *State) error {
	if len(re.Sub) == 0 {
		nfa.AddEpsilonTransition(start, accept)
		return nil
	}

	current := start
	for i, sub := range re.Sub {
		var next *State
		if i == len(re.Sub)-1 {
			next = accept
		} else {
			next = nfa.NewState()
		}

		if err := buildNFAFromRegexp(nfa, sub, current, next); err != nil {
			return err
		}

		current = next
	}

	return nil
}

// buildAlternate builds NFA for alternation (a|b).
func buildAlternate(nfa *NFA, re *syntax.Regexp, start, accept *State) error {
	// Split from start to each alternative, then join to accept
	for _, sub := range re.Sub {
		altStart := nfa.NewState()
		altEnd := nfa.NewState()

		nfa.AddEpsilonTransition(start, altStart)

		if err := buildNFAFromRegexp(nfa, sub, altStart, altEnd); err != nil {
			return err
		}

		nfa.AddEpsilonTransition(altEnd, accept)
	}

	return nil
}

// buildStar builds NFA for a* (zero or more).
func buildStar(nfa *NFA, re *syntax.Regexp, start, accept *State) error {
	if len(re.Sub) == 0 {
		nfa.AddEpsilonTransition(start, accept)
		return nil
	}

	loopStart := nfa.NewState()
	loopEnd := nfa.NewState()

	// Epsilon from start to loopStart and to accept (zero matches)
	nfa.AddEpsilonTransition(start, loopStart)
	nfa.AddEpsilonTransition(start, accept)

	// Build the repeated part
	if err := buildNFAFromRegexp(nfa, re.Sub[0], loopStart, loopEnd); err != nil {
		return err
	}

	// Loop back from end to start
	nfa.AddEpsilonTransition(loopEnd, loopStart)

	// Exit to accept
	nfa.AddEpsilonTransition(loopEnd, accept)

	return nil
}

// buildPlus builds NFA for a+ (one or more).
func buildPlus(nfa *NFA, re *syntax.Regexp, start, accept *State) error {
	if len(re.Sub) == 0 {
		nfa.AddEpsilonTransition(start, accept)
		return nil
	}

	loopStart := nfa.NewState()
	loopEnd := nfa.NewState()

	// Must match at least once
	nfa.AddEpsilonTransition(start, loopStart)

	// Build the repeated part
	if err := buildNFAFromRegexp(nfa, re.Sub[0], loopStart, loopEnd); err != nil {
		return err
	}

	// Loop back from end to start (for multiple matches)
	nfa.AddEpsilonTransition(loopEnd, loopStart)

	// Exit to accept
	nfa.AddEpsilonTransition(loopEnd, accept)

	return nil
}

// buildQuest builds NFA for a? (zero or one).
func buildQuest(nfa *NFA, re *syntax.Regexp, start, accept *State) error {
	if len(re.Sub) == 0 {
		nfa.AddEpsilonTransition(start, accept)
		return nil
	}

	// Epsilon to accept (zero matches)
	nfa.AddEpsilonTransition(start, accept)

	// Or match once
	return buildNFAFromRegexp(nfa, re.Sub[0], start, accept)
}

// buildRepeat builds NFA for a{n,m} (bounded repetition).
func buildRepeat(nfa *NFA, re *syntax.Regexp, start, accept *State) error {
	if len(re.Sub) == 0 {
		nfa.AddEpsilonTransition(start, accept)
		return nil
	}

	min, max := re.Min, re.Max

	// Build min required repetitions
	current := start
	for i := 0; i < min; i++ {
		next := nfa.NewState()
		if err := buildNFAFromRegexp(nfa, re.Sub[0], current, next); err != nil {
			return err
		}
		current = next
	}

	// Build optional repetitions up to max
	if max == -1 {
		// Unbounded: a{n,} is like a{n}a*
		loopStart := nfa.NewState()
		loopEnd := nfa.NewState()

		nfa.AddEpsilonTransition(current, loopStart)
		nfa.AddEpsilonTransition(current, accept)

		if err := buildNFAFromRegexp(nfa, re.Sub[0], loopStart, loopEnd); err != nil {
			return err
		}

		nfa.AddEpsilonTransition(loopEnd, loopStart)
		nfa.AddEpsilonTransition(loopEnd, accept)
	} else {
		// Bounded: add optional paths for each additional repetition
		for i := min; i < max; i++ {
			next := nfa.NewState()

			// Can skip this repetition
			nfa.AddEpsilonTransition(current, next)

			// Or match it
			if err := buildNFAFromRegexp(nfa, re.Sub[0], current, next); err != nil {
				return err
			}

			current = next
		}

		nfa.AddEpsilonTransition(current, accept)
	}

	return nil
}

// String returns a string representation of the NFA for debugging.
func (nfa *NFA) String() string {
	return fmt.Sprintf("NFA{States:%d, Start:%d, Accept:%d}",
		len(nfa.States), nfa.Start.ID, nfa.Accept.ID)
}

// ComputeEpsilonClosure computes the epsilon closure of a state.
// Returns all states reachable from the given state via epsilon transitions.
func ComputeEpsilonClosure(state *State) map[*State]bool {
	closure := make(map[*State]bool)
	computeEpsilonClosureHelper(state, closure)
	return closure
}

func computeEpsilonClosureHelper(state *State, closure map[*State]bool) {
	if closure[state] {
		return
	}

	closure[state] = true

	for _, next := range state.EpsilonTo {
		computeEpsilonClosureHelper(next, closure)
	}
}
