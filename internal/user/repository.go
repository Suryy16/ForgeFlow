package user

import "time"

type Repository struct {
	users []User
}

var location = time.FixedZone("UTC+8", 8*60*60)

func NewRepository() *Repository {
	return &Repository{
		users: []User{
			{
				ID:         1,
				Name:       "Budi Santoso",
				Email:      "budi.santoso@example.com",
				Password:   "hashed_password_1",
				Created_at: time.Date(2024, 1, 15, 10, 30, 0, 0, location),
				Updated_at: time.Date(2024, 1, 15, 10, 30, 0, 0, location),
			},
			{
				ID:         2,
				Name:       "Siti Nurhaliza",
				Email:      "siti.nurhaliza@example.com",
				Password:   "hashed_password_2",
				Created_at: time.Date(2024, 2, 20, 14, 45, 0, 0, location),
				Updated_at: time.Date(2024, 2, 20, 14, 45, 0, 0, location),
			},
			{
				ID:         3,
				Name:       "Ahmad Wijaya",
				Email:      "ahmad.wijaya@example.com",
				Password:   "hashed_password_3",
				Created_at: time.Date(2024, 3, 5, 9, 15, 0, 0, location),
				Updated_at: time.Date(2024, 3, 10, 11, 20, 0, 0, location),
			},
			{
				ID:         4,
				Name:       "Dewi Lestari",
				Email:      "dewi.lestari@example.com",
				Password:   "hashed_password_4",
				Created_at: time.Date(2024, 3, 12, 8, 0, 0, 0, location),
				Updated_at: time.Date(2024, 3, 12, 8, 0, 0, 0, location),
			},
			{
				ID:         5,
				Name:       "Rizky Pratama",
				Email:      "rizky.pratama@example.com",
				Password:   "hashed_password_5",
				Created_at: time.Date(2024, 4, 1, 16, 30, 0, 0, location),
				Updated_at: time.Date(2024, 4, 5, 10, 15, 0, 0, location),
			},
			{
				ID:         6,
				Name:       "Maya Sari",
				Email:      "maya.sari@example.com",
				Password:   "hashed_password_6",
				Created_at: time.Date(2024, 4, 18, 13, 45, 0, 0, location),
				Updated_at: time.Date(2024, 4, 18, 13, 45, 0, 0, location),
			},
			{
				ID:         7,
				Name:       "Andi Setiawan",
				Email:      "andi.setiawan@example.com",
				Password:   "hashed_password_7",
				Created_at: time.Date(2024, 5, 2, 11, 0, 0, 0, location),
				Updated_at: time.Date(2024, 5, 8, 15, 30, 0, 0, location),
			},
			{
				ID:         8,
				Name:       "Rina Marlina",
				Email:      "rina.marlina@example.com",
				Password:   "hashed_password_8",
				Created_at: time.Date(2024, 5, 20, 7, 30, 0, 0, location),
				Updated_at: time.Date(2024, 5, 20, 7, 30, 0, 0, location),
			},
			{
				ID:         9,
				Name:       "Fajar Nugroho",
				Email:      "fajar.nugroho@example.com",
				Password:   "hashed_password_9",
				Created_at: time.Date(2024, 6, 10, 14, 0, 0, 0, location),
				Updated_at: time.Date(2024, 6, 15, 9, 45, 0, 0, location),
			},
			{
				ID:         10,
				Name:       "Lina Kusuma",
				Email:      "lina.kusuma@example.com",
				Password:   "hashed_password_10",
				Created_at: time.Date(2024, 6, 25, 10, 20, 0, 0, location),
				Updated_at: time.Date(2024, 6, 25, 10, 20, 0, 0, location),
			},
		},
	}
}

func (r *Repository) GetAll() []User {
	return r.users
}
