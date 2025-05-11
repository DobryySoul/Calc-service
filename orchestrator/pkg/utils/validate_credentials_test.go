package utils_test

import (
	"testing"

	"github.com/DobryySoul/orchestrator/internal/controllers/http/models"
	"github.com/DobryySoul/orchestrator/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestValidateCredentials(t *testing.T) {
	tests := []struct {
		email    string
		password string
		wantErr  bool
	}{
		{
			email:    "test@example.com",
			password: "testpass",
			wantErr:  true,
		},
		{
			email:    "",
			password: "1231231УЦЙУЦЙ",
			wantErr:  true,
		},
		{
			email:    "test@example.ru",
			password: "123123QWEqwe!@#",
			wantErr:  false,
		},
		{
			email:    "example@test.com",
			password: "",
			wantErr:  true,
		},
		{
			email:    "netparolya@mail.ru",
			password: "",
			wantErr:  true,
		},
		{
			email:    "ktoti@gmail.com",
			password: "qwe123321Q",
			wantErr:  true,
		},
		{
			email:    "ktotinesobacagmail.com",
			password: "qweqwe123123QWEQWE!@#!@#",
			wantErr:  true,
		},
		{
			email:    "noname@proton.me",
			password: "1Qq!",
			wantErr:  true,
		},
		{
			email:    "nutignuti@proton.me",
			password: "nfsNUF21$$$",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.email, func(t *testing.T) {
			user := &models.User{
				Email:    tt.email,
				Password: tt.password,
			}
			err := utils.ValidateUserCredentials(user)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
