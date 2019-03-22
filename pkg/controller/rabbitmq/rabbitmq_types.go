package rabbitmq

type basicAuthCredentials struct {
	username string
	password string
}

type secretResouces struct {
	ServiceAccount string
	Credentials    string
}

type rabbitmqUserStruct struct {
	Name             string `json:"name"`
	Password         string `json:"password,omitempty"`
	PasswordHash     string `json:"password_hash,omitempty"`
	HashingAlgorithm string `json:"hashing_algorithm,omitempty"`
	Tags             string `json:"tags,omitempty"`
}

type rabbitmqUsersListStruct struct {
	Users []string `json:"users"`
}
