package models

import (
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"net"
	"time"
)

/*
import (
	// "database/sql"
	"fmt"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"minimon/database"
	"os"
	"github.com/tcnksm/go-httpstat"
	"log"
	"net/http"
	"time"
	"io"
	"io/ioutil"

	// https://github.com/davecheney/httpstat.git
	// "time"
)
*/


type IPInfo struct {
    CountryISOCode string
    Subdivision    string
    City           string
}

type Status string
type Action string
type Direction string

const (
	Pending  Status = "pending"
	Approved Status = "approved"
	Rejected Status = "rejected"
)

const (
	Allow   Action = "allow"
	Deny    Action = "deny"
	Observe Action = "observe"
)

const (
	Inbound Direction = "inbound"
	// Outbound    Direction = "outbound"
)

type User struct {
	ID       uint   `gorm:"primaryKey"`
	Username string `gorm:"unique"`
	Password string
	Secret   string // For storing the OTP secret
	Name     string `gorm:"not null"`
	Email    string `gorm:"unique;not null"`
}

// Define your JWT Claims structure
type JwtCustomClaims struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	jwt.StandardClaims
}

/*
type User struct {
        ID    uint   `gorm:"primaryKey"`
        Name  string `gorm:"not null"`
        Email string `gorm:"unique;not null"`
}
*/

type NetIPNet struct {
	*net.IPNet
}

// SrcIP   net.IP `gorm:"uniqueIndex"`
// SrcIPNet *net.IPNet `gorm:"uniqueIndex"`
type FWRule struct {
	ID uint `gorm:"primaryKey"`
	// SrcIPNet NetIPNet       `gorm:"uniqueIndex;type:jsonb"`
	// SrcIPNet string `gorm:"uniqueIndex"`
	// SrcIPNet string `gorm:"uniqueIndex:idx_action_src_ip_net"`
	SrcIPNet string `gorm:"idx_direction__action__src_ip_net"`
	// Action   Action
	// SrcIP    string `gorm:"uniqueIndex:idx_src_ip_action"`
	Action Action `gorm:"not null:default:'allow':idx_direction__action__src_ip_net"`
	// Priority  int `gorm:"not null:idx_action__priority"`
	Priority int `gorm:"not null:idx_direction__priority"`
	// Direction Direction `gorm:"not null:default:'inbound':idx_direction__priority:idx_direction__action__src_ip_net"`
	Direction Direction `gorm:"default:'inbound':idx_direction__priority:idx_direction__action__src_ip_net"`

	Active bool `gorm:"default:true"`
	Log    bool `gorm:"default:false"`
	Note   string
	// CreatedAt time.Time
	// UpdatedAt time.Time
	// DeletedAt gorm.DeletedAt `gorm:"index"`
}

type CountryCodeRule struct {
	ID        uint   `gorm:"primaryKey"`
	Code      string `gorm:"not null"`
	Action    Action `gorm:"not null"`
	Priority  int    `gorm:"not null:idx_action__priority"`
	Active    bool   `gorm:"default:true"`
	Log       bool   `gorm:"default:false"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type CityCodeRule struct {
	ID        uint   `gorm:"primaryKey"`
	Code      string `gorm:"not null"`
	Action    Action `gorm:"not null"`
	Priority  int    `gorm:"not null:idx_action__priority"`
	Active    bool   `gorm:"default:true"`
	Log       bool   `gorm:"default:false"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type Page struct {
	// ID       int    `gorm:"primary_key"`
	// ID             uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primary_key;"`
	// ID   string `json:"id" gorm:"type:uuid;primary_key"`
	ID       uuid.UUID `gorm:"type:uuid;primaryKey"`
	Title    string    `json:"title"`
	Content  string    `json:"content"`
	Template string    `json:"template"`
	Name     string    `json:"name"`
	Path     string    `json:"path"`
	Visits   int       `json:"visits"`
	// Slug string `json:"slug"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
