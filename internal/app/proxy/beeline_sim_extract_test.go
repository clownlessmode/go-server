package proxy

import "testing"

func TestExtractBeelineSimCaptureFromSlaveAccounts(t *testing.T) {
	body := []byte(`{
		"data": {
			"currentAccount": "9053099796",
			"slaveAccounts": [
				{"ctn": "9095413697", "login": "9095413697"},
				{"ctn": "9053099796", "login": "9053099796", "isActive": true}
			]
		}
	}`)

	capture := extractBeelineSimCaptureFromSlaveAccounts(body)
	if capture.Preferred != "9053099796" {
		t.Fatalf("preferred=%q want 9053099796", capture.Preferred)
	}
	if len(capture.CTNs) < 2 {
		t.Fatalf("ctns=%v want both slave numbers", capture.CTNs)
	}
}

func TestExtractBeelineSimCaptureFromContext(t *testing.T) {
	body := []byte(`{
		"data": {
			"selectedProduct": {"ctn": "9053099796"},
			"contract": {"ctn": "9095413697", "phone": {"number": "9095413697"}}
		}
	}`)

	capture := extractBeelineSimCaptureFromContext(body)
	if capture.Preferred != "9053099796" {
		t.Fatalf("preferred=%q want 9053099796", capture.Preferred)
	}
	if !containsString(capture.CTNs, "9095413697") {
		t.Fatalf("ctns=%v want master product", capture.CTNs)
	}
}

func TestCollectBeelineCTNsFromJSON(t *testing.T) {
	body := []byte(`{"data":{"items":[{"phone":{"number":"9053099796"}},{"login":"9095413697"}]}}`)
	numbers := collectBeelineCTNsFromJSON(body)
	if len(numbers) != 2 {
		t.Fatalf("numbers=%v want 2 entries", numbers)
	}
}

func TestMergeBeelineCTNs(t *testing.T) {
	merged := mergeBeelineCTNs([]string{"9053099796"}, []string{"9053099796", "9095413697"})
	if len(merged) != 2 {
		t.Fatalf("merged=%v want 2 unique numbers", merged)
	}
}
