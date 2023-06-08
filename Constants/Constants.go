package Constants

/*
dimensions increase -> parallel
points increase -> VPT
*/
const (
	NumOfVectors           uint32 = 1000000
	NumOfDimensions        uint16 = 20
	Min                           = -50
	Max                           = 50
	K                             = 20
	MaxCPUs                       = 12
	ThresholdToRunParallel        = 100
	MaximumGoroutines             = 3000
)
