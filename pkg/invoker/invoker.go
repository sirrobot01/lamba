package invoker

// Invoker is a struct that represents a function Invoker http, event, cli, etc

type Invoker interface {
	Start() error
	Stop() error
}
