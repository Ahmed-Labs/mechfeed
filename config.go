package main

// Discord Config

type Server struct {
	Name     string
	Channels []Channel
}

type Channel struct {
	Name string
	ID   string
}

var ServerList = []Server{
	{
		Name: "MechMarket",
		Channels: []Channel{
			{"selling", "427630953100476436"},
			{"buying", "427630933131395103"},
			{"selling-artisans", "539626107176091648"},
			{"buying-artisans", "539626087303348254"},
			{"gb-spot-selling-trading", "507695814223724546"},
		},
	},
	{
		Name: "Canadian Mechanical Keyboards",
		Channels: []Channel{
			{"market", "379790816568410112"},
		},
	},
	{
		Name: "MechGroupBuys",
		Channels: []Channel{
			{"sell", "683068530513674240"},
			{"buy", "683068253257596939"},
		},
	},
	{
		Name: "Deskhero.ca",
		Channels: []Channel{
			{"market", "689578698436771858"},
		},
	},
	{
		Name: "Top Clack",
		Channels: []Channel{
			{"mechmarket", "371096775467073536"},
		},
	},
	{
		Name: "Artisan Trading",
		Channels: []Channel{
			{"general-selling", "783779301328683028"},
			{"general-buying", "783779330315911190"},
			{"general-trading", "783777689063653396"},
			{"metal-novelties", "788549951428100136"},
		},
	},
	// {
	// 	Name:     "Mechfeed",
	// 	Channels: []Channel{{"mechfeed", "968791988465983518"}},
	// },
}