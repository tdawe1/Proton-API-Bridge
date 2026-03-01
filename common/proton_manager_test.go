package common

import (
	"context"
	"testing"

	"github.com/go-resty/resty/v2"
)

type fakePreRequestHookClient struct {
	hooks []resty.RequestMiddleware
}

func (f *fakePreRequestHookClient) AddPreRequestHook(hook resty.RequestMiddleware) {
	f.hooks = append(f.hooks, hook)
}

func TestAttachDriveSDKHeaderHookSkipsEmptyVersion(t *testing.T) {
	fakeClient := &fakePreRequestHookClient{}
	attachDriveSDKHeaderHook(fakeClient, "")

	if len(fakeClient.hooks) != 0 {
		t.Fatalf("expected no hooks, got %d", len(fakeClient.hooks))
	}
}

func TestAttachDriveSDKHeaderHookSetsHeader(t *testing.T) {
	fakeClient := &fakePreRequestHookClient{}
	attachDriveSDKHeaderHook(fakeClient, "js@0.10.0")

	if len(fakeClient.hooks) != 1 {
		t.Fatalf("expected one hook, got %d", len(fakeClient.hooks))
	}

	r := resty.New().R().SetContext(context.Background())
	if err := fakeClient.hooks[0](resty.New(), r); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if got := r.Header.Get("x-pm-drive-sdk-version"); got != "js@0.10.0" {
		t.Fatalf("expected drive sdk header to be set, got %q", got)
	}
}
