package contratos

type Evento struct {
	PK        string `json:"Evento"`
	Capacidad int
	Categoria string
	Estado    string
}

type REvento struct {
	SK        string `json:"Evento"`
	Capacidad int
	Categoria string
	Estado    string
}

type Reserva struct {
	SK       string `json:"Id"`
	Evento   string
	Estado   string
	Email    string
	Cantidad int
}

