package models

type Product struct {
	ProductName  string `json:"productName"`
	Urllabel     string `json:"urllabel"`
	Emoji        string `json:"logo"`
	ChannelID    string `json:"channelid"`
	SpaceID      string `json:"spaceid"`
	Docslink     string `json:"docslink"`
	Shoplink     string `json:"shoplink"`
	Payloadslink string `json:"payloadslink"`
	Githublink   string `json:"githublink"`
	InviteLink   string `json:"invitelink"`
	SupportLink  string `json:"supportlink"`
}

var GITBOOKAPI = "https://api.gitbook.com/v1/spaces/"
var SEARCHENDPOINT = "/search/ask"

var Default = Product{
	ProductName:  "",
	Urllabel:     "",
	Emoji:        "<:hak5:1063661436187975701>",
	ChannelID:    "824326326076702770",
	SpaceID:      "",
	Docslink:     "https://docs.hak5.org",
	Shoplink:     "https://shop.hak5.org",
	Payloadslink: "https://payloadhub.com",
	Githublink:   "https://github.com/hak5",
	InviteLink:   "https://discord.gg/hak5",
	SupportLink:  "https://hak5.org/support",
}

var Ducky = Product{
	ProductName:  "USB Rubber Ducky",
	Urllabel:     "usb-rubber-ducky",
	Emoji:        "<:Rubber_Ducky:1063661661795385384>",
	ChannelID:    "522275837651714048",
	SpaceID:      "-MiIkRK_o3RBhZzUkrzr",
	Docslink:     "https://docs.hak5.org/hak5-usb-rubber-ducky/",
	Shoplink:     "https://shop.hak5.org/products/usb-rubber-ducky",
	Payloadslink: "https://hak5.org/blogs/payloads/tagged/usb-rubber-ducky",
	Githublink:   "https://github.com/hak5/usbrubberducky-payloads",
	InviteLink:   "https://discord.gg/HMH5yyKuuA",
	SupportLink:  "https://hak5.org/support/ducky",
}

