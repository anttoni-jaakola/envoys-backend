package help

import ua "github.com/mileusna/useragent"

// MetaAgent - This function is used to create a UserAgent object from a meta string. It determines the type of device associated
// with the meta string and sets the device type accordingly.
func MetaAgent(meta string) *ua.UserAgent {

	// The purpose of this statement is to create a new user agent (ua) object based on the metadata specified in the
	// "meta" variable. This object can then be used to retrieve information about the user agent, such as its type,
	// version, etc.
	agent := ua.Parse(meta)

	// This switch statement is used to determine the type of device the user is using. If the length of the agent.Device is
	// 0, then the code will switch through the different options (mobile, tablet, desktop, and bot) to determine the type
	// of device and assign it to the device variable. If none of these cases are true, then the device is set to "unknown".
	switch {
	case len(agent.Device) == 0:
		switch {
		case agent.Mobile:
			agent.Device = "mobile"
		case agent.Tablet:
			agent.Device = "tablet"
		case agent.Desktop:
			agent.Device = "desktop"
		case agent.Bot:
			agent.Device = "bot"
		default:
			agent.Device = "unknown"
		}
	}

	return &agent
}
