package action

// ConjurClient is the composition of all the Conjur API interfaces used by the individual action clients.
type ConjurClient interface {
	VariableClient
}
