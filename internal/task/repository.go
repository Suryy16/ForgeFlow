package task

import "time"

type Repository struct {
	tasks []Task
}

var location = time.FixedZone("UTC+8", 8*60*60)

func NewRepository() *Repository {
	return &Repository{
		tasks: []Task{
			{
				ID:          1,
				Owner_ID:    101,
				Assignee_ID: 201,
				Title:       "Implementasi Login API",
				Description: "Membuat endpoint login dengan JWT authentication",
				Status:      "IN_PROGRES",
				Priority:    "HIGH",
				Due_date:    time.Date(2024, 7, 15, 17, 0, 0, 0, location),
				Created_at:  time.Date(2024, 7, 1, 9, 30, 0, 0, location),
				Updated_at:  time.Date(2024, 7, 5, 14, 20, 0, 0, location),
			},
			{
				ID:          2,
				Owner_ID:    101,
				Assignee_ID: 202,
				Title:       "Desain Halaman Dashboard",
				Description: "Membuat mockup dan UI dashboard admin menggunakan Figma",
				Status:      "TODO",
				Priority:    "MEDIUM",
				Due_date:    time.Date(2024, 7, 20, 17, 0, 0, 0, location),
				Created_at:  time.Date(2024, 7, 2, 10, 0, 0, 0, location),
				Updated_at:  time.Date(2024, 7, 2, 10, 0, 0, 0, location),
			},
			{
				ID:          3,
				Owner_ID:    102,
				Assignee_ID: 203,
				Title:       "Fix Bug pada Fitur Upload",
				Description: "File upload gagal ketika ukuran lebih dari 5MB",
				Status:      "DONE",
				Priority:    "HIGH",
				Due_date:    time.Date(2024, 7, 10, 12, 0, 0, 0, location),
				Created_at:  time.Date(2024, 6, 28, 8, 15, 0, 0, location),
				Updated_at:  time.Date(2024, 7, 9, 16, 45, 0, 0, location),
			},
			{
				ID:          4,
				Owner_ID:    102,
				Assignee_ID: 201,
				Title:       "Menulis Dokumentasi API",
				Description: "Membuat dokumentasi endpoint menggunakan Swagger",
				Status:      "TODO",
				Priority:    "LOW",
				Due_date:    time.Date(2024, 7, 30, 17, 0, 0, 0, location),
				Created_at:  time.Date(2024, 7, 3, 13, 0, 0, 0, location),
				Updated_at:  time.Date(2024, 7, 3, 13, 0, 0, 0, location),
			},
			{
				ID:          5,
				Owner_ID:    103,
				Assignee_ID: 204,
				Title:       "Setup CI/CD Pipeline",
				Description: "Konfigurasi GitHub Actions untuk deployment otomatis ke server staging",
				Status:      "IN_PROGRES",
				Priority:    "MEDIUM",
				Due_date:    time.Date(2024, 7, 25, 17, 0, 0, 0, location),
				Created_at:  time.Date(2024, 7, 4, 11, 30, 0, 0, location),
				Updated_at:  time.Date(2024, 7, 6, 9, 15, 0, 0, location),
			},
		},
	}
}

func (r *Repository) GetAll() []Task {
	return r.tasks
}
