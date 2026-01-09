package templates

import (
	"time"

	"github.com/lenaxia/tinyrsvp/internal/models"
)

type InviteEmailData struct {
	Event struct {
		Title        string
		Description  string
		StartTime    time.Time
		EndTime      *time.Time
		Timezone     string
		Location     string
		RSVPDeadline *time.Time
	}
	Invite struct {
		Name  string
		Email string
	}
	RSVPURL     string
	MaxPlusOnes int
}

type RSVPPageData struct {
	Event struct {
		Title        string
		Description  string
		StartTime    time.Time
		EndTime      *time.Time
		Timezone     string
		Location     string
		RSVPDeadline *time.Time
	}
	Token       string
	MaxPlusOnes int
	Questions   []QuestionData
}

type QuestionData struct {
	ID           int64
	QuestionText string
	QuestionType string
	Options      []OptionData
	Required     bool
	HelpText     string
}

type OptionData struct {
	Value string
	Label string
}

type ConfirmationPageData struct {
	Event struct {
		Title       string
		Description string
		StartTime   time.Time
		EndTime     *time.Time
		Timezone    string
		Location    string
	}
	Token string
	RSVP  struct {
		Response string
		PlusOnes int
		Notes    string
	}
	Answers []AnswerData
}

type AnswerData struct {
	QuestionText  string
	AnswerDisplay string
}

func BuildInviteEmailData(event *models.Event, invite *models.Invite, rsvpURL string) *InviteEmailData {
	data := &InviteEmailData{
		RSVPURL:     rsvpURL,
		MaxPlusOnes: invite.MaxPlusOnes,
	}

	data.Event.Title = event.Title
	if event.Description != nil {
		data.Event.Description = *event.Description
	}
	data.Event.StartTime = event.StartTime
	data.Event.EndTime = event.EndTime
	data.Event.Timezone = event.Timezone
	if event.Location != nil {
		data.Event.Location = *event.Location
	}
	data.Event.RSVPDeadline = event.RSVPDeadline

	if invite.Name != nil {
		data.Invite.Name = *invite.Name
	}
	if invite.Email != nil {
		data.Invite.Email = *invite.Email
	}

	return data
}

func BuildRSVPPageData(event *models.Event, invite *models.Invite, questions []*models.PreferenceQuestion, token string) *RSVPPageData {
	data := &RSVPPageData{
		Token:       token,
		MaxPlusOnes: invite.MaxPlusOnes,
	}

	data.Event.Title = event.Title
	if event.Description != nil {
		data.Event.Description = *event.Description
	}
	data.Event.StartTime = event.StartTime
	data.Event.EndTime = event.EndTime
	data.Event.Timezone = event.Timezone
	if event.Location != nil {
		data.Event.Location = *event.Location
	}
	data.Event.RSVPDeadline = event.RSVPDeadline

	for _, q := range questions {
		qData := QuestionData{
			ID:           q.ID,
			QuestionText: q.QuestionText,
			QuestionType: string(q.QuestionType),
			Required:     q.Required,
		}

		if q.HelpText != nil {
			qData.HelpText = *q.HelpText
		}

		if q.Options != nil {
			options, err := q.ParseOptions()
			if err == nil && options != nil {
				for _, opt := range options {
					qData.Options = append(qData.Options, OptionData{
						Value: opt,
						Label: opt,
					})
				}
			}
		}

		data.Questions = append(data.Questions, qData)
	}

	return data
}

func BuildConfirmationPageData(event *models.Event, rsvp *models.RSVP, answers []*models.RSVPAnswer, questions map[int64]*models.PreferenceQuestion, token string) *ConfirmationPageData {
	data := &ConfirmationPageData{
		Token: token,
	}

	data.Event.Title = event.Title
	if event.Description != nil {
		data.Event.Description = *event.Description
	}
	data.Event.StartTime = event.StartTime
	data.Event.EndTime = event.EndTime
	data.Event.Timezone = event.Timezone
	if event.Location != nil {
		data.Event.Location = *event.Location
	}

	data.RSVP.Response = string(rsvp.Response)
	data.RSVP.PlusOnes = rsvp.PlusOnes

	for _, ans := range answers {
		ansData := AnswerData{
			AnswerDisplay: formatAnswer(ans),
		}
		
		if q, ok := questions[ans.QuestionID]; ok {
			ansData.QuestionText = q.QuestionText
		}
		
		data.Answers = append(data.Answers, ansData)
	}

	return data
}

func formatAnswer(answer *models.RSVPAnswer) string {
	if answer.AnswerText != nil {
		return *answer.AnswerText
	}
	if answer.AnswerOption != nil {
		return *answer.AnswerOption
	}
	if answer.AnswerBoolean != nil {
		if *answer.AnswerBoolean {
			return "Yes"
		}
		return "No"
	}
	return ""
}
