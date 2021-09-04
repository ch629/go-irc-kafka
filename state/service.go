package state

//go:generate mockery --name=Service
type Service interface {
	// TODO: Is this a good way to handle it?
	JoinChannel(channelName string)
	LeaveChannel(channelName string) error
	IsInChannel(channelName string) bool
}
