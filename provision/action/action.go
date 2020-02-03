package action

var (
	before Action
	after  Action
	args   []interface{}
)

// Action represents an arbitrary execution that can be used to extend Hydroform
type Action interface {
	Run(args ...interface{}) (interface{}, error)
}

// SetBefore defines which action will be executed before an Hydroform operation.
func SetBefore(a Action) {
	before = a
}

// Before runs the action set with SetBefore. It is called and evaluated before each Hydroform operation (Provision, Status, Credentials and Deprovision)
// After running, the set action is cleared.
func Before() error {
	// clear the action after running it
	defer func() {
		before = nil
	}()

	if before != nil {
		_, err := before.Run(args...)
		return err
	}
	return nil
}

// SetAfter defines which action will be executed after an Hydroform operation.
func SetAfter(a Action) {
	after = a
}

// After runs the action set with SetAfter. It is called and evaluated after each Hydroform operation if there are no errors (Provision, Status, Credentials and Deprovision)
// After running, the set action is cleared.
func After() error {
	// clear the action after running it
	defer func() {
		after = nil
	}()

	if after != nil {
		_, err := after.Run(args...)
		return err
	}
	return nil
}

// SetArgs allows to define arbitrary arguments that Before and After actions will consume.
// Calling SetArgs a second time clears the args from the previous call.
func SetArgs(a ...interface{}) {
	args = a
}

// Args returns the defined arguments for the actions
func Args() []interface{} {
	return args
}

// FuncAction allows to use a pure function as an Action. By creating a function with this signature it cn be directly used as action.
// See examples on the unit tests.
type FuncAction func(args ...interface{}) (interface{}, error)

func (f FuncAction) Run(args ...interface{}) (interface{}, error) {
	return f(args...)
}
