package main

type message interface {
	Type() string
	Content() string
}

type errorMessage struct {
	C string
}

func (errorMessage) Type() string      { return "error" }
func (m errorMessage) Content() string { return m.C }

type neutralMessage struct {
	C string
}

func (neutralMessage) Type() string      { return "" }
func (m neutralMessage) Content() string { return m.C }

type infoMessage struct {
	C string
}

func (infoMessage) Type() string      { return "info" }
func (m infoMessage) Content() string { return m.C }

type successMessage struct {
	C string
}

func (successMessage) Type() string      { return "positive" }
func (m successMessage) Content() string { return m.C }

type warningMessage struct {
	C string
}

func (warningMessage) Type() string      { return "warning" }
func (m warningMessage) Content() string { return m.C }
