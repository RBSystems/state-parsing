package base

//BufferManager is meant to handle buffering events/updates to the eventual forever home of the information
type BufferManager interface {
	Send(toSend interface{}) error
}
