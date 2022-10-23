package help

import ua "github.com/mileusna/useragent"

func MetaAgent(meta string) *ua.UserAgent {

	agent := ua.Parse(meta)
	if len(agent.Device) == 0 {
		if agent.Mobile {
			agent.Device = "mobile"
		}
		if agent.Tablet {
			agent.Device = "tablet"
		}
		if agent.Desktop {
			agent.Device = "desktop"
		}
		if agent.Bot {
			agent.Device = "bot"
		}
	}

	if len(agent.Device) == 0 {
		agent.Device = "unknown"
	}

	return &agent
}
