package quantum

import "github.com/mitchellh/multistep"

// Executor executes steps
type Executor interface {
	Execute(statebag StateBag) error
}

// BasicExecutor implements Executor
type BasicExecutor struct {
	Steps []Step
}

// Run steps
func (r *BasicExecutor) Run(state StateBag) error {
	steps := make([]multistep.Step, len(r.Steps))
	for i, s := range r.Steps {
		steps[i] = ToMultiStep(s)
	}

	runner := &multistep.BasicRunner{Steps: steps}
	runner.Run(state)

	if rawErr, ok := state.GetOk("error"); ok {
		return rawErr.(error)
	}
	return nil
}

// Step defines a step that can run and clean up
type Step interface {
	Run(state StateBag) error
	Cleanup(state StateBag)
}

// StateBag defines a statebag
type StateBag struct {
	multistep.StateBag
}

// NewStateBag creates a new StateBag
func NewStateBag() StateBag {
	return StateBag{
		StateBag: new(multistep.BasicStateBag),
	}
}

type step struct {
	s Step
}

func (s *step) Run(statebag multistep.StateBag) multistep.StepAction {
	if err := s.s.Run(StateBag{statebag}); err != nil {
		statebag.Put("error", err)
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *step) Cleanup(statebag multistep.StateBag) {
	s.s.Cleanup(StateBag{statebag})
}

// ToMultiStep converts Step to multistep.Step
func ToMultiStep(s Step) multistep.Step {
	return &step{
		s: s,
	}
}
