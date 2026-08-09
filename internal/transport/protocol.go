package transport

import (
	"bytes"
	"strings"
)

const (
	CmdShutdown = "SHUTDOWN"
	CmdExit     = "EXIT"
	CmdStats    = "STATS"

	MsgChatStarted       = "System Chat started. You can send messages."
	MsgOtherDisconnected = "INFO: Other client disconnected. Chat ended."
	MsgServerShutdown    = "INFO: Server is shutting down."
	MsgMaxClients        = "ERROR: Server is full (max 2 clients)."
)

func ParseLine(line []byte) (string, bool) {
	s := strings.TrimSuffix(string(bytes.TrimRight(line, "\r\n")), "\n")
	if s == "" {
		return "", false
	}
	return s, true
}

func IsCommand(s string) bool {
	return s == CmdShutdown || s == CmdExit || s == CmdStats
}
