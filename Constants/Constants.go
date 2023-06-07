package Constants

/*
dimensions increase -> parallel
points increase -> VPT
*/
const (
	NumOfVectors           uint32 = 100000
	NumOfDimensions        uint16 = 100
	Min                           = -50
	Max                           = 50
	K                             = 20
	MaxCPUs                       = 12
	ThresholdToRunParallel        = 100
	MaximumGoroutines             = 3000
)
