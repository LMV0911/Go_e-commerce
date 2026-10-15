package utils_test

import (
	"strings"
	"testing"

	"github.com/LMV0911/Go_e-commerce/internal/utils"
)

func TestGenerateID(t *testing.T) {
	id := utils.GenerateID("PDT")
	if !strings.HasPrefix(id, "PDT-") {
		t.Errorf("Format d'ID invalide: attendu 'PDT-...', obtenu %s", id)
	}
	if len(id) != 10 {
		t.Errorf("Longueur d'ID invalide: attendu 10, obtenu %d (%s)", len(id), id)
	}

	id2 := utils.GenerateID("PDT")
	if id == id2 {
		t.Errorf("Les identifiants générés consécutivement devraient être uniques: %s == %s", id, id2)
	}
}

func TestGenerateConfirmationCode(t *testing.T) {
	code := utils.GenerateConfirmationCode()
	if len(code) != 6 {
		t.Errorf("Longueur de code invalide: attendu 6, obtenu %d (%s)", len(code), code)
	}

	code2 := utils.GenerateConfirmationCode()
	if code == code2 {
		t.Errorf("Les codes générés consécutivement devraient être aléatoires: %s == %s", code, code2)
	}
}
