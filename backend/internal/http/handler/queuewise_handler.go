package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/imamsyafii-dev/queuewise/backend/internal/model"
	"github.com/imamsyafii-dev/queuewise/backend/internal/realtime"
	"github.com/imamsyafii-dev/queuewise/backend/internal/repository"
	"github.com/jackc/pgx/v5"
)

type QueuewiseHandler struct {
	repository *repository.QueuewiseRepository
	events     *realtime.QueueEvents
}

func NewQueuewiseHandler(repository *repository.QueuewiseRepository, events *realtime.QueueEvents) *QueuewiseHandler {
	return &QueuewiseHandler{repository: repository, events: events}
}

func (h *QueuewiseHandler) Catalog(c *gin.Context) {
	catalog, err := h.repository.Catalog(c.Request.Context())
	if err != nil {
		internalError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": catalog})
}

func (h *QueuewiseHandler) CreateAppointment(c *gin.Context) {
	var input model.CreateAppointmentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid booking data", "details": err.Error()})
		return
	}
	appointment, err := h.repository.CreateAppointment(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": appointment})
}

func (h *QueuewiseHandler) Appointment(c *gin.Context) {
	appointment, err := h.repository.AppointmentByCode(c.Request.Context(), c.Param("bookingCode"))
	if errors.Is(err, pgx.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Booking not found"})
		return
	}
	if err != nil {
		internalError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": appointment})
}

func (h *QueuewiseHandler) QueueBoard(c *gin.Context) {
	board, err := h.repository.QueueBoard(c.Request.Context(), c.Param("branchSlug"))
	if errors.Is(err, pgx.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Branch not found"})
		return
	}
	if err != nil {
		internalError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": board})
}

func (h *QueuewiseHandler) QueueEvents(c *gin.Context) {
	branchSlug := c.Param("branchSlug")
	if _, err := h.repository.QueueBoard(c.Request.Context(), branchSlug); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Branch not found"})
			return
		}
		internalError(c, err)
		return
	}

	pubsub := h.events.Subscribe(c.Request.Context(), branchSlug)
	defer pubsub.Close()
	if _, err := pubsub.Receive(c.Request.Context()); err != nil {
		internalError(c, err)
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")
	c.Status(http.StatusOK)
	c.Writer.Flush()

	messages := pubsub.Channel()
	for {
		select {
		case <-c.Request.Context().Done():
			return
		case message, ok := <-messages:
			if !ok {
				return
			}
			c.SSEvent("queue.updated", message.Payload)
			c.Writer.Flush()
		}
	}
}

func (h *QueuewiseHandler) ListAppointments(c *gin.Context) {
	appointments, err := h.repository.ListAppointments(c.Request.Context())
	if err != nil {
		internalError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": appointments})
}

func (h *QueuewiseHandler) CheckIn(c *gin.Context) {
	appointment, err := h.repository.CheckIn(c.Request.Context(), c.Param("appointmentID"))
	if errors.Is(err, pgx.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Appointment not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	if err := h.events.Publish(c.Request.Context(), appointment.BranchSlug); err != nil {
		_ = c.Error(err)
	}
	c.JSON(http.StatusOK, gin.H{"data": appointment})
}

func (h *QueuewiseHandler) QueueAction(c *gin.Context) {
	branchSlug, err := h.repository.UpdateQueueStatus(c.Request.Context(), c.Param("queueID"), c.Param("action"))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Queue entry not found"})
			return
		}
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	if err := h.events.Publish(c.Request.Context(), branchSlug); err != nil {
		_ = c.Error(err)
	}
	c.Status(http.StatusNoContent)
}

func internalError(c *gin.Context, err error) {
	c.Error(err)
	c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
}
