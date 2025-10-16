package parser

import (
	"testing"
)

func TestNewNFA(t *testing.T) {
	nfa := NewNFA()

	if nfa == nil {
		t.Fatal("NewNFA() returned nil")
	}

	if len(nfa.States) != 0 {
		t.Errorf("New NFA should have 0 states, got %d", len(nfa.States))
	}

	if nfa.StateCount != 0 {
		t.Errorf("New NFA should have StateCount 0, got %d", nfa.StateCount)
	}
}

func TestNFA_NewState(t *testing.T) {
	nfa := NewNFA()

	s1 := nfa.NewState()
	if s1.ID != 0 {
		t.Errorf("First state should have ID 0, got %d", s1.ID)
	}

	s2 := nfa.NewState()
	if s2.ID != 1 {
		t.Errorf("Second state should have ID 1, got %d", s2.ID)
	}

	if len(nfa.States) != 2 {
		t.Errorf("NFA should have 2 states, got %d", len(nfa.States))
	}
}

func TestNFA_AddTransition(t *testing.T) {
	nfa := NewNFA()
	s1 := nfa.NewState()
	s2 := nfa.NewState()

	trans := nfa.AddTransition(s1, s2, TransitionLabel{
		Type:  TransitionLiteral,
		Runes: []rune{'a'},
	})

	if trans == nil {
		t.Fatal("AddTransition() returned nil")
	}

	if trans.From != s1 {
		t.Error("Transition From state incorrect")
	}

	if trans.To != s2 {
		t.Error("Transition To state incorrect")
	}

	if len(s1.Transitions) != 1 {
		t.Errorf("State should have 1 transition, got %d", len(s1.Transitions))
	}
}

func TestNFA_AddEpsilonTransition(t *testing.T) {
	nfa := NewNFA()
	s1 := nfa.NewState()
	s2 := nfa.NewState()

	trans := nfa.AddEpsilonTransition(s1, s2)

	if !trans.IsEpsilon {
		t.Error("Epsilon transition should have IsEpsilon = true")
	}

	if len(s1.EpsilonTo) != 1 {
		t.Errorf("State should have 1 epsilon transition, got %d", len(s1.EpsilonTo))
	}

	if s1.EpsilonTo[0] != s2 {
		t.Error("Epsilon transition points to wrong state")
	}
}

func TestBuildNFA_SimpleLiteral(t *testing.T) {
	p := NewParser()
	re := p.MustParse("abc")

	nfa, err := BuildNFA(re)
	if err != nil {
		t.Fatalf("BuildNFA() error = %v", err)
	}

	if nfa.Start == nil {
		t.Error("NFA should have a start state")
	}

	if nfa.Accept == nil {
		t.Error("NFA should have an accept state")
	}

	if !nfa.Accept.IsAccept {
		t.Error("Accept state should be marked as IsAccept")
	}

	// Should have at least start and accept states
	if len(nfa.States) < 2 {
		t.Errorf("NFA should have at least 2 states, got %d", len(nfa.States))
	}
}

func TestBuildNFA_Alternation(t *testing.T) {
	p := NewParser()
	// Use multi-char alternatives to prevent simplification to char class
	re := p.MustParse("ab|cd")

	nfa, err := BuildNFA(re)
	if err != nil {
		t.Fatalf("BuildNFA() error = %v", err)
	}

	// Alternation creates multiple paths from start
	if len(nfa.Start.EpsilonTo) == 0 {
		t.Error("Alternation should create epsilon transitions from start")
	}

	t.Logf("NFA for 'ab|cd': %s", nfa)
}

func TestBuildNFA_Star(t *testing.T) {
	p := NewParser()
	re := p.MustParse("a*")

	nfa, err := BuildNFA(re)
	if err != nil {
		t.Fatalf("BuildNFA() error = %v", err)
	}

	// Star should allow skipping (epsilon to accept)
	// and looping back
	if len(nfa.States) < 3 {
		t.Errorf("Star should create additional states, got %d", len(nfa.States))
	}

	t.Logf("NFA for 'a*': %s", nfa)
}

func TestBuildNFA_Plus(t *testing.T) {
	p := NewParser()
	re := p.MustParse("a+")

	nfa, err := BuildNFA(re)
	if err != nil {
		t.Fatalf("BuildNFA() error = %v", err)
	}

	// Plus requires at least one match, so different structure from star
	if nfa.Start == nfa.Accept {
		t.Error("Plus should not have start == accept")
	}

	t.Logf("NFA for 'a+': %s", nfa)
}

func TestBuildNFA_Quest(t *testing.T) {
	p := NewParser()
	re := p.MustParse("a?")

	nfa, err := BuildNFA(re)
	if err != nil {
		t.Fatalf("BuildNFA() error = %v", err)
	}

	// Quest allows zero or one, should have direct path to accept
	hasDirectPath := false
	for _, state := range nfa.Start.EpsilonTo {
		if state == nfa.Accept {
			hasDirectPath = true
			break
		}
	}

	if !hasDirectPath {
		t.Error("Quest should allow direct epsilon to accept")
	}

	t.Logf("NFA for 'a?': %s", nfa)
}

