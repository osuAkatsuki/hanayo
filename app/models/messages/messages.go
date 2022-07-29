package models

type Message interface {
	Type() string
	Content() string
}

type ErrorMessage struct {
	C string
}

func (ErrorMessage) Type() string      { return "error" }
func (m ErrorMessage) Content() string { return m.C }

type NeutralMessage struct {
	C string
}

func (NeutralMessage) Type() string      { return "" }
func (m NeutralMessage) Content() string { return m.C }

type InfoMessage struct {
	C string
}

func (InfoMessage) Type() string      { return "info" }
func (m InfoMessage) Content() string { return m.C }

type SuccessMessage struct {
	C string
}

func (SuccessMessage) Type() string      { return "positive" }
func (m SuccessMessage) Content() string { return m.C }

type WarningMessage struct {
	C string
}

func (WarningMessage) Type() string      { return "warning" }
func (m WarningMessage) Content() string { return m.C }
