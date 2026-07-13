package headroom

import "testing"

func TestRouteContent(t *testing.T) {
	tests := []struct {
		name string
		text string
		want contentKind
	}{
		{name: "json", text: `{"items":[1,2,3]}`, want: contentKindJSON},
		{name: "code", text: "```go\nfunc main() {}\n```", want: contentKindCode},
		{name: "log", text: "INFO start\nERROR failed\nWARN retry", want: contentKindLog},
		{name: "text", text: "This is a normal paragraph with details.", want: contentKindText},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := routeContent(tt.text); got != tt.want {
				t.Fatalf("routeContent() = %v, want %v", got, tt.want)
			}
		})
	}
}
