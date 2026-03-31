package storage

import (
	"database/sql"
	"fmt"
	"time"
)

const automationSelectCols = `id, COALESCE(account, ''), name, COALESCE(trigger_type, 'list_add'), trigger_list_id, delay_hours, campaign_id, COALESCE(status, 'active'), created_at, updated_at`

func scanAutomation(sc rowScanner, a *Automation) error {
	return sc.Scan(
		&a.ID, &a.Account, &a.Name, &a.TriggerType, &a.TriggerListID,
		&a.DelayHours, &a.CampaignID, &a.Status, &a.CreatedAt, &a.UpdatedAt,
	)
}

const runSelectCols = `id, automation_id, contact_id, COALESCE(status, 'pending'), scheduled_at, processed_at, created_at`

func scanAutomationRun(sc rowScanner, r *AutomationRun) error {
	var scheduledAt, processedAt sql.NullString
	err := sc.Scan(
		&r.ID, &r.AutomationID, &r.ContactID, &r.Status,
		&scheduledAt, &processedAt, &r.CreatedAt,
	)
	if err != nil {
		return err
	}
	r.ScheduledAt = parseNullTime(scheduledAt)
	r.ProcessedAt = parseNullTime(processedAt)
	return nil
}

func (s *Storage) CreateAutomation(a *Automation) (*Automation, error) {
	if a.Status == "" {
		a.Status = "active"
	}
	now := time.Now().Format(time.RFC3339)
	res, err := s.db.Exec(
		`INSERT INTO automations (account, name, trigger_type, trigger_list_id, delay_hours, campaign_id, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		a.Account, a.Name, a.TriggerType, a.TriggerListID,
		a.DelayHours, a.CampaignID, a.Status, now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("creating automation: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("getting automation insert id: %w", err)
	}
	return s.GetAutomationByID(id)
}

func (s *Storage) GetAutomationByID(id int64) (*Automation, error) {
	var a Automation
	row := s.db.QueryRow(`SELECT `+automationSelectCols+` FROM automations WHERE id = ?`, id)
	if err := scanAutomation(row, &a); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("getting automation: %w", err)
	}
	return &a, nil
}

func (s *Storage) ListAutomations(account string) ([]Automation, error) {
	rows, err := s.db.Query(`SELECT `+automationSelectCols+` FROM automations WHERE account = ? ORDER BY created_at DESC`, account)
	if err != nil {
		return nil, fmt.Errorf("listing automations: %w", err)
	}
	defer rows.Close()

	var automations []Automation
	for rows.Next() {
		var a Automation
		if err := scanAutomation(rows, &a); err != nil {
			return nil, fmt.Errorf("scanning automation: %w", err)
		}
		automations = append(automations, a)
	}
	if automations == nil {
		automations = []Automation{}
	}
	return automations, rows.Err()
}

func (s *Storage) UpdateAutomation(a *Automation) error {
	now := time.Now().Format(time.RFC3339)
	_, err := s.db.Exec(
		`UPDATE automations SET account = ?, name = ?, trigger_type = ?, trigger_list_id = ?, delay_hours = ?, campaign_id = ?, status = ?, updated_at = ? WHERE id = ?`,
		a.Account, a.Name, a.TriggerType, a.TriggerListID,
		a.DelayHours, a.CampaignID, a.Status, now, a.ID,
	)
	return err
}

func (s *Storage) DeleteAutomation(id int64) error {
	_, err := s.db.Exec(`DELETE FROM automations WHERE id = ?`, id)
	return err
}

func (s *Storage) CreateAutomationRun(r *AutomationRun) (*AutomationRun, error) {
	res, err := s.db.Exec(
		`INSERT INTO automation_runs (automation_id, contact_id, status, scheduled_at) VALUES (?, ?, ?, ?)`,
		r.AutomationID, r.ContactID, r.Status, nilIfZero(r.ScheduledAt),
	)
	if err != nil {
		return nil, fmt.Errorf("creating automation run: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("getting automation run insert id: %w", err)
	}
	var run AutomationRun
	row := s.db.QueryRow(`SELECT `+runSelectCols+` FROM automation_runs WHERE id = ?`, id)
	if err := scanAutomationRun(row, &run); err != nil {
		return nil, fmt.Errorf("reading automation run: %w", err)
	}
	return &run, nil
}

func (s *Storage) ListPendingRuns() ([]AutomationRun, error) {
	rows, err := s.db.Query(`SELECT ` + runSelectCols + ` FROM automation_runs WHERE status = 'pending' AND scheduled_at <= CURRENT_TIMESTAMP ORDER BY scheduled_at ASC`)
	if err != nil {
		return nil, fmt.Errorf("listing pending runs: %w", err)
	}
	defer rows.Close()

	var runs []AutomationRun
	for rows.Next() {
		var r AutomationRun
		if err := scanAutomationRun(rows, &r); err != nil {
			return nil, fmt.Errorf("scanning automation run: %w", err)
		}
		runs = append(runs, r)
	}
	if runs == nil {
		runs = []AutomationRun{}
	}
	return runs, rows.Err()
}

func (s *Storage) UpdateRunStatus(id int64, status string) error {
	_, err := s.db.Exec(
		`UPDATE automation_runs SET status = ?, processed_at = CURRENT_TIMESTAMP WHERE id = ?`,
		status, id,
	)
	return err
}

func (s *Storage) ClaimRun(id int64) (bool, error) {
	res, err := s.db.Exec(`UPDATE automation_runs SET status = 'processing' WHERE id = ? AND status = 'pending'`, id)
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}

func (s *Storage) ListActiveAutomationsByTriggerList(listID int64) ([]Automation, error) {
	rows, err := s.db.Query(`SELECT `+automationSelectCols+` FROM automations WHERE trigger_list_id = ? AND status = 'active'`, listID)
	if err != nil {
		return nil, fmt.Errorf("listing active automations by trigger list: %w", err)
	}
	defer rows.Close()

	var automations []Automation
	for rows.Next() {
		var a Automation
		if err := scanAutomation(rows, &a); err != nil {
			return nil, fmt.Errorf("scanning automation: %w", err)
		}
		automations = append(automations, a)
	}
	if automations == nil {
		automations = []Automation{}
	}
	return automations, rows.Err()
}

func (s *Storage) ListRunsByAutomation(automationID int64) ([]AutomationRun, error) {
	rows, err := s.db.Query(`SELECT `+runSelectCols+` FROM automation_runs WHERE automation_id = ? ORDER BY created_at DESC`, automationID)
	if err != nil {
		return nil, fmt.Errorf("listing runs by automation: %w", err)
	}
	defer rows.Close()

	var runs []AutomationRun
	for rows.Next() {
		var r AutomationRun
		if err := scanAutomationRun(rows, &r); err != nil {
			return nil, fmt.Errorf("scanning automation run: %w", err)
		}
		runs = append(runs, r)
	}
	if runs == nil {
		runs = []AutomationRun{}
	}
	return runs, rows.Err()
}
