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
	CommandKey string `json:"command_key"`

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

	// Auto Moderation. The *_enforce flags default to false, which is dry-run
	// mode: actions are reported to the moderator channel but not taken.
	SpamDetection             bool   `json:"spam_detection"`
	SpamEnforce               bool   `json:"spam_enforce"`
	SpamWindowSeconds         int    `json:"spam_window_seconds"`
	SpamDuplicateThreshold    int    `json:"spam_duplicate_threshold"`
	SpamCrossChannelThreshold int    `json:"spam_cross_channel_threshold"`
	SpamTimeoutMinutes        int    `json:"spam_timeout_minutes"`
	FreeBansChannel           string `json:"free_bans_channel"`
	FreeBansEnforce           bool   `json:"free_bans_enforce"`

	ConfigPath string

	FilterLanguage       bool
	JoinFloodThreshold   int
	UserMessageThreshold int
	MaxUserWarnings      int
	MaxUserKicks         int
	EnforceMode          bool
}
