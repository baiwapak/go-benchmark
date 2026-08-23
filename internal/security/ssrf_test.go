package security

import "testing"

func TestValidateTargetURL(t *testing.T) {
	allowed := []string{
		"https://example.com",
		"http://api.example.org/path",
	}
	for _, u := range allowed {
		if !ValidateTargetURL(u) {
			t.Fatalf("expected allowed: %s", u)
		}
	}

	blocked := []string{
		"http://127.0.0.1",
		"http://localhost/admin",
		"http://10.0.0.1",
		"http://192.168.1.1",
		"http://172.16.0.1",
		"http://[::1]/",
		"http://169.254.169.254",
		"http://metadata.local",
		"ftp://example.com",
		"not-a-url",
	}
	for _, u := range blocked {
		if ValidateTargetURL(u) {
			t.Fatalf("expected blocked: %s", u)
		}
	}
}
