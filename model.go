package go_inapp_parser

type Type string

const Android Type = "android"
const Apple Type = "ios"

//easyjson:json
type Info struct {
	ID        string
	Name      string
	Type      Type
	Category  []string
	Publisher string
	Url       string
	Icon      string
	Rate      float64
}
