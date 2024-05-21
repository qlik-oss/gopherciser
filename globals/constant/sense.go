package constant

// HyperCubeDataMode
const (
	//HyperCubeDataModeStraight • S or DATA_MODE_STRAIGHT
	HyperCubeDataModeStraight  = "S"
	HyperCubeDataModeStraightL = "DATA_MODE_STRAIGHT"
	//HyperCubeDataModePivot • P or DATA_MODE_PIVOT
	HyperCubeDataModePivot  = "P"
	HyperCubeDataModePivotL = "DATA_MODE_PIVOT"
	//HyperCubeDataModePivotStack • K or DATA_MODE_PIVOT_STACK
	HyperCubeDataModePivotStack  = "K"
	HyperCubeDataModePivotStackL = "DATA_MODE_PIVOT_STACK"
	//HyperCubeDataModeTree • T or DATA_MODE_TREE
	HyperCubeDataModeTree  = "T"
	HyperCubeDataModeTreeL = "DATA_MODE_TREE"
)

// NxDimCellType
const (
	NxDimCellValue     = "V"
	NxDimCellEmpty     = "E"
	NxDimCellNormal    = "N"
	NxDimCellTotal     = "T"
	NxDimCellOther     = "O"
	NxDimCellAggr      = "A"
	NxDimCellPseudo    = "P"
	NxDimCellRoot      = "R"
	NxDimCellNull      = "U"
	NxDimCellGenerated = "G"
)

// DataReductionMode
const (
	DataReductionModeNone      = "N"
	DataReductionModeOneDim    = "D1"
	DataReductionModeScattered = "S"
	DataReductionModeClustered = "C"
	DataReductionModeStacked   = "ST"
)

// NxDimensionInfoGrouping
const (
	NxDimensionInfoGroupingNone       = "N"
	NxDimensionInfoGroupingHiearchy   = "H"
	NxDimensionInfoGroupingCollection = "C"
)
