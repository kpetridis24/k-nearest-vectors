package Constants

/*
dimensions increase -> parallel
points increase -> VPT
*/
const (
	NumOfVectors             uint32 = 200000
	NumOfDimensions          uint16 = 300
	Min                             = -50
	Max                             = 50
	K                               = 20
	MaxRoutinesForParallel          = 10
	ThresholdToRunParallel          = 100
	MaxRoutinesForTreeSearch        = 3000
)
