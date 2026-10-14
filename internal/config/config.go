package config

import (
	"fmt"
	"net/url"
	"os"

	"github.com/joho/godotenv"
)

func must(key string) (string, error) {
	v := os.Getenv(key)
	if v == "" {
		return "", fmt.Errorf("env %s belum diset", key)
	}
	return v, nil
}

// DatabaseURL membangun DSN dari env; mengembalikan error bila ada yang kosong.
func DatabaseURL() (string, error) {
	_ = godotenv.Load() // abaikan error: di production env biasanya dari sistem

	var vals [6]string
	for i, k := range []string{"DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME", "DB_SSLMODE"} {
		v, err := must(k)
		if err != nil {
			return "", err
		}
		vals[i] = v
	}
	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(vals[2], vals[3]), // otomatis meng-escape karakter khusus
		Host:     vals[0] + ":" + vals[1],
		Path:     vals[4],
		RawQuery: "sslmode=" + vals[5],
	}
	return u.String(), nil
}
