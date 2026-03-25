package domain

import "errors"

// Errores de dominio para Order — tipados para que el handler los mapee
// correctamente a códigos HTTP o gRPC.
//
// REGLA: estos errores representan casos de negocio, NO errores técnicos.
var (
	ErrOrderNotFound      = errors.New("order not found")
	ErrOrderDuplicated    = errors.New("order already exists")
	ErrInvalidOrderUserId = errors.New("user_id cannot be empty")
	ErrInvalidOrderTotal  = errors.New("total cannot be empty")
)
