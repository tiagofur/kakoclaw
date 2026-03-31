package marketing

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/sipeed/makoclaw/pkg/storage"
)

type AutomationStore interface {
	ListPendingRuns() ([]storage.AutomationRun, error)
	ClaimRun(id int64) (bool, error)
	UpdateRunStatus(id int64, status string) error
	GetAutomationByID(id int64) (*storage.Automation, error)
	GetContactByID(id int64) (*storage.Contact, error)
	GetCampaignByID(id int64) (*storage.EmailCampaign, error)
	ListVariants(campaignID int64) ([]storage.ExperimentVariant, error)
	CreateDelivery(d *storage.EmailDelivery) (*storage.EmailDelivery, error)
}

type AutomationRunner struct {
	store             AutomationStore
	sender            *SMTPSender
	unsubscribeSecret string
	interval          time.Duration
	ctx               context.Context
	cancel            context.CancelFunc
}

func NewAutomationRunner(store AutomationStore, sender *SMTPSender, unsubscribeSecret string) *AutomationRunner {
	ctx, cancel := context.WithCancel(context.Background())
	return &AutomationRunner{
		store:             store,
		sender:            sender,
		unsubscribeSecret: unsubscribeSecret,
		interval:          1 * time.Minute,
		ctx:               ctx,
		cancel:            cancel,
	}
}

func (r *AutomationRunner) Start() {
	go func() {
		ticker := time.NewTicker(r.interval)
		defer ticker.Stop()
		for {
			select {
			case <-r.ctx.Done():
				return
			case <-ticker.C:
				r.Tick()
			}
		}
	}()
}

func (r *AutomationRunner) Stop() {
	r.cancel()
}

func (r *AutomationRunner) Tick() {
	runs, err := r.store.ListPendingRuns()
	if err != nil {
		log.Printf("[automation-runner] error listing pending runs: %v", err)
		return
	}

	for _, run := range runs {
		claimed, err := r.store.ClaimRun(run.ID)
		if err != nil || !claimed {
			continue
		}

		if err := r.processRun(run); err != nil {
			log.Printf("[automation-runner] error processing run %d: %v", run.ID, err)
			r.store.UpdateRunStatus(run.ID, "failed")
			continue
		}
		r.store.UpdateRunStatus(run.ID, "sent")
	}
}

func (r *AutomationRunner) processRun(run storage.AutomationRun) error {
	auto, err := r.store.GetAutomationByID(run.AutomationID)
	if err != nil || auto == nil {
		return fmt.Errorf("getting automation %d: %w", run.AutomationID, err)
	}

	campaign, err := r.store.GetCampaignByID(auto.CampaignID)
	if err != nil || campaign == nil {
		return fmt.Errorf("getting campaign %d: %w", auto.CampaignID, err)
	}

	contact, err := r.store.GetContactByID(run.ContactID)
	if err != nil || contact == nil {
		return fmt.Errorf("getting contact %d: %w", run.ContactID, err)
	}

	subject := campaign.Subject
	html := campaign.BodyHTML
	text := campaign.BodyText
	var variantID int64

	variants, _ := r.store.ListVariants(campaign.ID)
	if len(variants) > 0 {
		bucket := contact.ID % 100
		cumulative := 0
		for _, v := range variants {
			cumulative += v.SplitPercent
			if int(bucket) < cumulative {
				variantID = v.ID
				if v.Subject != "" {
					subject = v.Subject
				}
				if v.BodyHTML != "" {
					html = v.BodyHTML
				}
				if v.BodyText != "" {
					text = v.BodyText
				}
				break
			}
		}
	}

	contactFields := map[string]string{
		"email":      contact.Email,
		"first_name": contact.FirstName,
		"last_name":  contact.LastName,
		"phone":      contact.Phone,
		"company":    contact.Company,
		"title":      contact.Title,
		"tags":       contact.Tags,
		"source":     contact.Source,
		"status":     contact.Status,
	}
	html = Render(html, contactFields)
	text = Render(text, contactFields)

	token := r.sender.GenerateUnsubscribeToken(contact.ID, campaign.ID, r.unsubscribeSecret)
	unsubURL := r.sender.baseURL + "/api/v1/marketing/unsubscribe/" + token
	html, text = AppendUnsubscribeFooter(html, text, unsubURL)

	headers := map[string]string{
		"List-Unsubscribe": "<" + unsubURL + ">",
	}

	sendErr := r.sender.Send(contact.Email, subject, html, text, headers)

	deliveryStatus := "sent"
	deliveryError := ""
	if sendErr != nil {
		deliveryStatus = "failed"
		deliveryError = sendErr.Error()
	}

	cid := contact.ID
	r.store.CreateDelivery(&storage.EmailDelivery{
		CampaignAccount: campaign.Account,
		CampaignName:    campaign.Name,
		ContactID:       &cid,
		VariantID:       variantID,
		Subject:         subject,
		Status:          deliveryStatus,
		Error:           deliveryError,
	})

	if sendErr != nil {
		return sendErr
	}
	return nil
}
