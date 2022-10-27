package fack

type NodeStatus uint8

const (
	Startup NodeStatus = 0
	Running            = 1
	Frozen             = 2
	Killed             = 3
)

func (nodeStatus NodeStatus) ToString() string {
	switch nodeStatus {
	case 0:
		return "Startup"
		break
	case 1:
		return "Running"
		break
	case 2:
		return "Frozen"
		break
	default:
		return "Killed"
	}

	return "Killed"
}
