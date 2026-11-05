package model

import "time"

type Branch struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Slug    string `json:"slug"`
	Address string `json:"address"`
}

type Service struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	DurationMinutes int    `json:"duration_minutes"`
	PriceAmount     int64  `json:"price_amount"`
}

type StaffMember struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	ServiceID string `json:"service_id,omitempty"`
}

type Catalog struct {
	Branches []Branch      `json:"branches"`
	Services []Service     `json:"services"`
	Staff    []StaffMember `json:"staff"`
}

type CreateAppointmentInput struct {
	BranchID         string    `json:"branch_id" binding:"required,uuid"`
	ServiceID        string    `json:"service_id" binding:"required,uuid"`
	StaffMemberID    string    `json:"staff_member_id" binding:"required,uuid"`
	CustomerName     string    `json:"customer_name" binding:"required,min=2,max=150"`
	CustomerPhone    string    `json:"customer_phone" binding:"required,min=8,max=30"`
	CustomerEmail    string    `json:"customer_email" binding:"omitempty,email"`
	ScheduledStartAt time.Time `json:"scheduled_start_at" binding:"required"`
	Notes            string    `json:"notes" binding:"max=500"`
}

type Appointment struct {
	ID               string    `json:"id"`
	BookingCode      string    `json:"booking_code"`
	BranchName       string    `json:"branch_name"`
	BranchSlug       string    `json:"branch_slug"`
	CustomerName     string    `json:"customer_name"`
	CustomerPhone    string    `json:"customer_phone"`
	ServiceName      string    `json:"service_name"`
	StaffName        string    `json:"staff_name"`
	ScheduledStartAt time.Time `json:"scheduled_start_at"`
	ScheduledEndAt   time.Time `json:"scheduled_end_at"`
	Status           string    `json:"status"`
	QueueID          *string   `json:"queue_id,omitempty"`
	QueueNumber      *int      `json:"queue_number,omitempty"`
	QueueStatus      *string   `json:"queue_status,omitempty"`
}

type QueueEntry struct {
	ID           string `json:"id"`
	QueueNumber  int    `json:"queue_number"`
	Status       string `json:"status"`
	CustomerName string `json:"customer_name"`
	ServiceName  string `json:"service_name"`
	StaffName    string `json:"staff_name"`
}

type QueueBoard struct {
	BranchName string       `json:"branch_name"`
	Current    *QueueEntry  `json:"current"`
	Waiting    []QueueEntry `json:"waiting"`
}
