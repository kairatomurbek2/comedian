package config

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

// Config struct used for configuration of app with env variables
type Config struct {
	SlackToken         string `envconfig:"SLACK_TOKEN" required:"true"`
	DatabaseURL        string `envconfig:"DATABASE" required:"true" default:"comedian:comedian@/comedian?parseTime=true"`
	HTTPBindAddr       string `envconfig:"HTTP_BIND_ADDR" required:"true" default:"0.0.0.0:8080"`
	NotifierInterval   int    `envconfig:"NOTIFIER_INTERVAL" required:"true" default:2`
	ManagerSlackUserID string `envconfig:"MANAGER_SLACK_USER_ID" required:"true"`
	ReportTime         string `envconfig:"REPORT_TIME" required:"true" default:"13:05"`
	Language           string `envconfig:"LANGUAGE" required:"true" default:"en_US"`
	CollectorURL       string `envconfig:"COLLECTOR_URL" required:"true"`
	CollectorToken     string `envconfig:"COLLECTOR_TOKEN" required:"true"`
	ChanGeneral        string `envconfig:"MANAGER_SLACK_CHAN_GENERAL" required:"true"`
	ReminderRepeatsMax int    `envconfig:"REMINDER_REPEATS_MAX" required:"true" default:5`
	ReminderTime       int64  `envconfig:"REMINDER_TIME" required:"true" default:5`
	Translate          Translate
	Debug              bool
}

// Get method processes env variables and fills Config struct
func Get() (Config, error) {
	var c Config
	err := envconfig.Process("comedian", &c)
	if err != nil {
		return c, err
	}
	t, err := GetTranslation(c.Language)
	if err != nil {
		return c, err
	}
	c.Translate = t
	fmt.Println(c.Translate)
	return c, nil
}
