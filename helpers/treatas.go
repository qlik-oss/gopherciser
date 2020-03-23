package helpers

// Implement this interface to override the type a struct is treated as by the GUI
type TreatAs interface {
	/*
		Supported types:
			String
			Int
			Float
			Enum
			Bool
			Slice
			SliceElement
			StringMap
			StringMapElement
	*/
	TreatAs() string
}
