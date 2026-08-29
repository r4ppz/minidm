package session

type Type string

const (
	Wayland Type = "wayland"
	X11     Type = "x11"
)

type Session struct {
	ID          string
	Name        string
	Exec        string
	DesktopName string
	Type        Type
}
