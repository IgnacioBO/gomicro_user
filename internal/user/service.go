package user

import (
	"log"

	"github.com/IgnacioBO/gomicro_domain/domain"
)

//**Capa service o business layer**
//Parecido a la capa endpoint
//Crearemos una interface llamda Service
//En la capa controlador (endpoint) manejamos con struct
//Pero en capa sevicio y capa repositorio SE MANEJRA CON INTERFACE -> porque es mas facil mockearlo o utilizarlo de manera mas generica

// Aqui definiremos lo metodos que las struct deberan tener
type Service interface {
	Create(firstName, lastName, email, phone string) (*domain.User, error) //Metodo que recibira datos de creacion y devolvera un error (y la entidad domain.User)
	GetAll(filtros Filtros, offset, limit int) ([]domain.User, error)      //Le agregamos filtros (con el struct filtro sque creamos)
	Get(id string) (*domain.User, error)
	Delete(id string) error
	Update(id string, firstName, lastName, email, phone *string) error
	Count(Filtros Filtros) (int, error) //Servirá para contar cantidad de registrosy recibe los mismo filtros del getall y devolera int(cantidad de registros) y error
}

// Ahora crearemos un struct PRIVADA (pq desde afuera accederemoa a traves de Servivce)
// Recibira un repository (de la capa repositry)
// Tambien recibira un logger
type service struct {
	log  *log.Logger
	repo Repository
}

// Crea (instanciar) un servicio que sera la interfaz (devovlerá una interface de tupo Service [creado arriba], PERO hara un RETURN especificamente del STRUCT service (con minusculas))
// Recibirá un objeo Repositor y devovlera un service con el repo
// Tambien recibira un logger
func NewService(log *log.Logger, repo Repository) Service {
	return &service{
		log:  log,
		repo: repo,
	}
}

// Ahora crearemos esta stuct que llamaremos Filters o Filtro que servira para filtrar en lso GET
type Filtros struct {
	FirstName string
	LastName  string
}

// Crearemos un metodo Create que será de la struct service (OJO NO CONFUNDIR CON EL INTERFACE)
// Aqui crear un USER usando el repositry (s.repo) y usando un (del domain)
// Devolvera un domain.User (para devolverlo al cliente por api) y un errorr
func (s service) Create(firstName, lastName, email, phone string) (*domain.User, error) {
	s.log.Println("Create user service")
	usuarioNuevo := domain.User{
		FirstName: firstName,
		LastName:  lastName,
		Email:     email,
		Phone:     phone,
	}
	//Le pasamo al repo el domain.User (del domain.go) a la capa repo a la funcion Create (que recibe puntero)
	err := s.repo.Create(&usuarioNuevo)
	//Si hay un error (por ejemplo al insertar, se devuelve el error y la capa endpoitn lo maneja con un status code y todo)
	if err != nil {
		return nil, err
	}
	return &usuarioNuevo, nil
}

func (s service) GetAll(filtros Filtros, offset, limit int) ([]domain.User, error) {
	s.log.Println("GetAll users service")

	allUsers, err := s.repo.GetAll(filtros, offset, limit)
	if err != nil {
		return nil, err
	}
	return allUsers, nil
}

func (s service) Get(id string) (*domain.User, error) {
	s.log.Println("Get by id users service")

	usuario, err := s.repo.Get(id)
	if err != nil {
		return nil, err
	}
	return usuario, nil
}

func (s service) Delete(id string) error {
	s.log.Println("Delete by id users service")

	err := s.repo.Delete(id)
	return err
}

func (s service) Update(id string, firstName, lastName, email, phone *string) error {
	s.log.Println("Update user service")
	err := s.repo.Update(id, firstName, lastName, email, phone)
	return err
}

func (s service) Count(filtros Filtros) (int, error) {
	s.log.Println("Count users service")
	return s.repo.Count(filtros)
}
