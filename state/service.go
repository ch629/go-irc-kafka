package state

//go:generate mockery --name=Service
type Service interface {
	// TODO: Where do we want to hold this state & do we actually need it?
	JoinChannel(channelName string)
	LeaveChannel(channelName string) error
	IsInChannel(channelName string) bool
}
