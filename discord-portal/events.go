package discordportal

const (
	// Websocket OP Codes
	OP_HEARTBEAT     int = 1
	OP_IDENTIFY      int = 2
	OP_RESUME        int = 6
	OP_RECONNECT     int = 7
	OP_HELLO         int = 10
	OP_HEARTBEAT_ACK int = 11

	// Events
	READY          string = "READY"
	MESSAGE_CREATE string = "MESSAGE_CREATE"
)
