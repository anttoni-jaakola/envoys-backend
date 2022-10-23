package help

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