var PayloadStudio = Product{
	ProductName:  "PayloadStudio",
	Urllabel:     Default.Urllabel,
	Emoji:        "<:payload_studio:1053210968064262234>",
	ChannelID:    "1006233482957164634",
	SpaceID:      "RgTCQkzfO7AUWTT3gFAq",
	Docslink:     "https://docs.hak5.org/payload-studio/",
	Shoplink:     "https://shop.hak5.org/products/payload-studio-pro",
	Payloadslink: Default.Payloadslink,
	Githublink:   Default.Githublink,
	InviteLink:   "https://discord.gg/AGBSbUdfGx",
	SupportLink:  "https://hak5.org/support/payloadstudio",
}
var CloudC2 = Product{
	ProductName:  "Cloud C2",
	Urllabel:     Default.Urllabel,
	Emoji:        "<:Cloud_C2:1063661578181943336>",
	ChannelID:    "522276096746717184",
	SpaceID:      "<:Cloud_C2:1063661578181943336>",
	Docslink:     "https://docs.hak5.org/cloud-c2/",
	Shoplink:     "https://shop.hak5.org/products/cloud-c2",
	Payloadslink: Default.Payloadslink,
	Githublink:   Default.Githublink,
	InviteLink:   "https://discord.gg/gFBhAEXycP",
	SupportLink:  Default.SupportLink,
}
var Crab = Product{
	ProductName:  "Screen Crab",
	Urllabel:     Default.Urllabel,
	Emoji:        "<:Screen_Crab:1063661831756988427>",
	ChannelID:    "608057573517557809",
	SpaceID:      "-MiWySN4BHDJlUatEfm3",
	Docslink:     "https://docs.hak5.org/screen-crab/",
	Shoplink:     "https://shop.hak5.org/products/screen-crab",
	Payloadslink: Default.Payloadslink,
	Githublink:   Default.Githublink,
	InviteLink:   "https://discord.gg/mxdb7x2DqD",
	SupportLink:  "https://hak5.org/support/crab",
}
var Coconut = Product{
	ProductName:  "WiFi Coconut",
	Urllabel:     Default.Urllabel,
	Emoji:        "<:WiFi_Coconut:1063661830309953606>",
	ChannelID:    "1007363521098551349",
	SpaceID:      "DkilLranx3TNqKgzbBZ9",
	Docslink:     "https://docs.hak5.org/wifi-coconut/",
	Shoplink:     "https://shop.hak5.org/products/wifi-coconut",
	Payloadslink: Default.Payloadslink,
	Githublink:   "https://github.org/hak5/wifi-coconut",
	InviteLink:   Default.InviteLink,
	SupportLink:  "https://hak5.org/support/coconut",
}
var Pineapple = Product{
	ProductName:  "WiFi Pineapple",
	Urllabel:     Default.Urllabel,
	Emoji:        "<:WiFi_Pineapple:1063661628207411251>",
	ChannelID:    "522275782731497473",
	SpaceID:      "-Mhuhsyl_byoEWXOc5EU",
	Docslink:     "https://docs.hak5.org/wifi-pineapple/",
	Shoplink:     "https://shop.hak5.org/products/wifi-pineapple",
	Payloadslink: Default.Payloadslink,
	Githublink:   "https://github.com/hak5/pineapple-modules",
	InviteLink:   "https://discord.gg/WEbS5DY6Ar",
	SupportLink:  "https://hak5.org/support/pineapple",
}
var Croc = Product{
	ProductName:  "Key Croc",
	Urllabel:     "key-croc",
	Emoji:        "<:Key_Croc:1063661834093199431>",
	ChannelID:    "709235715120168970",
	SpaceID:      "-MhLOzjhonMdC6SLKqRt",
	Docslink:     "https://docs.hak5.org/key-croc/",
	Shoplink:     "https://shop.hak5.org/products/key-croc",
	Payloadslink: "https://hak5.org/blogs/payloads/tagged/key-croc",
	Githublink:   "https://github.com/hak5/keycroc-payloads",
	InviteLink:   "https://discord.gg/rgq5ST86gW",
	SupportLink:  "https://hak5.org/support/croc",
}
var Bunny = Product{
	ProductName:  "Bash Bunny",
	Urllabel:     "bash-bunny",
	Emoji:        "<:Bash_Bunny:1063661828753858680>",
	ChannelID:    "1009919913877590146",
	SpaceID:      "nxJgJ9UdPfrcuL1U8DpL",
	Docslink:     "https://docs.hak5.org/bash-bunny/",
	Shoplink:     "https://shop.hak5.org/products/bash-bunny",
	Payloadslink: "https://hak5.org/blogs/payloads/tagged/bash-bunny",
	Githublink:   "https://github.com/hak5/bashbunny-payloads",
	InviteLink:   "https://discord.gg/jF4VrbEwVW",
	SupportLink:  "https://hak5.org/support/bunny",
}
var Squirrel = Product{
	ProductName:  "Packet Squirrel",
	Urllabel:     "packet-squirrel",
	Emoji:        "<:packet_squirrel:1063661837398315079>",
	ChannelID:    "522275913031745548",
	SpaceID:      "-MiX86qQFvjhg-a6RwWr",
	Docslink:     "https://docs.hak5.org/packet-squirrel",
	Shoplink:     "https://shop.hak5.org/products/packet-squirrel",
	Payloadslink: "https://hak5.org/blogs/payloads/tagged/packet-squirrel",
	Githublink:   "https://github.com/hak5/packetsquirrel-payloads",
	InviteLink:   "https://discord.gg/ggW5DR7tpz",
	SupportLink:  "https://hak5.org/support/squirrel",
}
var Turtle = Product{
	ProductName:  "LAN Turtle",
	Urllabel:     "lan-turtle",
	Emoji:        "<:lan_turtle:1063661838937620480>",
	ChannelID:    "522275913031745548",
	SpaceID:      "N8g6UIGOyv4mW7rOOuC8",
	Docslink:     "https://docs.hak5.org/lan-turtle",
	Shoplink:     "https://shop.hak5.org/products/lan-turtle",
	Payloadslink: "https://hak5.org/blogs/payloads/tagged/lan-turtle",
	Githublink:   "https://github.com/hak5/lanturtle-modules",
	InviteLink:   "https://discord.gg/ggW5DR7tpz",
	SupportLink:  "https://hak5.org/support/turtle",
}
var Shark = Product{
	ProductName:  "Shark Jack",
	Urllabel:     "shark-jack",
	Emoji:        "<:Shark_Jack:1063661835888381952>",
	ChannelID:    "610344558655438858",
	SpaceID:      "-MhxW6geyenAGJvaKW11",
	Docslink:     "https://docs.hak5.org/shark-jack",
	Shoplink:     "https://shop.hak5.org/products/shark-jack",
	Payloadslink: "https://hak5.org/blogs/payloads/tagged/shark-jack",
	Githublink:   "https://github.com/hak5/sharkjack-payloads",
	InviteLink:   "https://discord.gg/CsnaGN2CEd",
	SupportLink:  "https://hak5.org/support/shark",
}
var OMG = Product{
	ProductName:  "O.MG Devices",
	Urllabel:     "omg",
	Emoji:        "<:omg:1063696145022468116>",
	ChannelID:    "953129985131040838",
	SpaceID:      Default.SpaceID,
	Docslink:     "https://docs.hak5.org/hak5-docs/#omg",
	Shoplink:     "https://shop.hak5.org/collections/mischief-gadgets",
	Payloadslink: "https://hak5.org/blogs/payloads/tagged/omg",
	Githublink:   "https://github.com/hak5/omg-payloads",
	InviteLink:   "https://discord.gg/Sw2pzaHT6N",
	SupportLink:  "https://hak5.org/support/omg",
}
