package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/resend/resend-go/v2"
	"github.com/robfig/cron/v3"
)

var resendAPIKey = "re_BvTE5Zor_74x57SLrPu7AUFuahwFMr9hU"

type EmailRequest struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Subject string `json:"subject"`
	Text    string `json:"text"`
}

func Invitation(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req struct {
		Email     string `json:"email"`
		MeetLink  string `json:"meetLink"`
		EventTime string `json:"eventTime"` // Expected format: RFC3339
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := sendEmail(req.Email, "Has sido invitado al webinar presentación de TalentHub. Por favor has click aquí: "+req.MeetLink); err != nil {
		http.Error(w, "Unable to send email: "+err.Error(), http.StatusInternalServerError)
		return
	}

	addCronJob(req.Email, req.MeetLink, req.EventTime)

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Invitation sent to: %s\n", req.Email)
}

func sendEmail(to string, body string) error {
	client := resend.NewClient(resendAPIKey)

	params := &resend.SendEmailRequest{
		From:    "onboarding@resend.dev",
		To:      []string{to},
		Subject: "TalentHub Webinar Invitación",
		Html:    body,
	}

	_, err := client.Emails.Send(params)
	if err != nil {
		return fmt.Errorf("failed to send email: %s", err.Error())
	}

	return nil
}

func addCronJob(email string, meetLink string, eventTime string) {
	c := cron.New()
	eventTimeParsed, err := time.Parse(time.RFC3339, eventTime)
	if err != nil {
		log.Fatalf("Unable to parse event time: %v", err)
	}
	reminderTime := eventTimeParsed.AddDate(0, 0, -1)
	reminderSpec := fmt.Sprintf("%d %d %d %d %d ?", reminderTime.Second(), reminderTime.Minute(), reminderTime.Hour(), reminderTime.Day(), reminderTime.Month())

	c.AddFunc(reminderSpec, func() {
		err := sendEmail(email, "Este es un recordatorio para unirte al Webinar de TalentHub. Únete aquí: "+meetLink)
		if err != nil {
			log.Fatalf("Unable to send reminder email: %v", err)
		}
	})
	c.Start()
}
