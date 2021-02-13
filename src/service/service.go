package service

type Service interface {
	Start() error
	Stop()
	Running() bool
	Status() (error, Status)
	
}
