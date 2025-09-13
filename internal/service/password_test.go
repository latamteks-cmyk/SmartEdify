package service

import (
	"testing"
)

func TestHashAndCheckPassword(t *testing.T) {
	password := "Abcdef1!"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Error al hashear la contraseña: %v", err)
	}
	if !CheckPassword(hash, password) {
		t.Errorf("La verificación de contraseña debería ser exitosa")
	}
	if CheckPassword(hash, "otraClave123!") {
		t.Errorf("La verificación con contraseña incorrecta debería fallar")
	}
}

func TestValidatePassword(t *testing.T) {
	cases := []struct {
		password string
		wantErr  bool
		msg      string
	}{
		{"Abcdef1!", false, "válida"},
		{"abcdef1!", true, "falta mayúscula"},
		{"ABCDEF1!", true, "falta minúscula"},
		{"Abcdefgh!", true, "falta número"},
		{"Abcdef12", true, "falta especial"},
		{"123456", true, "demasiado corta y común"},
		{"Password1!", false, "válida"},
		{"password", true, "común"},
	}

	for _, tc := range cases {
		err := ValidatePassword(tc.password)
		if tc.wantErr && err == nil {
			t.Errorf("%s: se esperaba error, pero fue nil", tc.msg)
		}
		if !tc.wantErr && err != nil {
			t.Errorf("%s: no se esperaba error, pero fue %v", tc.msg, err)
		}
	}
}
