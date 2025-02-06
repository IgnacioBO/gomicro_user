package user

import (
	"fmt"
	"log"
	"strings"

	"github.com/IgnacioBO/gomicro_domain/domain"
	"gorm.io/gorm"
)

//**Capa repositorio o persistencia**
//Se crea similar a la capa de servicio

// Generaremos una interface
type Repository interface {
	Create(user *domain.User) error                                   //Metodo create y recibe un Puntero de un domain.User (Struct creado en el de domain.go, que tiene los campso de BBDD en gorn)
	GetAll(filtros Filtros, offset, limit int) ([]domain.User, error) //Le agregamos que getAll reciba filtros
	Get(id string) (*domain.User, error)
	Delete(id string) error
	Update(id string, firstName *string, lastName *string, email *string, phone *string) error //Campos por separado y como punteros (porque si no lo pongo puntero, si llega un string vacio TENDRA valor y actualizará VACIO)
	Count(Filtros Filtros) (int, error)                                                        //Servirá para contar cantidad de registrosy recibe los mismo filtros del getall y devolera int(cantidad de registros) y error
}

// Ahora una struct que hacer referncia de bbdd de GORN
// Repositorio tendra la bbdd que hemos configurado
// Tambien tendra un logger
type repo struct {
	log *log.Logger
	db  *gorm.DB
}

// Funcion que se encargará de instanciar este Repositry
// Recibirá una BBDD desde el main de gorm y devolvera una interface de Repository (Creada arriba)
// Recibira un logger tambien
func NewRepo(log *log.Logger, db *gorm.DB) Repository {
	return &repo{
		log: log,
		db:  db, //Devuevle un struct repo con la bbdd
	}

}

func (r *repo) Create(user *domain.User) error {
	r.log.Println("repository Create:", user)
	//Aqui zcraeremos el UUID (pq es la capa repository) del usuario usando el package uuid: go get github.com/google/uuid
	//Ese UUID se lo asignaremos al campo ID del user recibido
	//Ahora usaremos HOOKs de GORN, ahora el domain.go se encargara de hacer el uuid SIEMPRE, antes de un CREATE de manera AUTOMATICA con la funcion "BeforeCreate()"
	//user.ID = uuid.New().String()

	//Objeto db tiene el metodo Create (de GORM) y le pasamos la entidad
	result := r.db.Create(user)
	//Si hay error al insertar (x ejemplo nombre muy largo), retornara el error (a la capa servicio)
	//Una manera mas rapida es (por ahora lo omito por enredad) if err := r.db.Create(user).Error; err =! nil {}
	if result.Error != nil {
		r.log.Println(result.Error)
		return result.Error
	}
	r.log.Printf("user created with id: %s, rows affected: %d\n", user.ID, result.RowsAffected)
	return nil
}

func (r *repo) GetAll(filtros Filtros, offset, limit int) ([]domain.User, error) {
	r.log.Println("repository GetAll:")

	var allUsers []domain.User //Variable que almacenará los usuarios obtenidos

	//yo lo hice asi: result := r.db.Find(&alldomain.Users)
	//Desde objeto repo (r) obtenemso bbdd y usamos model para indicar el "modelo" a usar (strct)
	//Order para indicar como queremo devolver (order by) y el Find nos pobla/llkena la estructura con los datos devueltor por la bbdd
	//ORIGINAL SIN FILTOS: result := r.db.Model(&alldomain.Users).Order("created_at desc").Find(&alldomain.Users)
	//AHora se cambiara y le podnremos filtros

	//Primero especificamos el modelo y nos devovlera un gorm.DB* con el modelo listo
	tx := r.db.Model(&allUsers)
	//Luego a esta db con el modelo le aplicaremos filtros
	tx = aplicarFiltros(tx, filtros)
	//AGREGAMOS NUEVO QUE PEMRITE CALCULAR EL OFFSET Y LIMT // offset es a parti de que resultado se muestra, por ejemplo si es 4, se parte del 5* y limit es cantidad desde ese offset
	tx = tx.Limit(limit).Offset(offset)
	//Luego le ponemos un order by y el find para buscar
	result := tx.Order("created_at desc").Find(&allUsers)
	if result.Error != nil {
		r.log.Println(result.Error)
		return nil, result.Error
	}
	r.log.Printf("all users retrieved, rows affected: %d\n", result.RowsAffected)
	return allUsers, nil
}

func (r *repo) Get(id string) (*domain.User, error) {
	r.log.Println("repository Get by id:")

	//Creamos un domain.User y le pasamos el ID a buscar
	usuario := domain.User{ID: id}

	//yo lo hice asi: result := r.db.First(&usuario, "id=?", id)
	//Aqui usuando First se le puede pasar el struct y lo analiza, como pusimos a este usaurio le pusimos ID, buscara por ese ID
	//Ojo usar First y no FIND, porque Find devolvera 0, pero no error
	result := r.db.First(&usuario)
	if result.Error != nil {
		r.log.Println(result.Error)
		return nil, result.Error
	}
	r.log.Printf("user retrieved with id: %s, rows affected: %d\n", id, result.RowsAffected)
	return &usuario, nil
}

