package repository

import (
	"context"
	"crypto/rand"
	"fmt"
	"strings"
	"time"

	"github.com/imamsyafii-dev/queuewise/backend/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

type QueuewiseRepository struct {
	db *pgxpool.Pool
}

func NewQueuewiseRepository(db *pgxpool.Pool) *QueuewiseRepository {
	return &QueuewiseRepository{db: db}
}

func (r *QueuewiseRepository) Catalog(ctx context.Context) (model.Catalog, error) {
	catalog := model.Catalog{Branches: []model.Branch{}, Services: []model.Service{}, Staff: []model.StaffMember{}}

	branchRows, err := r.db.Query(ctx, `SELECT id, name, slug, COALESCE(address, '') FROM branches WHERE is_active = TRUE ORDER BY name`)
	if err != nil {
		return catalog, err
	}
	defer branchRows.Close()
	for branchRows.Next() {
		var branch model.Branch
		if err := branchRows.Scan(&branch.ID, &branch.Name, &branch.Slug, &branch.Address); err != nil {
			return catalog, err
		}
		catalog.Branches = append(catalog.Branches, branch)
	}

	serviceRows, err := r.db.Query(ctx, `SELECT id, name, COALESCE(description, ''), duration_minutes, price_amount FROM services WHERE is_active = TRUE ORDER BY name`)
	if err != nil {
		return catalog, err
	}
	defer serviceRows.Close()
	for serviceRows.Next() {
		var service model.Service
		if err := serviceRows.Scan(&service.ID, &service.Name, &service.Description, &service.DurationMinutes, &service.PriceAmount); err != nil {
			return catalog, err
		}
		catalog.Services = append(catalog.Services, service)
	}

	staffRows, err := r.db.Query(ctx, `SELECT sm.id, sm.name, ss.service_id FROM staff_members sm JOIN staff_services ss ON ss.staff_member_id = sm.id WHERE sm.is_active = TRUE ORDER BY sm.name`)
	if err != nil {
		return catalog, err
	}
	defer staffRows.Close()
	for staffRows.Next() {
		var staff model.StaffMember
		if err := staffRows.Scan(&staff.ID, &staff.Name, &staff.ServiceID); err != nil {
			return catalog, err
		}
		catalog.Staff = append(catalog.Staff, staff)
	}
	return catalog, branchRows.Err()
}

func (r *QueuewiseRepository) CreateAppointment(ctx context.Context, input model.CreateAppointmentInput) (model.Appointment, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return model.Appointment{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var organizationID string
	var duration int
	err = tx.QueryRow(ctx, `
SELECT b.organization_id, s.duration_minutes
FROM branches b
JOIN services s ON s.id = $2 AND s.organization_id = b.organization_id AND s.is_active = TRUE
JOIN staff_members sm ON sm.id = $3 AND sm.branch_id = b.id AND sm.is_active = TRUE
JOIN staff_services ss ON ss.staff_member_id = sm.id AND ss.service_id = s.id
WHERE b.id = $1 AND b.is_active = TRUE`, input.BranchID, input.ServiceID, input.StaffMemberID).Scan(&organizationID, &duration)
	if err != nil {
		return model.Appointment{}, fmt.Errorf("selected branch, service, or staff is unavailable")
	}

	endAt := input.ScheduledStartAt.Add(timeDurationMinutes(duration))
	var overlaps bool
	err = tx.QueryRow(ctx, `
SELECT EXISTS (
 SELECT 1 FROM appointments
 WHERE staff_member_id = $1
   AND status NOT IN ('CANCELLED', 'NO_SHOW')
   AND scheduled_start_at < $3
   AND scheduled_end_at > $2
)`, input.StaffMemberID, input.ScheduledStartAt, endAt).Scan(&overlaps)
	if err != nil {
		return model.Appointment{}, err
	}
	if overlaps {
		return model.Appointment{}, fmt.Errorf("staff already has an appointment at that time")
	}

	var customerID string
	err = tx.QueryRow(ctx, `
INSERT INTO customers (organization_id, name, phone, email)
VALUES ($1, $2, $3, NULLIF($4, ''))
ON CONFLICT (organization_id, phone)
DO UPDATE SET name = EXCLUDED.name, email = COALESCE(EXCLUDED.email, customers.email)
RETURNING id`, organizationID, strings.TrimSpace(input.CustomerName), strings.TrimSpace(input.CustomerPhone), strings.TrimSpace(input.CustomerEmail)).Scan(&customerID)
	if err != nil {
		return model.Appointment{}, err
	}

	code, err := newBookingCode()
	if err != nil {
		return model.Appointment{}, err
	}
	var appointmentID string
	err = tx.QueryRow(ctx, `
INSERT INTO appointments (booking_code, organization_id, branch_id, customer_id, service_id, staff_member_id, scheduled_start_at, scheduled_end_at, notes)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,NULLIF($9,'')) RETURNING id`, code, organizationID, input.BranchID, customerID, input.ServiceID, input.StaffMemberID, input.ScheduledStartAt, endAt, strings.TrimSpace(input.Notes)).Scan(&appointmentID)
	if err != nil {
		return model.Appointment{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return model.Appointment{}, err
	}
	return r.AppointmentByID(ctx, appointmentID)
}

func (r *QueuewiseRepository) AppointmentByCode(ctx context.Context, code string) (model.Appointment, error) {
	return r.appointment(ctx, `a.booking_code = $1`, strings.ToUpper(strings.TrimSpace(code)))
}

func (r *QueuewiseRepository) AppointmentByID(ctx context.Context, id string) (model.Appointment, error) {
	return r.appointment(ctx, `a.id = $1`, id)
}

func (r *QueuewiseRepository) appointment(ctx context.Context, condition string, value string) (model.Appointment, error) {
	query := `SELECT a.id, a.booking_code, b.name, b.slug, c.name, c.phone, s.name, sm.name, a.scheduled_start_at, a.scheduled_end_at, a.status, q.id, q.queue_number, q.status
FROM appointments a JOIN branches b ON b.id=a.branch_id JOIN customers c ON c.id=a.customer_id JOIN services s ON s.id=a.service_id JOIN staff_members sm ON sm.id=a.staff_member_id
LEFT JOIN queue_entries q ON q.appointment_id=a.id WHERE ` + condition
	var item model.Appointment
	err := r.db.QueryRow(ctx, query, value).Scan(&item.ID, &item.BookingCode, &item.BranchName, &item.BranchSlug, &item.CustomerName, &item.CustomerPhone, &item.ServiceName, &item.StaffName, &item.ScheduledStartAt, &item.ScheduledEndAt, &item.Status, &item.QueueID, &item.QueueNumber, &item.QueueStatus)
	return item, err
}

func (r *QueuewiseRepository) ListAppointments(ctx context.Context) ([]model.Appointment, error) {
	rows, err := r.db.Query(ctx, `SELECT a.id, a.booking_code, b.name, b.slug, c.name, c.phone, s.name, sm.name, a.scheduled_start_at, a.scheduled_end_at, a.status, q.id, q.queue_number, q.status
FROM appointments a JOIN branches b ON b.id=a.branch_id JOIN customers c ON c.id=a.customer_id JOIN services s ON s.id=a.service_id JOIN staff_members sm ON sm.id=a.staff_member_id LEFT JOIN queue_entries q ON q.appointment_id=a.id ORDER BY a.scheduled_start_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []model.Appointment{}
	for rows.Next() {
		var item model.Appointment
		if err := rows.Scan(&item.ID, &item.BookingCode, &item.BranchName, &item.BranchSlug, &item.CustomerName, &item.CustomerPhone, &item.ServiceName, &item.StaffName, &item.ScheduledStartAt, &item.ScheduledEndAt, &item.Status, &item.QueueID, &item.QueueNumber, &item.QueueStatus); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *QueuewiseRepository) CheckIn(ctx context.Context, appointmentID string) (model.Appointment, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return model.Appointment{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var branchID string
	var status string
	err = tx.QueryRow(ctx, `SELECT branch_id, status FROM appointments WHERE id=$1 FOR UPDATE`, appointmentID).Scan(&branchID, &status)
	if err != nil {
		return model.Appointment{}, err
	}
	if status != "BOOKED" {
		return model.Appointment{}, fmt.Errorf("only booked appointments can check in")
	}
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtext($1))`, branchID); err != nil {
		return model.Appointment{}, err
	}
	var queueNumber int
	err = tx.QueryRow(ctx, `SELECT COALESCE(MAX(queue_number), 0) + 1 FROM queue_entries WHERE branch_id=$1 AND queue_date=CURRENT_DATE`, branchID).Scan(&queueNumber)
	if err != nil {
		return model.Appointment{}, err
	}
	_, err = tx.Exec(ctx, `INSERT INTO queue_entries (appointment_id, branch_id, queue_date, queue_number) VALUES ($1,$2,CURRENT_DATE,$3)`, appointmentID, branchID, queueNumber)
	if err != nil {
		return model.Appointment{}, err
	}
	_, err = tx.Exec(ctx, `UPDATE appointments SET status='IN_QUEUE' WHERE id=$1`, appointmentID)
	if err != nil {
		return model.Appointment{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return model.Appointment{}, err
	}
	return r.AppointmentByID(ctx, appointmentID)
}

func (r *QueuewiseRepository) QueueBoard(ctx context.Context, branchSlug string) (model.QueueBoard, error) {
	board := model.QueueBoard{Waiting: []model.QueueEntry{}}
	err := r.db.QueryRow(ctx, `SELECT name FROM branches WHERE slug=$1`, branchSlug).Scan(&board.BranchName)
	if err != nil {
		return board, err
	}
	rows, err := r.db.Query(ctx, `SELECT q.id, q.queue_number, q.status, c.name, s.name, sm.name FROM queue_entries q JOIN appointments a ON a.id=q.appointment_id JOIN customers c ON c.id=a.customer_id JOIN services s ON s.id=a.service_id JOIN staff_members sm ON sm.id=a.staff_member_id JOIN branches b ON b.id=q.branch_id WHERE b.slug=$1 AND q.queue_date=CURRENT_DATE AND q.status IN ('WAITING','CALLED','IN_SERVICE') ORDER BY q.queue_number`, branchSlug)
	if err != nil {
		return board, err
	}
	defer rows.Close()
	for rows.Next() {
		var item model.QueueEntry
		if err := rows.Scan(&item.ID, &item.QueueNumber, &item.Status, &item.CustomerName, &item.ServiceName, &item.StaffName); err != nil {
			return board, err
		}
		if item.Status == "CALLED" || item.Status == "IN_SERVICE" {
			board.Current = &item
		} else {
			board.Waiting = append(board.Waiting, item)
		}
	}
	return board, rows.Err()
}

func (r *QueuewiseRepository) UpdateQueueStatus(ctx context.Context, queueID, action string) (string, error) {
	updates := map[string]struct{ queueStatus, appointmentStatus, timestamp string }{
		"call":     {"CALLED", "CALLED", "called_at=NOW()"},
		"start":    {"IN_SERVICE", "IN_SERVICE", "started_at=NOW()"},
		"complete": {"COMPLETED", "COMPLETED", "completed_at=NOW()"},
	}
	update, ok := updates[action]
	if !ok {
		return "", fmt.Errorf("unknown queue action")
	}
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var appointmentID string
	var branchID string
	err = tx.QueryRow(ctx, `UPDATE queue_entries SET status=$2, `+update.timestamp+` WHERE id=$1 RETURNING appointment_id, branch_id`, queueID, update.queueStatus).Scan(&appointmentID, &branchID)
	if err != nil {
		return "", err
	}
	_, err = tx.Exec(ctx, `UPDATE appointments SET status=$2 WHERE id=$1`, appointmentID, update.appointmentStatus)
	if err != nil {
		return "", err
	}
	var branchSlug string
	if err := tx.QueryRow(ctx, `SELECT slug FROM branches WHERE id=$1`, branchID).Scan(&branchSlug); err != nil {
		return "", err
	}
	if err := tx.Commit(ctx); err != nil {
		return "", err
	}
	return branchSlug, nil
}

func newBookingCode() (string, error) {
	const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	bytes := make([]byte, 6)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	var builder strings.Builder
	builder.WriteString("QW-")
	for _, b := range bytes {
		builder.WriteByte(alphabet[int(b)%len(alphabet)])
	}
	return builder.String(), nil
}

func timeDurationMinutes(minutes int) time.Duration { return time.Duration(minutes) * time.Minute }
