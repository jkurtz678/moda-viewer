package webview

type PlaqueManagerStub struct {
	PlaqueInit bool
}

func (p *PlaqueManagerStub) InitPlaque() {
	p.PlaqueInit = true
}
