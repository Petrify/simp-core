package db

type Service struct {
	id uint
	sources [struct{}]Source

}

type Source interface {
	Id() struct{}
	
}

type 