package models

type ColorRole struct {
	Name string `json:"color"`
	ID   string `json:"id"`
}

type CustomRole struct {
	Name      string `json:"role_name"`
	ShortName string `json:"short_name"`
	ID        string `json:"id"`
}

type CommandRestriction struct {
	Command  string   `json:"command"`
	Mode     string   `json:"mode"`
	Channels []string `json:"channels"`
}

type Configuration struct {
	GuildID   string `json:"guild_id"`
	GuildName string `json:"guild_name"`

	StatusText  string `json:"status_text"`
	WelcomeText string `json:"welcome_text"`
	RulesText   string `json:"rules_text"`

	AdminRoles  []string `json:"admin_roles"`
	JoinRoleIDs []string `json:"join_role_ids"`

	CommandRestrictions []CommandRestriction `json:"command_restrictions"`

	ColorRoles  []ColorRole  `json:"color_roles"`
	CustomRoles []CustomRole `json:"custom_roles"`

	IgnoredUsers []string `json:"ignored_users"`

	LoggingChannel   string `json:"logging_channel"`
	ModeratorChannel string `json:"moderator_channel"`
	ModeratorRoleID  string `json:"moderator_role_id"`

	// Message Content is a privileged intent; enable this only once Discord has approved it for the app,
	// otherwise the gateway rejects the connection (close 4014). Spam detection and deleted message
	// logging need it to see message text.
	MessageContentIntent bool `json:"message_content_intent"`

	// Auto Moderation. The *_enforce flags default to false, which is dry-run
	// mode: actions are reported to the moderator channel but not taken.
	SpamDetection             bool `json:"spam_detection"`
	SpamEnforce               bool `json:"spam_enforce"`
	SpamWindowSeconds         int  `json:"spam_window_seconds"`
	SpamDuplicateThreshold    int  `json:"spam_duplicate_threshold"`
	SpamCrossChannelThreshold int  `json:"spam_cross_channel_threshold"`
	SpamTimeoutMinutes        int  `json:"spam_timeout_minutes"`

	// Burst detection flags users posting in many channels within a short window, whatever
	// the messages say, so it works without message_content_intent. Shares spam_timeout_minutes.
	BurstDetection        bool `json:"burst_detection"`
	BurstEnforce          bool `json:"burst_enforce"`
	BurstWindowSeconds    int  `json:"burst_window_seconds"`
	BurstChannelThreshold int  `json:"burst_channel_threshold"`

	FreeBansChannel string `json:"free_bans_channel"`
	FreeBansEnforce bool   `json:"free_bans_enforce"`

	// YouTube feed posts new videos, shorts and live streams from youtube_channel_id (a UC... ID)
	// to youtube_feed_channel. Live detection needs the SCUZZY_YOUTUBE_API_KEY environment variable.
	// youtube_state_file defaults to youtube_seen.json next to the config file.
	YoutubeChannelID   string `json:"youtube_channel_id"`
	YoutubeFeedChannel string `json:"youtube_feed_channel"`
	YoutubePollMinutes int    `json:"youtube_poll_minutes"`
	YoutubeStateFile   string `json:"youtube_state_file"`

	ConfigPath string

	FilterLanguage       bool
	JoinFloodThreshold   int
	UserMessageThreshold int
	MaxUserWarnings      int
	MaxUserKicks         int
	EnforceMode          bool
}
