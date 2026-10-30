package textfmt

import "testing"

func TestBold(t *testing.T) {
	tests := []struct{ in, want string }{
		{"plain text", "plain text"},
		{"achieving **99% network uptime** across", "achieving <strong>99% network uptime</strong> across"},
		{"**a** and **b**", "<strong>a</strong> and <strong>b</strong>"},
		{"<script>alert(1)</script>", "&lt;script&gt;alert(1)&lt;/script&gt;"},
		{"**<script>x</script>**", "<strong>&lt;script&gt;x&lt;/script&gt;</strong>"},
		{"Cisco switches & routers", "Cisco switches &amp; routers"},
		{"unclosed **bold here", "unclosed **bold here"},
		{"**a** then **unclosed", "<strong>a</strong> then **unclosed"},
		{"", ""},
	}
	for _, tt := range tests {
		if got := Bold(tt.in); got != tt.want {
			t.Errorf("Bold(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
