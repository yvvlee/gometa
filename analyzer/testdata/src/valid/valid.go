package valid

import "github.com/yvvlee/gometa"

type Service struct {
	ID int64
}

func (*Service) Get(id int64) (string, error) { return "", nil }

var _ = gometa.Register[Service](
	gometa.Field("ID"),
	gometa.Method("Get", gometa.Param(0), gometa.Result(0), gometa.Result(1)),
)
