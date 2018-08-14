package cmd

// VariableValuesAddOptions are the options used when add a value to a variable.
type VariableValuesAddOptions struct {
	ID    string
	Value string
}

// VariableValuesAdder uses the client to add a variable value as described by options.
type VariableValuesAdder interface {
	Do(options VariableValuesAddOptions) error
}

// NewVariableValuesAdder returns a VariableValuersAdder that uses client to add a variable value.
func NewVariableValuesAdder(client VariableClient) VariableValuesAdder {
	return adderImpl{
		VariableClient: client,
	}
}

type adderImpl struct {
	VariableClient
}

func (v adderImpl) Do(options VariableValuesAddOptions) error {
	return v.AddSecret(options.ID, options.Value)
}
