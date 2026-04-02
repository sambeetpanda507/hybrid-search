package database

import (
	"github.com/pgvector/pgvector-go"
	"gorm.io/gorm"
)

type Job struct {
	gorm.Model
	Title          string          `json:"title"`
	Experience     int             `json:"experience"`
	EducationLevel string          `json:"educationLevel"`
	SkillsCount    int             `json:"skillCount"`
	Industry       string          `json:"industry"`
	CompanySize    string          `json:"companySize"`
	Location       string          `json:"location"`
	RemoteWork     string          `json:"remoteWork"`
	Certifications int             `json:"certifications"`
	Salary         int             `json:"salary"`
	Embedding      pgvector.Vector `gorm:"type:vector(384)" json:"embedding"`
}
