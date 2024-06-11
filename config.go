package main

// Discord Config

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
		Channels: []Channel{{"selling", "427630953100476436"}, {"buying", "427630933131395103"}},
		Enabled:  true,
	},
	{
		Name:     "Canadian Mechanical Keyboards",
		Channels: []Channel{{"selling", "379790816568410112"}},
		Enabled:  true,
	},
	{
		Name:     "Mechfeed",
		Channels: []Channel{{"mechfeed", "1113657844856782868"}},
		Enabled:  true,
	},
}