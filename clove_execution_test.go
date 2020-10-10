package golik

import (
	"context"
	"testing"
)

type TestBehavior struct{}

var (
	behaviorA = func(i interface{}) {}
	
	behaviorB = func(i interface{}) error {
		return nil
	}
	behaviorC = func(i interface{}) interface{} {
		return i
	}
	behaviorD = func(i interface{}) (interface{}, error) {
		switch i.(type) {
		case error:
			return nil, i.(error)
		default:
			return i, nil
		}
	}

	behaviorE = func(ctx CloveContext) {}
	behaviorF = func(ctx CloveContext) error {
		return nil
	}
	behaviorG = func(ctx CloveContext) interface{} {
		return nil
	}
	behaviorH = func(ctx CloveContext) (interface{}, error) {
		return nil, nil
	}

	behaviorI = func(ctx context.Context, i interface{}) {}
	behaviorJ = func(ctx context.Context, i interface{}) error {
		return nil
	}
	behaviorK = func(ctx context.Context, i interface{}) interface{} {
		return i
	}
	behaviorL = func(ctx context.Context, i interface{}) (interface{}, error) {
		switch i.(type) {
		case error:
			return nil, i.(error)
		default:
			return i, nil
		}
	}

	behaviorM = TestBehavior{}

	behaviorN = &behaviorM


	behaviorErrA = func() error { 
		return nil 
	}

	behaviorErrB = func() {}

	behaviorErrC = func(ctx context.Context, i interface{}, j interface{}) interface{} {
		return i
	}

	behaviorErrD = 28
	behaviorErrE = &behaviorErrD

	behaviorValid = []interface{}{behaviorA, behaviorB, behaviorC, behaviorD, behaviorE, behaviorF, behaviorG, behaviorH, behaviorI, behaviorJ, behaviorK, behaviorL, behaviorM, behaviorN}
	behaviorNotValid = []interface{}{behaviorErrA, behaviorErrB, behaviorErrC, behaviorErrD, behaviorErrE}


	lifecycleA = func() {}
	lifecycleB = func() error { return nil }
	lifecycleC = func(ctx context.Context) {}
	lifecycleD = func(ctx context.Context) error { return nil }

	lifecycleValid = []interface{}{lifecycleA, lifecycleB, lifecycleC, lifecycleD}
	lifecycleNotValid = []interface{}{behaviorA, behaviorB, behaviorC, behaviorD, behaviorI, behaviorJ, behaviorK, behaviorL, behaviorM, behaviorN}
)

func TestCheckBehavior(t *testing.T) {
	for _, fx := range behaviorValid {
		t.Logf("Check %T", fx)
		if err := checkBehavior(fx); err != nil {
			t.Errorf("\tcheckBehavior %T must be valid, but got error: %v", fx, err)
		}
	}
	for _, fe := range behaviorNotValid {
		t.Logf("Check unvalid %T", fe)
		if err := checkBehavior(fe); err == nil {
			t.Errorf("\tcheckBehavior %T must return an error", fe)
		}
	}
}

func TestCheckLifecycleFunc(t *testing.T) {
	for _, fx := range lifecycleValid {
		t.Logf("Check %T", fx)
		if err := checkLifecycleFunc(fx); err != nil {
			t.Errorf("\tcheckLifecycleFunc %T must be valid, but got error: %v", fx, err)
		}
	}
	for _, fe := range lifecycleNotValid {
		t.Logf("Check unvalid %T", fe)
		if err := checkLifecycleFunc(fe); err == nil {
			t.Errorf("\tcheckLifecycleFunc %T must return an error", fe)
		}
	}
}
