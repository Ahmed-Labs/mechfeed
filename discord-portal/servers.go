package discordportal

type Server struct {
	Name     string
	Channels []Channel
	Enabled  bool
}

type Channel struct {
	Name string
	ID   string
}

var ServerList = []Server{
	{
		Name:     "MechMarket",
		Channels: []Channel{{"selling", "427630953100476436"}},
		Enabled:  true,
	},
	{
		Name:     "MechMarket",
		Channels: []Channel{{"selling", "427630953100476436"}},
		Enabled:  true,
	},
}


