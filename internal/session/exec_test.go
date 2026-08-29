package session

import (
	"os"
	"testing"
)

func TestParseExec(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want []string
	}{
		{
			"simple",
			"Hyprland",
			[]string{"Hyprland"},
		},
		{
			"args",
			"sway --some-option",
			[]string{"sway", "--some-option"},
		},
		{
			"field codes stripped",
			"some-program --foo %U",
			[]string{"some-program", "--foo"},
		},
		{
			"escaped percent",
			"foo %% bar",
			[]string{"foo", "%", "bar"},
		},
		{
			"double quotes preserve spaces",
			`"/path with spaces/bin" --opt "value with spaces"`,
			[]string{"/path with spaces/bin", "--opt", "value with spaces"},
		},
		{
			"double quote escapes",
			`"a\"b" "c\\d"`,
			[]string{`a"b`, `c\d`},
		},
		{
			"single quotes literal",
			`'a $b ` + "`c`" + `'`,
			[]string{"a $b `c`"},
		},
		{
			"backslash outside quotes",
			`a\ b c\\d`,
			[]string{"a b", `c\d`},
		},
		{
			"empty quoted arg preserved",
			`foo "" bar`,
			[]string{"foo", "", "bar"},
		},
		{
			"multiple spaces collapse",
			"foo    bar",
			[]string{"foo", "bar"},
		},
		{
			"trailing space",
			"foo bar ",
			[]string{"foo", "bar"},
		},
		{
			"escaped percent at end",
			"foo %",
			[]string{"foo"},
		},
		{
			"real world sway",
			"sway",
			[]string{"sway"},
		},
		{
			"real world xorg",
			"startx -- /usr/bin/X :0",
			[]string{"startx", "--", "/usr/bin/X", ":0"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ParseExec(tc.in)
			if len(got) != len(tc.want) {
				t.Fatalf("len: got %d (%v), want %d (%v)", len(got), got, len(tc.want), tc.want)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Errorf("arg %d: got %q, want %q", i, got[i], tc.want[i])
				}
			}
		})
	}
}

func TestParseDesktopFile(t *testing.T) {
	content := `[Desktop Entry]
Type=Application
Name=Sway
Exec=sway
DesktopNames=sway
X-GNOME-Does-Not-Exist=whatever

[Other Section]
Name=Ignore Me
`
	tmp := t.TempDir() + "/sway.desktop"
	if err := os.WriteFile(tmp, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	sess, err := parseDesktopFile(tmp, Wayland)
	if err != nil {
		t.Fatal(err)
	}
	if sess.Exec != "sway" {
		t.Errorf("Exec: got %q, want %q", sess.Exec, "sway")
	}
	if sess.Name != "Sway" {
		t.Errorf("Name: got %q, want %q", sess.Name, "Sway")
	}
	if sess.DesktopName != "sway" {
		t.Errorf("DesktopName: got %q, want %q", sess.DesktopName, "sway")
	}
	if sess.Type != Wayland {
		t.Errorf("Type: got %q, want %q", sess.Type, Wayland)
	}
}
