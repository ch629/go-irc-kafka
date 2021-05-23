package inbound

type SubTags struct {
	BadgeInfo          string `mapstructure:"badge-info" json:"badge_info,omitempty"`
	Badges             string `mapstructure:"badges" json:"badges,omitempty"`
	Color              string `mapstructure:"color" json:"color,omitempty"`
	DisplayName        string `mapstructure:"display-name" json:"display_name,omitempty"`
	Id                 string `mapstructure:"id" json:"id,omitempty"`
	Login              string `mapstructure:"login" json:"login,omitempty"`
	Mod                string `mapstructure:"mod" json:"mod,omitempty"`
	MsgId              string `mapstructure:"msg-id" json:"msg_id,omitempty"`
	CumulativeMonths   string `mapstructure:"msg-param-cumulative-months" json:"msg_param_cumulative_months,omitempty"`
	Months             string `mapstructure:"msg-param-months" json:"msg_param_months,omitempty"`
	MultiMonthDuration string `mapstructure:"msg-param-multimonth-duration" json:"msg_param_multi_month_duration,omitempty"`
	MultiMonthTenure   string `mapstructure:"msg-param-multimonth-tenure" json:"msg_param_multi_month_tenure,omitempty"`
	ShouldShareStreak  string `mapstructure:"msg-param-should-share-streak" json:"msg_param_should_share_streak,omitempty"`
	SubPlan            string `mapstructure:"msg-param-sub-plan" json:"msg_param_sub_plan,omitempty"`
	SubPlanName        string `mapstructure:"msg-param-sub-plan-name" json:"msg_param_sub_plan_name,omitempty"`
	WasGifted          string `mapstructure:"msg-param-was-gifted" json:"msg_param_was_gifted,omitempty"`
	RoomId             string `mapstructure:"room-id" json:"room_id,omitempty"`
	Subscriber         string `mapstructure:"subscriber" json:"subscriber,omitempty"`
	SystemMsg          string `mapstructure:"system-msg" json:"system_msg,omitempty"`
	TmiSentTs          string `mapstructure:"tmi-sent-ts" json:"tmi_sent_ts,omitempty"`
	UserId             string `mapstructure:"user-id" json:"user_id,omitempty"`
}

type GiftSubTags struct {
	BadgeInfo            string `mapstructure:"badge-info"`
	Badges               string `mapstructure:"badges"`
	DisplayName          string `mapstructure:"display-name"`
	ID                   string `mapstructure:"id"`
	Login                string `mapstructure:"login"`
	Mod                  string `mapstructure:"mod"`
	MsgID                string `mapstructure:"msg-id"`
	GiftMonths           string `mapstructure:"msg-param-gift-months"`
	Months               string `mapstructure:"msg-param-months"`
	OriginID             string `mapstructure:"msg-param-origin-id"`
	RecipientDisplayName string `mapstructure:"msg-param-recipient-display-name"`
	RecipientID          string `mapstructure:"msg-param-recipient-id"`
	RecipientUserName    string `mapstructure:"msg-param-recipient-user-name"`
	SenderCount          string `mapstructure:"msg-param-sender-count"`
	SubPlan              string `mapstructure:"msg-param-sub-plan"`
	SubPlanName          string `mapstructure:"msg-param-sub-plan-name"`
	RoomId               string `mapstructure:"room-id"`
	Subscriber           string `mapstructure:"subscriber"`
	SystemMsg            string `mapstructure:"system-msg"`
	TmiSentTs            string `mapstructure:"tmi-sent-ts"`
	UserID               string `mapstructure:"user-id"`
}
