package monitoring

import (
	"fmt"
	"time"
)

type AuditEvent struct {
	Timestamp time.Time
	UserID    string
	Action    string
	Resource  string
	Success   bool
	Details   string
}

func NewAuditEvent(userID, action, resource string, success bool, details string) AuditEvent {
	return AuditEvent{
		Timestamp: time.Now(),
		UserID:    userID,
		Action:    action,
		Resource:  resource,
		Success:   success,
		Details:   details,
	}
}

// Aquí se puede extender para enviar a archivo, base de datos, o sistema externo
func LogAuditEvent(event AuditEvent) {
	// TODO: Implementar persistencia (archivo, base de datos, etc.)
	// Ejemplo simple: imprimir en consola
	fmt.Printf("AUDIT | %s | user=%s | action=%s | resource=%s | success=%t | details=%s\n", event.Timestamp.Format(time.RFC3339), event.UserID, event.Action, event.Resource, event.Success, event.Details)
}
