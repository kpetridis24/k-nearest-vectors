package Constants

/*
dimensions increase -> parallel
points increase -> VPT
*/
const (
	NumOfVectors             uint32 = 100000
	NumOfDimensions          uint16 = 300
	Min                             = -127
	Max                             = 127
	K                               = 20
	MaxRoutinesForParallel          = 20
	ThresholdToRunParallel          = 10
	MaxRoutinesForTreeSearch        = 2000
)
