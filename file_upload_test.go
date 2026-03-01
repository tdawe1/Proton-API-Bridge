package proton_api_bridge

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	"github.com/rclone/go-proton-api"
)

func TestBuildVerificationTokenXorsWithEncryptedBlock(t *testing.T) {
	verificationCode := []byte{0x10, 0x20, 0x30, 0x40}
	encData := []byte{0x01, 0x02, 0x03, 0x04}

	got := buildVerificationToken(verificationCode, encData)
	want := []byte{0x11, 0x22, 0x33, 0x44}

	if !bytes.Equal(got, want) {
		t.Fatalf("unexpected token: got=%v want=%v", got, want)
	}
}

func TestBuildVerificationTokenKeepsTailWhenBlockShorter(t *testing.T) {
	verificationCode := []byte{0x10, 0x20, 0x30, 0x40}
	encData := []byte{0x01, 0x02}

	got := buildVerificationToken(verificationCode, encData)
	want := []byte{0x11, 0x22, 0x30, 0x40}

	if !bytes.Equal(got, want) {
		t.Fatalf("unexpected token: got=%v want=%v", got, want)
	}
}

func TestRecoverBrokenConflictStateRecreatesWhenCode2501(t *testing.T) {
	apiErr := &proton.APIError{Code: 2501, Status: 422, Message: "File or folder not found"}
	err := fmt.Errorf("wrapped transport error: %w", apiErr)

	deleteCalls := 0
	shouldRecreate, gotErr := recoverBrokenConflictState(err, func() error {
		deleteCalls++
		return nil
	})

	if gotErr != nil {
		t.Fatalf("expected nil error, got %v", gotErr)
	}
	if !shouldRecreate {
		t.Fatalf("expected recreate=true")
	}
	if deleteCalls != 1 {
		t.Fatalf("expected one delete call, got %d", deleteCalls)
	}
}

func TestRecoverBrokenConflictStatePropagatesDeleteError(t *testing.T) {
	apiErr := &proton.APIError{Code: 2501, Status: 422, Message: "File or folder not found"}
	err := fmt.Errorf("wrapped transport error: %w", apiErr)
	deleteErr := errors.New("delete failed")

	shouldRecreate, gotErr := recoverBrokenConflictState(err, func() error {
		return deleteErr
	})

	if !errors.Is(gotErr, deleteErr) {
		t.Fatalf("expected delete error, got %v", gotErr)
	}
	if shouldRecreate {
		t.Fatalf("expected recreate=false")
	}
}

func TestRecoverBrokenConflictStateReturnsOriginalErrorOnOtherCode(t *testing.T) {
	originalErr := fmt.Errorf("wrapped transport error: %w", &proton.APIError{Code: 2500, Status: 422, Message: "conflict"})

	deleteCalls := 0
	shouldRecreate, gotErr := recoverBrokenConflictState(originalErr, func() error {
		deleteCalls++
		return nil
	})

	if !errors.Is(gotErr, originalErr) {
		t.Fatalf("expected original error, got %v", gotErr)
	}
	if shouldRecreate {
		t.Fatalf("expected recreate=false")
	}
	if deleteCalls != 0 {
		t.Fatalf("expected no delete call, got %d", deleteCalls)
	}
}
