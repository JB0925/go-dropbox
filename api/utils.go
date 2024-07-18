package api

func isInvalidData(sd SignupData) bool {
	return sd.Username == "" || sd.Password == "" || len(sd.Username) < 4 || len(sd.Password) < 8
}