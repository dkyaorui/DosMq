package modules

type AuthKey struct {
    Value string `json:"auth_key" binding:"required"`
}
