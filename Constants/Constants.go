package Constants

/*
dimensions increase -> parallel
points increase -> VPT
*/
const (
	NumOfVectors             uint32 = 1000000
	NumOfDimensions          uint16 = 20
	Min                             = -50
	Max                             = 50
	K                               = 20
	MaxRoutinesForParallel          = 12
	ThresholdToRunParallel          = 100
	MaxRoutinesForTreeSearch        = 3000
)
