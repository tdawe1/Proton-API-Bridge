package proton_api_bridge

import (
	"context"
	"testing"

	"github.com/rclone/go-proton-api"
)

type moveLinkByVolumeOnlyClient struct {
	calledByVolume bool
}

func (c *moveLinkByVolumeOnlyClient) MoveLinkByVolume(context.Context, string, string, proton.MoveLinkReq) error {
	c.calledByVolume = true
	return nil
}

type moveLinkByShareOnlyClient struct {
	calledByShare bool
}

func (c *moveLinkByShareOnlyClient) MoveLink(context.Context, string, string, proton.MoveLinkReq) error {
	c.calledByShare = true
	return nil
}

type moveLinkInvalidSignatureClient struct{}

func (c *moveLinkInvalidSignatureClient) MoveLinkByVolume(context.Context, string) error {
	return nil
}

type moveLinkPanicClient struct{}

func (c *moveLinkPanicClient) MoveLinkByVolume(context.Context, string, string, proton.MoveLinkReq) error {
	panic("boom")
}

func TestMoveLinkCompatPrefersVolumeRoute(t *testing.T) {
	client := &moveLinkByVolumeOnlyClient{}
	err := moveLinkCompat(context.Background(), client, "share", "volume", "link", proton.MoveLinkReq{})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !client.calledByVolume {
		t.Fatalf("expected MoveLinkByVolume to be called")
	}
}

func TestMoveLinkCompatFallsBackToShareRoute(t *testing.T) {
	client := &moveLinkByShareOnlyClient{}
	err := moveLinkCompat(context.Background(), client, "share", "volume", "link", proton.MoveLinkReq{})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !client.calledByShare {
		t.Fatalf("expected MoveLink to be called")
	}
}

func TestSetNilPointerFieldIfPresent(t *testing.T) {
	contentHash := "hash"
	req := proton.MoveLinkReq{}
	setStringFieldIfPresent(&req, "ParentLinkID", "parent")
	setStringFieldIfPresent(&req, "OriginalHash", "orig")
	setStringFieldIfPresent(&req, "Hash", "next")

	reqWithField := struct {
		proton.MoveLinkReq
		ContentHash *string
	}{
		MoveLinkReq: req,
		ContentHash: &contentHash,
	}

	setNilPointerFieldIfPresent(&reqWithField, "ContentHash")
	if reqWithField.ContentHash != nil {
		t.Fatalf("expected ContentHash to be nil")
	}
}

func TestMoveLinkCompatHandlesInvalidSignature(t *testing.T) {
	err := moveLinkCompat(context.Background(), &moveLinkInvalidSignatureClient{}, "share", "volume", "link", proton.MoveLinkReq{})
	if err == nil {
		t.Fatalf("expected error for invalid move method signature")
	}
}

func TestMoveLinkCompatHandlesPanickingMethod(t *testing.T) {
	err := moveLinkCompat(context.Background(), &moveLinkPanicClient{}, "share", "volume", "link", proton.MoveLinkReq{})
	if err == nil {
		t.Fatalf("expected panic to be converted into error")
	}
}
