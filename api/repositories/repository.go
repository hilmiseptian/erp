// repository.go
package repositories

// shared types & helpers untuk semua repository
type scanner interface{ Scan(...any) error }
