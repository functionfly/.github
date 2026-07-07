package notification

import (
	"encoding/json"
	"fmt"
	"time"
)

type SlackSeverity string

const (
	SeverityCritical  SlackSeverity = "critical"
	SeverityHigh      SlackSeverity = "high"
	SeverityMedium    SlackSeverity = "medium"
	SeverityLow       SlackSeverity = "low"
	SeverityInfo      SlackSeverity = "info"
	SeverityMaintenance SlackSeverity = "maintenance"
)

type SlackPayload struct {
	Blocks []SlackBlock `json:"blocks,omitempty"`
}

type SlackBlock struct {
	Type      string      `json:"type"`
	Text      *SlackText  `json:"text,omitempty"`
	Fields    []SlackText `json:"fields,omitempty"`
	Elements  []SlackElement `json:"elements,omitempty"`
}

type SlackText struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type SlackElement struct {
	Type     string `json:"type"`
	Text     *SlackText `json:"text,omitempty"`
	URL      string `json:"url,omitempty"`
	ActionID string `json:"action_id,omitempty"`
	Value    string `json:"value,omitempty"`
}

func BuildSlackPayload(n *Notification, severity string) SlackPayload {
	emoji, _ := getSeverityEmojiAndColor(severity)
	
	header := fmt.Sprintf("%s %s", emoji, n.Title)
	if len(header) > 150 {
		header = header[:147] + "..."
	}

	var fields []SlackText
	if n.Body != "" {
		fields = append(fields, SlackText{
			Type: "mrkdwn",
			Text: fmt.Sprintf("*Status:*\n%s", n.Body),
		})
	}

	uptime := ""
	if uptimeVal, ok := n.Data["uptime"].(float64); ok {
		uptime = fmt.Sprintf("%.1f%%", uptimeVal)
	}
	if uptime != "" {
		fields = append(fields, SlackText{
			Type: "mrkdwn",
			Text: fmt.Sprintf("*Uptime:*\n%s", uptime),
		})
	}

	responseTime := ""
	if rt, ok := n.Data["response_time"].(float64); ok {
		responseTime = fmt.Sprintf("%.0fms", rt)
	}
	if responseTime != "" {
		fields = append(fields, SlackText{
			Type: "mrkdwn",
			Text: fmt.Sprintf("*Response Time:*\n%s", responseTime),
		})
	}

	detected := fmt.Sprintf("<!date^%d^{date_short_pretty} at {time}|%s>", 
		n.CreatedAt.Unix(), n.CreatedAt.Format(time.RFC3339))
	fields = append(fields, SlackText{
		Type: "mrkdwn",
		Text: fmt.Sprintf("*Detected:*\n%s", detected),
	})

	blocks := []SlackBlock{
		{
			Type: "header",
			Text: &SlackText{
				Type: "plain_text",
				Text: header,
			},
		},
	}

	if len(fields) > 0 {
		blocks = append(blocks, SlackBlock{
			Type:   "section",
			Fields: fields,
		})
	}

	blocks = append(blocks, SlackBlock{
		Type: "actions",
		Elements: []SlackElement{
			{
				Type: "button",
				Text: &SlackText{
					Type: "plain_text",
					Text: "View Status",
				},
				URL: "https://status.functionfly.com",
			},
		},
	})

	return SlackPayload{Blocks: blocks}
}

func getSeverityEmojiAndColor(severity string) (emoji string, color string) {
	switch severity {
	case "critical":
		return "🔴", "#FF0000"
	case "high":
		return "🟠", "#FF6600"
	case "medium":
		return "🟡", "#FFCC00"
	case "low":
		return "🔵", "#0066FF"
	case "info", "recovery":
		return "🟢", "#00CC00"
	case "maintenance":
		return "🔵", "#0066FF"
	default:
		return "🟡", "#FFCC00"
	}
}

func BuildStatusReportPayload(components []ComponentStatus, period string) SlackPayload {
	var blocks []SlackBlock

	periodText := "24 Hours"
	if period == "7d" {
		periodText = "7 Days"
	} else if period == "30d" {
		periodText = "30 Days"
	}

	blocks = append(blocks, SlackBlock{
		Type: "header",
		Text: &SlackText{
			Type: "plain_text",
			Text: fmt.Sprintf("📊 Platform Status Report — %s", periodText),
		},
	})

	for _, comp := range components {
		status := getComponentStatus(comp)
		blocks = append(blocks, SlackBlock{
			Type: "section",
			Text: &SlackText{
				Type: "mrkdwn",
				Text: fmt.Sprintf("%s *%s*\n   Uptime: %.2f%% | Response: %dms",
					getStatusEmoji(status), comp.Name, comp.Uptime, comp.ResponseTime),
			},
		})
	}

	return SlackPayload{Blocks: blocks}
}

type ComponentStatus struct {
	Name         string  `json:"name"`
	Status       string  `json:"status"`
	Uptime       float64 `json:"uptime"`
	ResponseTime int     `json:"response_time"`
}

func getComponentStatus(comp ComponentStatus) string {
	return comp.Status
}

func getStatusEmoji(status string) string {
	switch status {
	case "operational", "healthy":
		return "🟢"
	case "degraded", "slow":
		return "🟡"
	case "down", "major_outage":
		return "🔴"
	case "maintenance":
		return "🔵"
	default:
		return "⚪"
	}
}

func BuildIncidentPayload(incident *Incident) SlackPayload {
	var blocks []SlackBlock

	severity := incident.Severity
	if severity == "" {
		severity = "medium"
	}

	emoji, _ := getSeverityEmojiAndColor(severity)
	title := fmt.Sprintf("%s *%s* — %s", emoji, incident.Title, incident.Status)

	blocks = append(blocks, SlackBlock{
		Type: "header",
		Text: &SlackText{
			Type: "plain_text",
			Text: title,
		},
	})

	if incident.Description != "" {
		blocks = append(blocks, SlackBlock{
			Type: "section",
			Text: &SlackText{
				Type: "mrkdwn",
				Text: incident.Description,
			},
		})
	}

	if len(incident.AffectedComponents) > 0 {
		blocks = append(blocks, SlackBlock{
			Type: "section",
			Text: &SlackText{
				Type: "mrkdwn",
				Text: fmt.Sprintf("*Affected Components:*\n%s", joinStrings(incident.AffectedComponents, ", ")),
			},
		})
	}

	return SlackPayload{Blocks: blocks}
}

type Incident struct {
	ID                 string   `json:"id"`
	Title              string   `json:"title"`
	Description        string   `json:"description"`
	Severity           string   `json:"severity"`
	Status             string   `json:"status"`
	AffectedComponents []string `json:"affected_components"`
	CreatedAt          time.Time `json:"created_at"`
}

func joinStrings(strs []string, sep string) string {
	result := ""
	for i, s := range strs {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}

func (p *SlackPayload) ToJSON() ([]byte, error) {
	return json.Marshal(p)
}
