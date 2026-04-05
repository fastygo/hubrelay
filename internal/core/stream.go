package core

import "sshbot/pkg/contract"

type StreamChunk = contract.StreamChunk
type StreamWriter = contract.StreamWriter
type StreamingPlugin = contract.StreamingPlugin
type BufferedStreamWriter = contract.BufferedStreamWriter

func cloneCommandResult(result CommandResult) CommandResult {
	return contract.CloneCommandResult(result)
}
