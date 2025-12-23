package service

import "errors"

var (
	// В аргументе Conf нет указателя.
	NilPtrArgumentConf = errors.New("в аргументе conf нет указателя")
)
