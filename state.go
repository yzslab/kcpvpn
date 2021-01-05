package main

type State interface {
	HasNextState() bool
	NextState(arguments interface{}) (interface{}, error)
	setNextState(func(arguments interface{}) (interface{}, error))
}

type HasState struct {
	nextState func(arguments interface{}) (interface{}, error)
}

func (n *HasState) setNextState(nextState func(arguments interface{}) (interface{}, error)) {
	n.nextState = nextState
}

func (n *HasState) HasNextState() bool {
	return n.nextState != nil
}

func (n *HasState) NextState(arguments interface{}) (interface{}, error) {
	return n.nextState(arguments)
}

func IterateState(state State) (interface{}, error) {
	var arguments interface{}
	var err error
	for state.HasNextState() {
		arguments, err = state.NextState(arguments)
		if err != nil {
			return nil, err
		}
	}
	return arguments, nil
}
