package main

import "testing"

func TestBuildMigratedContent(t *testing.T) {
	tests := []struct {
		name     string
		original string
		targets  []targetRow
		want     string
	}{
		{
			name:     "valid target with note, above the body",
			original: "main body",
			targets: []targetRow{
				{TargetReplyID: 99, Note: "note A", TargetFloor: 3, TargetUserID: 42, TargetExists: true},
			},
			want: "[@](kungal-user:42) [#3](kungal-reply:99)\n\nnote A\n\nmain body",
		},
		{
			name:     "dangling target keeps the note, drops the link",
			original: "body",
			targets: []targetRow{
				{TargetReplyID: 0, Note: "orphan note", TargetExists: false},
			},
			want: "orphan note\n\nbody",
		},
		{
			name:     "valid target with blank note → header only",
			original: "",
			targets: []targetRow{
				{TargetReplyID: 5, Note: "  ", TargetFloor: 2, TargetUserID: 7, TargetExists: true},
			},
			want: "[@](kungal-user:7) [#2](kungal-reply:5)",
		},
		{
			name:     "multiple targets stack above the body",
			original: "C",
			targets: []targetRow{
				{TargetReplyID: 1, Note: "A", TargetFloor: 1, TargetUserID: 10, TargetExists: true},
				{TargetReplyID: 2, Note: "B", TargetFloor: 2, TargetUserID: 20, TargetExists: true},
			},
			want: "[@](kungal-user:10) [#1](kungal-reply:1)\n\nA\n\n[@](kungal-user:20) [#2](kungal-reply:2)\n\nB\n\nC",
		},
		{
			name:     "empty everything → empty",
			original: "",
			targets:  []targetRow{{TargetReplyID: 0, Note: "", TargetExists: false}},
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildMigratedContent(tt.original, tt.targets); got != tt.want {
				t.Errorf("buildMigratedContent()\n got: %q\nwant: %q", got, tt.want)
			}
		})
	}
}
