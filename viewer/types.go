package viewer

type Col struct {
	Name        string
	Description string
	Length      int
	Var_key     string
	Precision   int
	Units       string
}

type Colgroup struct {
	Name        string
	Description string
	Cols        []Col
}

type View struct {
	Name        string     `yaml:"name"`
	Description string     `yaml:"description"`
	Groups      []Colgroup `yaml:"groups"`
}
