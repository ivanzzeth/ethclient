package message

type Sequencer interface {
	PushMsg(msg Request) error
	// block if no any msgs return
	PopMsg() (Request, error)
	PeekMsg() (Request, error)
	QueuedMsgCount() (int, error)
	PendingMsgCount() (int, error)
	Close()
}
