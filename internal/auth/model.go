package auth

type User struct {
	UserID        string   `json:"userId"`
	Username      string   `json:"username"`
	DisplayName   string   `json:"displayName"`
	Email         string   `json:"email,omitempty"`
	PlatformRoles []string `json:"platformRoles"`
}

type userRecord struct {
	ID           string
	Username     string
	PasswordHash string
	DisplayName  string
	Email        string
	Status       string
}

type Actor struct {
	UserID        string
	Username      string
	PlatformRoles []string
}

func (a Actor) HasRole(role string) bool {
	for _, current := range a.PlatformRoles {
		if current == role {
			return true
		}
	}
	return false
}

func toUser(record userRecord) User {
	return User{
		UserID:        record.ID,
		Username:      record.Username,
		DisplayName:   record.DisplayName,
		Email:         record.Email,
		PlatformRoles: platformRolesFor(record),
	}
}

func platformRolesFor(record userRecord) []string {
	if record.Username == "admin" {
		return []string{"SUPER_ADMIN"}
	}
	return []string{}
}
