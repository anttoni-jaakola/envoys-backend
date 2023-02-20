package help

// Depth - This function creates an array of strings that represent different depths of data. The depths are 0, 60, 300, 900,
// 1800, 3600, and 86400. This list could be used to set the depth of data requested from a particular source.
func Depth() []string {
	return []string{
		"0",
		"60",
		"300",
		"900",
		"1800",
		"3600",
		"86400",
	}
}

// Resolution - The purpose of this function is to return a string that describes the resolution of a given depth. The depth is given
// as a string, and the function uses a switch statement to determine the resolution of the depth. The function will
// return a string corresponding to the resolution, or a default of "15 minutes" if no resolution is specified.
func Resolution(depth string) string {

	switch depth {
	case "1", "60":
		return "1 minute"
	case "5", "300":
		return "5 minutes"
	case "15", "900":
		return "15 minutes"
	case "30", "1800":
		return "30 minutes"
	case "1h", "3600":
		return "1 hour"
	case "1D", "86400":
		return "1 day"
	}

	return "15 minutes"
}
