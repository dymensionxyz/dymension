package types

import "testing"

func TestEIBCMemo_ValidateBasic_Escalation(t *testing.T) {
	tests := []struct {
		name    string
		memo    EIBCMemo
		wantErr bool
	}{
		{"static (no escalation)", EIBCMemo{Fee: "100"}, false},
		{"valid escalation", EIBCMemo{Fee: "100", FeeMax: "200", FeeEscalationBlocks: 10}, false},
		{"only fee_max set", EIBCMemo{Fee: "100", FeeMax: "200"}, true},
		{"only blocks set", EIBCMemo{Fee: "100", FeeEscalationBlocks: 10}, true},
		{"fee_max < fee", EIBCMemo{Fee: "100", FeeMax: "50", FeeEscalationBlocks: 10}, true},
		{"fee_max == fee ok", EIBCMemo{Fee: "100", FeeMax: "100", FeeEscalationBlocks: 10}, false},
		{"unparseable fee_max", EIBCMemo{Fee: "100", FeeMax: "abc", FeeEscalationBlocks: 10}, true},
		{"negative fee_max", EIBCMemo{Fee: "100", FeeMax: "-5", FeeEscalationBlocks: 10}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.memo.ValidateBasic()
			if (err != nil) != tt.wantErr {
				t.Fatalf("ValidateBasic() err = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