func (r *repo) Delete(id string) error {
	r.log.Println("repository Delete by id:")

	//Creamos un domain.User y le pasamos el ID a eliminar
	usuario := domain.User{ID: id}

	//Si esta el campo deleteAt en el domain (domain.User{}), es un SofDelete, si no esta es un delete normal
	//Si tiengo el campo deleteAt, y quiero hacer un delete normal : db.Unscoped().Delete(&order)
	result := r.db.Delete(&usuario)
	if result.Error != nil {
		r.log.Println(result.Error)
		return result.Error
	}
	if result.RowsAffected == 0 {
		r.log.Println("user with id: %s not found, rows affected: %d\n", id, result.RowsAffected)
		return fmt.Errorf("user with id: %s not found", id)
	}
	r.log.Printf("user deleted with id: %s, rows affected: %d\n", id, result.RowsAffected)
	return nil
}

// Recibo String pero como PUNTEROS *, porque asi si podemos distinguir entre vacío (por ejemplo cliente envia phone="") y nil (nil seria que NO envío el campo)
// Si no usamso puntero un string sin valor seria "", en cambio un string puntero sin valor seria nil
func (r *repo) Update(id string, firstName *string, lastName *string, email *string, phone *string) error {
	r.log.Println("repository Update")
	//Usaremos un MAP, porque si usamos el struct, NO ACTUALIZA VALORES CERO (osea "", 0, false)
	//Al usar un map es [string]intareface{}, se usa interface en el valor porque peude ser numerico, string, bool
	valores := make(map[string]interface{})

	if firstName != nil { //Si viene en nulo NO FUE ENVIADO, ya que el puntero no tednria valor. Si el string original viene vacio (por ejemplo "") singifica que si ha sido enviado en el endpoit y por lo tal el puntero NO SERIA NIL (tendria una direccino)
		valores["first_name"] = *firstName //Recordar que al hacer *firstName con asterisco accedemos al valor del puntero *firstName (por ejemplo "Juan"). (si ponemos = firstName devolveria la mmeoria)
	}

	if lastName != nil {
		valores["last_name"] = *lastName
	}

	if email != nil {
		valores["email"] = *email
	}

	if phone != nil {
		valores["phone"] = *phone
	}

	result := r.db.Model(domain.User{}).Where("id = ?", id).Updates(valores)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		r.log.Println("user with id: %s not found, rows affected: %d\n", id, result.RowsAffected)
		return fmt.Errorf("user with id: %s not found", id)
	}
	r.log.Printf("user updated with id: %s, rows affected: %d\n", id, result.RowsAffected)

	return nil
}

// Funcion que servira para filtrar, recibe la base da datos (tx) y el struct de filtros
func aplicarFiltros(tx *gorm.DB, filtros Filtros) *gorm.DB {
	//Si el filtro es distinto de blanco (osea VIENE con filtro), le agregaremos un fultros
	if filtros.FirstName != "" {
		//Primero se hace lowervase para luegos buscar tambein en lowercase en la bbdd
		//Se usea %%%s%% para que termine al final como "%%s%%", porque se usara LIKE y el LIKE con %% permite que sea una especie de "INCLUDE
		//Osea buscar que la apalabra que se busca puede estar al principio, al medio o al final de una palabra
		filtros.FirstName = fmt.Sprintf("%%%s%%", strings.ToLower(filtros.FirstName))
		//El Where filtra el valor que le paso, osea el Where permite AGREGAR un Where a la consulta
		tx = tx.Where("lower(first_name) like ?", filtros.FirstName)
	}

	if filtros.LastName != "" {
		filtros.LastName = fmt.Sprintf("%%%s%%", strings.ToLower(filtros.LastName))
		tx = tx.Where("lower(last_name) like ?", filtros.LastName)
	}
	return tx
}

// Funcion que permitira contar la cantidad de registros devueltos en un get
func (r *repo) Count(filtros Filtros) (int, error) {
	var cantidad int64
	//Creamos un db usando el modelo de user vacio
	tx := r.db.Model(domain.User{})
	//Luego le aplicamos filtros (los where)
	tx = aplicarFiltros(tx, filtros)
	//Ahora le aplicamos COunt a la base da datos que permite consutlar con filtros y devuelev SOLO LA CANTIADA DE RESUTLADOS y se guardara en &cantidad
	//Luego se hará la consulta completa en otro metodo
	//¿Hare doble consulta entonces (pq despues del count debo hacer un select)? SI, pero esto permitira hacer una paginacion, asi preguntar catnidad de resultados primero y luego paginar
	tx = tx.Count(&cantidad)
	if tx.Error != nil {
		return 0, tx.Error
	}

	return int(cantidad), nil
}
