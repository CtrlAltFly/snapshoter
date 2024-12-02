package main

import (
	"fmt"
	"github.com/spf13/viper"
	"grafana-snapshoter/grafana"
	"grafana-snapshoter/slack"
	"log"
)

func main() {
	viper.SetConfigName("config")
	viper.AddConfigPath("/tap/config")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	viper.SetConfigName("secrets")
	viper.AddConfigPath("/tap/secrets")
	err = viper.MergeInConfig()
	if err != nil {
		log.Fatalf("Error reading secrets file: %v", err)
	}
	grafanaToken := viper.GetString("grafana.api_token")

	slackWebhook := viper.GetString("slack.webhook_url")

	grafanaURL := viper.GetString("grafana.url")
	timeRangeHours := viper.GetInt("snapshot.time_range_hours")
	if timeRangeHours == 0 {
		timeRangeHours = 24
	}

	dashboards := viper.GetStringSlice("snapshot.dashboards")

	for _, dashboard := range dashboards {
		log.Printf("Processing snapshot for dashboard: %s", dashboard)

		snapshotURL, err := grafana.ClickSnapshot(grafanaURL, grafanaToken, dashboard, timeRangeHours)
		if err != nil {
			log.Printf("Error creating snapshot for dashboard '%s': %v", dashboard, err)
			continue
		}

		message := formatSlackMessage(dashboard, timeRangeHours, snapshotURL)
		err = slack.SendSlackMessage(slackWebhook, message)
		if err != nil {
			log.Printf("Error sending Slack message for dashboard '%s': %v", dashboard, err)
		} else {
			log.Printf("Successfully sent snapshot for dashboard '%s' to Slack", dashboard)
		}
	}

	log.Println("All requested dashboards snap-shoted and sent to slack successfully.")
}

func formatSlackMessage(dashboard string, hours int, snapshotURL string) string {
	return fmt.Sprintf("Dashboard Snapshot for `%s` (Last %d Hours):\n%s", dashboard, hours, snapshotURL)
}
