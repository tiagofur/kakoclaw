package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const segmentSelectCols = `id, name, COALESCE(slug, ''), COALESCE(description, ''), COALESCE(rules, '[]'), COALESCE(account, ''), COALESCE(contact_count, 0), created_at, updated_at`

func (s *Storage) CreateSegment(seg *Segment) (*Segment, error) {
	if seg.Rules == "" {
		seg.Rules = "[]"
	}
	slug := generateSlug(seg.Name)
	if slug == "" {
		slug = fmt.Sprintf("segment-%d", time.Now().Unix())
	}
	now := time.Now().Format(time.RFC3339)

	var rules []SegmentRule
	if err := json.Unmarshal([]byte(seg.Rules), &rules); err != nil {
		return nil, fmt.Errorf("parsing segment rules: %w", err)
	}
	contactIDs, err := s.EvaluateSegmentRules(rules)
	if err != nil {
		return nil, err
	}
	contactCount := len(contactIDs)

	res, err := s.db.Exec(
		`INSERT INTO segments (name, slug, description, rules, account, contact_count, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		seg.Name, slug, seg.Description, seg.Rules, seg.Account, contactCount, now, now,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint") {
			return nil, fmt.Errorf("segment with slug %s already exists", slug)
		}
		return nil, fmt.Errorf("creating segment: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("getting insert id: %w", err)
	}
	return s.GetSegmentByID(id)
}

func (s *Storage) GetSegmentByID(id int64) (*Segment, error) {
	var seg Segment
	err := s.db.QueryRow(`SELECT `+segmentSelectCols+` FROM segments WHERE id = ?`, id).
		Scan(&seg.ID, &seg.Name, &seg.Slug, &seg.Description, &seg.Rules, &seg.Account, &seg.ContactCount, &seg.CreatedAt, &seg.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getting segment: %w", err)
	}
	return &seg, nil
}

func (s *Storage) UpdateSegment(seg *Segment) error {
	if seg.Rules == "" {
		seg.Rules = "[]"
	}
	var rules []SegmentRule
	if err := json.Unmarshal([]byte(seg.Rules), &rules); err != nil {
		return fmt.Errorf("parsing segment rules: %w", err)
	}
	contactIDs, err := s.EvaluateSegmentRules(rules)
	if err != nil {
		return err
	}
	contactCount := len(contactIDs)

	_, err = s.db.Exec(
		`UPDATE segments SET name=?, description=?, rules=?, account=?, contact_count=?, updated_at=CURRENT_TIMESTAMP WHERE id=?`,
		seg.Name, seg.Description, seg.Rules, seg.Account, contactCount, seg.ID,
	)
	return err
}

func (s *Storage) DeleteSegment(id int64) error {
	_, err := s.db.Exec(`DELETE FROM segments WHERE id = ?`, id)
	return err
}

func (s *Storage) ListSegments(account string) ([]Segment, error) {
	rows, err := s.db.Query(
		`SELECT `+segmentSelectCols+` FROM segments WHERE account = ? ORDER BY created_at DESC`, account,
	)
	if err != nil {
		return nil, fmt.Errorf("listing segments: %w", err)
	}
	defer rows.Close()

	var segments []Segment
	for rows.Next() {
		var seg Segment
		if err := rows.Scan(&seg.ID, &seg.Name, &seg.Slug, &seg.Description, &seg.Rules, &seg.Account, &seg.ContactCount, &seg.CreatedAt, &seg.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning segment: %w", err)
		}
		segments = append(segments, seg)
	}
	return segments, rows.Err()
}

func (s *Storage) EvaluateSegmentRules(rules []SegmentRule) ([]int64, error) {
	rows, err := s.db.Query(`SELECT ` + contactSelectCols + ` FROM contacts`)
	if err != nil {
		return nil, fmt.Errorf("loading contacts for segment evaluation: %w", err)
	}
	defer rows.Close()

	var matchingIDs []int64
	for rows.Next() {
		var c Contact
		if err := scanContact(rows, &c); err != nil {
			return nil, fmt.Errorf("scanning contact: %w", err)
		}
		allMatch := true
		for _, rule := range rules {
			match, evalErr := evaluateRuleAgainstContact(&c, rule)
			if evalErr != nil {
				return nil, evalErr
			}
			if !match {
				allMatch = false
				break
			}
		}
		if allMatch {
			matchingIDs = append(matchingIDs, c.ID)
		}
	}
	return matchingIDs, rows.Err()
}

func getContactFieldValue(contact *Contact, field string) (string, error) {
	switch field {
	case "email":
		return contact.Email, nil
	case "first_name":
		return contact.FirstName, nil
	case "last_name":
		return contact.LastName, nil
	case "phone":
		return contact.Phone, nil
	case "company":
		return contact.Company, nil
	case "title":
		return contact.Title, nil
	case "status":
		return contact.Status, nil
	case "source":
		return contact.Source, nil
	case "tags":
		return contact.Tags, nil
	default:
		if contact.CustomFields != "" && contact.CustomFields != "{}" {
			var cf map[string]interface{}
			if err := json.Unmarshal([]byte(contact.CustomFields), &cf); err == nil {
				if v, ok := cf[field]; ok {
					return fmt.Sprintf("%v", v), nil
				}
			}
		}
		return "", nil
	}
}

func evaluateRuleAgainstContact(contact *Contact, rule SegmentRule) (bool, error) {
	valueStr := fmt.Sprintf("%v", rule.Value)

	if rule.Field == "tags" {
		return evaluateTagsRule(contact.Tags, rule.Operator, valueStr)
	}

	fieldValue, err := getContactFieldValue(contact, rule.Field)
	if err != nil {
		return false, err
	}

	switch rule.Operator {
	case "equals":
		return fieldValue == valueStr, nil
	case "not_equals":
		return fieldValue != valueStr, nil
	case "contains":
		return strings.Contains(fieldValue, valueStr), nil
	case "starts_with":
		return strings.HasPrefix(fieldValue, valueStr), nil
	case "ends_with":
		return strings.HasSuffix(fieldValue, valueStr), nil
	case "greater_than":
		return compareNumbers(fieldValue, valueStr) > 0, nil
	case "less_than":
		return compareNumbers(fieldValue, valueStr) < 0, nil
	case "in_list":
		for _, item := range strings.Split(valueStr, ",") {
			if strings.TrimSpace(item) == fieldValue {
				return true, nil
			}
		}
		return false, nil
	case "not_in_list":
		for _, item := range strings.Split(valueStr, ",") {
			if strings.TrimSpace(item) == fieldValue {
				return false, nil
			}
		}
		return true, nil
	default:
		return false, fmt.Errorf("unsupported operator '%s'; valid operators: equals, not_equals, contains, starts_with, ends_with, greater_than, less_than, in_list, not_in_list", rule.Operator)
	}
}

func evaluateTagsRule(tagsJSON, operator, valueStr string) (bool, error) {
	var tags []string
	if err := json.Unmarshal([]byte(tagsJSON), &tags); err != nil {
		tags = []string{}
	}

	switch operator {
	case "equals":
		for _, tag := range tags {
			if tag == valueStr {
				return true, nil
			}
		}
		return false, nil
	case "not_equals":
		for _, tag := range tags {
			if tag == valueStr {
				return false, nil
			}
		}
		return true, nil
	case "contains":
		for _, tag := range tags {
			if strings.Contains(tag, valueStr) {
				return true, nil
			}
		}
		return false, nil
	case "starts_with":
		for _, tag := range tags {
			if strings.HasPrefix(tag, valueStr) {
				return true, nil
			}
		}
		return false, nil
	case "ends_with":
		for _, tag := range tags {
			if strings.HasSuffix(tag, valueStr) {
				return true, nil
			}
		}
		return false, nil
	case "in_list":
		items := strings.Split(valueStr, ",")
		for _, tag := range tags {
			for _, item := range items {
				if strings.TrimSpace(item) == tag {
					return true, nil
				}
			}
		}
		return false, nil
	case "not_in_list":
		items := strings.Split(valueStr, ",")
		for _, tag := range tags {
			for _, item := range items {
				if strings.TrimSpace(item) == tag {
					return false, nil
				}
			}
		}
		return true, nil
	case "greater_than":
		for _, tag := range tags {
			if compareNumbers(tag, valueStr) > 0 {
				return true, nil
			}
		}
		return false, nil
	case "less_than":
		for _, tag := range tags {
			if compareNumbers(tag, valueStr) < 0 {
				return true, nil
			}
		}
		return false, nil
	default:
		return false, fmt.Errorf("unsupported operator '%s' for tags field; valid operators: equals, not_equals, contains, starts_with, ends_with, greater_than, less_than, in_list, not_in_list", operator)
	}
}

func compareNumbers(a, b string) int {
	aFloat, aErr := strconv.ParseFloat(a, 64)
	bFloat, bErr := strconv.ParseFloat(b, 64)
	if aErr != nil || bErr != nil {
		return strings.Compare(a, b)
	}
	if aFloat < bFloat {
		return -1
	}
	if aFloat > bFloat {
		return 1
	}
	return 0
}
