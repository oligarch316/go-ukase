package ukcore

import (
	"context"
	"errors"
	"fmt"
)

var (
	muxDefaultHandler = Handler(muxHandleUnspecified)
	muxDefaultCommand = Command{Executor: muxDefaultHandler, DefaultParams: struct{}{}}
	muxDefaultConfig  = MuxConfig{
		AllowOverwrite: false,
		DefaultCommand: muxDefaultCommand,
	}
)

func muxHandleUnspecified(_ context.Context, input Input) error {
	return ErrorMuxNotSpecified{Input: input}
}

type muxCommand struct {
	executor      Executor
	paramsDefault any
	paramsInfo    ParamsInfo
}

func newMuxCommand(executor Executor, paramsDefault any) (command muxCommand, err error) {
	command.executor = executor
	command.paramsDefault = paramsDefault
	command.paramsInfo, err = ParamsInfoOf(paramsDefault)
	return
}

type muxNode struct {
	muxCommand
	children  map[string]*muxNode
	specified bool
}

func newMuxNode(command muxCommand) *muxNode {
	children := make(map[string]*muxNode)
	return &muxNode{children: children, muxCommand: command}
}

type MuxConfig struct {
	AllowOverwrite bool
	DefaultCommand Command
}

type Mux struct {
	rootNode *muxNode

	allowOverwrite bool
	defaultCommand muxCommand
}

func NewMux(opts ...func(*MuxConfig)) (*Mux, error) {
	config := muxDefaultConfig
	for _, opt := range opts {
		opt(&config)
	}

	allowOverwrite := config.AllowOverwrite
	defaultExecutor := config.DefaultCommand.Executor
	defaultParams := config.DefaultCommand.DefaultParams

	defaultCommand, err := newMuxCommand(defaultExecutor, defaultParams)
	if err != nil {
		return nil, err
	}

	mux := &Mux{
		rootNode:       newMuxNode(defaultCommand),
		allowOverwrite: allowOverwrite,
		defaultCommand: defaultCommand,
	}

	return mux, nil
}

func (m *Mux) Register(executor Executor, defaultParams any, target ...string) error {
	command, err := newMuxCommand(executor, defaultParams)
	if err != nil {
		return err
	}

	node := m.rootNode

	for _, childName := range target {
		child, ok := node.children[childName]
		if !ok {
			child = newMuxNode(m.defaultCommand)
			node.children[childName] = child
		}

		node = child
	}

	if node.specified && !m.allowOverwrite {
		return errors.New("[TODO Add] diallowed overwrite")
	}

	node.muxCommand, node.specified = command, true
	return nil
}

func (m *Mux) Execute(ctx context.Context, values []string) error {
	if len(values) == 0 {
		return errors.New("[TODO Run] empty vales")
	}

	// Consume the 1st value as the program name, remainder as raw input
	programName, parser := values[0], NewParser(values[1:])

	// Initialize
	input := Input{Target: []string{programName}}
	node := m.rootNode

	for {
		// Consume flag values for the current node
		flags, err := parser.ParseFlags(node.paramsInfo)
		if err != nil {
			return err
		}

		input.Flags = append(input.Flags, flags...)

		// Check for subcommand
		token := parser.ParseToken()

		if token.Kind == KindDelim || token.Kind == KindEOF {
			break
		}

		if token.Kind != KindString {
			return fmt.Errorf("[TODO Execute] <INTERNAL> got an unexpected token kind (%s)", token.Kind)
		}

		child, ok := node.children[token.Value]
		if !ok {
			input.Args = append(input.Args, token.Value)
			break
		}

		input.Target = append(input.Target, token.Value)
		node = child
	}

	// Consume any remaining values as arguments
	input.Args = append(input.Args, parser.Flush()...)

	// Delegate to executor
	return node.executor.Execute(ctx, input)
}
