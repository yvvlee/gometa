package invalid

import "github.com/yvvlee/gometa"

type Service struct {
	ID int64
}

func (Service) Get(id int64) (string, error) { return "", nil }

var _ = gometa.Register[Service](
	gometa.Field("Missing"), // want `unknown field Service.Missing`
	gometa.Field("ID"),
	gometa.Field("ID"),       // want `duplicate gometa field "ID"`
	gometa.Method("Missing"), // want `unknown method Service.Missing`
	gometa.Method(
		"Get",
		gometa.Param(1),  // want `parameter index 1 out of range for Service.Get`
		gometa.Result(2), // want `result index 2 out of range for Service.Get`
	),
)
