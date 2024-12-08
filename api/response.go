package api

const adminUrl string = "http://localhost:8080/admin"

type PrequalParametersResponse struct {
	Id                int     `json:"id"`
	MaxLifeTime       int     `json:"max_life_time"`
	PoolSize          int     `json:"pool_size"`
	ProbeFactor       float64 `json:"probe_factor"`
	ProbeRemoveFactor int     `json:"probe_remove_factor"`
	Mu                int     `json:"mu"`
	CreatedAt         string  `json:"created_at"`
	UpdatedAt         string  `json:"updated_at"`
}

type ReplicaResponse struct {
	Id                  int    `json:"id"`
	Name                string `json:"name"`
	Url                 string `json:"url"`
	Status              string `json:"status"`
	HealthCheckEndpoint string `json:"HealthCheckEndpoint"`
	CreatedAt           string `json:"created_at"`
	UpdatedAt           string `json:"updated_at"`
}
