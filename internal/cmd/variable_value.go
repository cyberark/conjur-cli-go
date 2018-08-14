package cmd

// VariableValueOptions are the options used when retrieving a variable value.
type VariableValueOptions struct {
	ID string
}

// VariableValuer uses the client to retrieve the value of the variable described by options.
type VariableValuer interface {
	Do(options VariableValueOptions) ([]byte, error)
}

// NewVariableValuer returns a VariableValuer implementation that uses client to retrieve a value.
func NewVariableValuer(client VariableClient) VariableValuer {
	return valuerImpl{
		VariableClient: client,
	}
}

type valuerImpl struct {
	VariableClient
}

func (v valuerImpl) Do(options VariableValueOptions) ([]byte, error) {
	return v.RetrieveSecret(options.ID)
}
