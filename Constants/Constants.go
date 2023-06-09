package Constants

/*
dimensions increase -> parallel
points increase -> VPT
*/
const (
	NumOfVectors             uint32 = 10000
	NumOfDimensions          uint16 = 3000
	Min                             = -127
	Max                             = 127
	K                               = 20
	MaxRoutinesForParallel          = 20
	ThresholdToRunParallel          = 10
	MaxRoutinesForTreeSearch        = 2000
)
