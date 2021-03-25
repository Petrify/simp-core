package commands

type InvalidCommandError struct{ error }

type InvalidArgsError struct{ error }

type ExecutionError struct{ error }
