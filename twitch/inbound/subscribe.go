package inbound

type SubTags struct {
	BadgeInfo                  string `mapstructure:"badge-info" json:"badge_info,omitempty"`
	Badges                     string `mapstructure:"badges" json:"badges,omitempty"`
	Color                      string `mapstructure:"color" json:"color,omitempty"`
	DisplayName                string `mapstructure:"display-name" json:"display_name,omitempty"`
	Id                         string `mapstructure:"id" json:"id,omitempty"`
	Login                      string `mapstructure:"login" json:"login,omitempty"`
	Mod                        string `mapstructure:"mod" json:"mod,omitempty"`
	MsgId                      string `mapstructure:"msg-id" json:"msg_id,omitempty"`
	MsgParamCumulativeMonths   string `mapstructure:"msg-param-cumulative-months" json:"msg_param_cumulative_months,omitempty"`
	MsgParamMonths             string `mapstructure:"msg-param-months" json:"msg_param_months,omitempty"`
	MsgParamMultiMonthDuration string `mapstructure:"msg-param-multimonth-duration" json:"msg_param_multi_month_duration,omitempty"`
	MsgParamMultiMonthTenure   string `mapstructure:"msg-param-multimonth-tenure" json:"msg_param_multi_month_tenure,omitempty"`
	MsgParamShouldShareStreak  string `mapstructure:"msg-param-should-share-streak" json:"msg_param_should_share_streak,omitempty"`
	MsgParamSubPlan            string `mapstructure:"msg-param-sub-plan" json:"msg_param_sub_plan,omitempty"`
	MsgParamSubPlanName        string `mapstructure:"msg-param-sub-plan-name" json:"msg_param_sub_plan_name,omitempty"`
	MsgParamWasGifted          string `mapstructure:"msg-param-was-gifted" json:"msg_param_was_gifted,omitempty"`
	RoomId                     string `mapstructure:"room-id" json:"room_id,omitempty"`
	Subscriber                 string `mapstructure:"subscriber" json:"subscriber,omitempty"`
	SystemMsg                  string `mapstructure:"system-msg" json:"system_msg,omitempty"`
	TmiSentTs                  string `mapstructure:"tmi-sent-ts" json:"tmi_sent_ts,omitempty"`
	UserId                     string `mapstructure:"user-id" json:"user_id,omitempty"`
}

type GiftSubTags struct {
	BadgeInfo                    string `mapstructure:"badge-info"`
	Badges                       string `mapstructure:"badges"`
	DisplayName                  string `mapstructure:"display-name"`
	Id                           string `mapstructure:"id"`
	Login                        string `mapstructure:"login"`
	Mod                          string `mapstructure:"mod"`
	MsgId                        string `mapstructure:"msg-id"`
	MsgParamGiftMonths           string `mapstructure:"msg-param-gift-months"`
	MsgParamMonths               string `mapstructure:"msg-param-months"`
	MsgParamOriginId             string `mapstructure:"msg-param-origin-id"`
	MsgParamRecipientDisplayName string `mapstructure:"msg-param-recipient-display-name"`
	MsgParamRecipientId          string `mapstructure:"msg-param-recipient-id"`
	MsgParamRecipientUserName    string `mapstructure:"msg-param-recipient-user-name"`
	MsgParamSenderCount          string `mapstructure:"msg-param-sender-count"`
	MsgParamSubPlan              string `mapstructure:"msg-param-sub-plan"`
	MsgParamSubPlanName          string `mapstructure:"msg-param-sub-plan-name"`
	RoomId                       string `mapstructure:"room-id"`
	Subscriber                   string `mapstructure:"subscriber"`
	SystemMsg                    string `mapstructure:"system-msg"`
	TmiSentTs                    string `mapstructure:"tmi-sent-ts"`
	UserId                       string `mapstructure:"user-id"`
}