func TestBuildNFA_NestedQuantifiers(t *testing.T) {
	p := NewParser()
	re := p.MustParse("(a+)+")

	nfa, err := BuildNFA(re)
	if err != nil {
		t.Fatalf("BuildNFA() error = %v", err)
	}

	// Nested quantifiers create complex NFA with multiple loops
	if len(nfa.States) < 5 {
		t.Errorf("Nested quantifiers should create more states, got %d", len(nfa.States))
	}

	t.Logf("NFA for '(a+)+': %s with %d states", nfa, len(nfa.States))
}

func TestBuildNFA_CharClass(t *testing.T) {
	p := NewParser()
	re := p.MustParse("[a-z]")

	nfa, err := BuildNFA(re)
	if err != nil {
		t.Fatalf("BuildNFA() error = %v", err)
	}

	// Character class should create a single transition
	if len(nfa.Start.Transitions) == 0 {
		t.Error("Character class should create at least one transition")
	}

	trans := nfa.Start.Transitions[0]
	if trans.Label.Type != TransitionClass {
		t.Errorf("Expected TransitionClass, got %v", trans.Label.Type)
	}
}

func TestBuildNFA_AnyChar(t *testing.T) {
	p := NewParser()
	re := p.MustParse(".")

	nfa, err := BuildNFA(re)
	if err != nil {
		t.Fatalf("BuildNFA() error = %v", err)
	}

	if len(nfa.Start.Transitions) == 0 {
		t.Error("AnyChar should create a transition")
	}

	trans := nfa.Start.Transitions[0]
	if trans.Label.Type != TransitionAny {
		t.Errorf("Expected TransitionAny, got %v", trans.Label.Type)
	}
}

func TestBuildNFA_Concat(t *testing.T) {
	p := NewParser()
	re := p.MustParse("abc")

	nfa, err := BuildNFA(re)
	if err != nil {
		t.Fatalf("BuildNFA() error = %v", err)
	}

	// Concatenation should create a chain of states
	// Start -> a -> b -> c -> Accept
	// So at least 5 states (or simplified)
	if len(nfa.States) < 2 {
		t.Errorf("Concat should create multiple states, got %d", len(nfa.States))
	}
}

func TestComputeEpsilonClosure(t *testing.T) {
	nfa := NewNFA()
	s1 := nfa.NewState()
	s2 := nfa.NewState()
	s3 := nfa.NewState()
	s4 := nfa.NewState()

	// Create epsilon transitions: s1 -> s2 -> s3 -> s4
	nfa.AddEpsilonTransition(s1, s2)
	nfa.AddEpsilonTransition(s2, s3)
	nfa.AddEpsilonTransition(s3, s4)

	closure := ComputeEpsilonClosure(s1)

	// Should include all reachable states
	expected := []*State{s1, s2, s3, s4}
	for _, state := range expected {
		if !closure[state] {
			t.Errorf("Epsilon closure should include state %d", state.ID)
		}
	}

	if len(closure) != 4 {
		t.Errorf("Epsilon closure should have 4 states, got %d", len(closure))
	}
}

func TestComputeEpsilonClosure_NoEpsilon(t *testing.T) {
	nfa := NewNFA()
	s1 := nfa.NewState()
	s2 := nfa.NewState()

	// Add non-epsilon transition
	nfa.AddTransition(s1, s2, TransitionLabel{
		Type:  TransitionLiteral,
		Runes: []rune{'a'},
	})

	closure := ComputeEpsilonClosure(s1)

	// Should only include self
	if len(closure) != 1 {
		t.Errorf("Epsilon closure with no epsilon transitions should have 1 state, got %d", len(closure))
	}

	if !closure[s1] {
		t.Error("Epsilon closure should include the state itself")
	}
}

func TestComputeEpsilonClosure_Cycle(t *testing.T) {
	nfa := NewNFA()
	s1 := nfa.NewState()
	s2 := nfa.NewState()

	// Create epsilon cycle: s1 -> s2 -> s1
	nfa.AddEpsilonTransition(s1, s2)
	nfa.AddEpsilonTransition(s2, s1)

	closure := ComputeEpsilonClosure(s1)

	// Should handle cycle without infinite loop
	if len(closure) != 2 {
		t.Errorf("Epsilon closure with cycle should have 2 states, got %d", len(closure))
	}

	if !closure[s1] || !closure[s2] {
		t.Error("Epsilon closure should include both states in cycle")
	}
}

func TestBuildNFA_ComplexPattern(t *testing.T) {
	patterns := []string{
		"(a|b)*c",
		"a+b*c?",
		"(ab)+|(cd)*",
		"[a-z]+@[a-z]+\\.[a-z]{2,}",
	}

	p := NewParser()
	for _, pattern := range patterns {
		t.Run(pattern, func(t *testing.T) {
			re, err := p.Parse(pattern)
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}

			nfa, err := BuildNFA(re)
			if err != nil {
				t.Fatalf("BuildNFA() error = %v", err)
			}

			if nfa.Start == nil || nfa.Accept == nil {
				t.Error("NFA should have start and accept states")
			}

			t.Logf("Built NFA for '%s': %s", pattern, nfa)
		})
	}
}
