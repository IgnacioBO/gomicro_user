package user

//**Capa endpoint o controlador**

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/IgnacioBO/go_lib_response/response"
	"github.com/IgnacioBO/gomicro_meta/meta"
)

// Struct que tenga todos los endpoints que vayamos a utilizar
// Que teng una fucion que recibe un request y un response
type (
	//Como usaremos gokit cambiaremos el conrtoller para que tenga los datos que necesita gokit (reciviendo context, interface) y devolvendo interface y error
	//OSEA ahora el controller tendra el request YA MAPEADO en el struct que corresponda (por ejempo el struct CreateRequest o UpdateRequest)
	Controller func(ctx context.Context, request interface{}) (interface{}, error)

	Endpoints struct {
		Create        Controller //Esto es lo mismo que decir Create func(w http.ResponseWriter, r *http.Request), pero como TODOS SON tipo Controller (Definido arriba) nos ahorramos ahcerlo
		Get           Controller
		GetAll        Controller
		Update        Controller
		Delete        Controller
		DeleteClassic Controller
	}
	//Definiremos una struct para definir el request del Craete, con los campos que quiero recibir y los tags de json
	CreateRequest struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Email     string `json:"email"`
		Phone     string `json:"phone"`
	}
	//Definiremos una struct para definir el request del UPDATE, con los campos que quiero y SE PODRAN ACTUALIZAR y los tags de json
	//Seran de tipo puntero * para que puedan venir vacios y poder separar entre vacios "" y que no vengan
	UpdateRequest struct {
		ID        string  `json:"-"`
		FirstName *string `json:"first_name"`
		LastName  *string `json:"last_name"`
		Email     *string `json:"email"`
		Phone     *string `json:"phone"`
	}

	//Un struct base del Get que recibe un string que es el id
	GetRequest struct {
		ID string
	}

	//Un struct base del Delete que recibe un string que es el id
	DeleteRequest struct {
		ID string `json:"id"`
	}

	//Este struct tendra los PARAMETROS de la URL para pasarselo
	GetAllRequest struct {
		FirstName string
		LastName  string
		Limit     int
		Page      int
	}

	//Struct para guardar la cant page por defecto y otras conf
	Config struct {
		LimitPageDefault string
	}
)

// Funcion que se encargará de hacer los endopints
// Para eso necesitaremos una struct que se llamara endpoints
// Esta funcion va a DEVOLVER una struct de Endpoints, estos endpoints son los que vamos a poder utuaizlar en unestro dominio (user)
func MakeEndpoints(s Service, config Config) Endpoints {
	return Endpoints{
		Create: makeCreateEndpoint(s),
		Get:    makeGetEndpoint(s),
		Update: makeUpdateEndpoint(s),
		Delete: makeDeleteEndpoint(s),
		GetAll: makeGetAllEndpoint(s, config),
	}
}

// Este devolver un Controller, retora una función de tipo Controller (que definimos arriba) con esta caractesitica
// Es privado porque se llamar solo de este dominio

func makeDeleteEndpoint(s Service) Controller {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		fmt.Println("delete user")

		delStruct := request.(DeleteRequest)
		err := s.Delete(ctx, delStruct.ID)
		if err != nil {
			return nil, response.NotFound(err.Error())
		}

		//Aqui le pasamos Response como struct para ahorrar memoria
		//Con puntero (&Response): Encode accede al struct original a través de la dirección de memoria. Esto evita copiar los datos.
		//Sin puntero (Response): Encode recibe una copia del struct completo, lo que puede ocupar más memoria si el struct es muy grande.
		return response.OK("success", delStruct, nil), nil
	}
}

// request (interface{}) lo pasará el middleware y TENDRA ya los datos del request en un struct listo para usar
func makeCreateEndpoint(s Service) Controller {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		fmt.Println("create user")
		//w.Header().Add("Content-Type", "application/json; charset=utf-8") //Linea miea para que se determine que respondera un json

		//Variable con struct de request (datos usaurio)
		//Aqui hacemo una asersion type assertion (.(CreateRequest)) sobre request
		//Osea asumimos que request (interface{}) es de tupo CreteRequest y se le asignamos a reqStrut
		//Si request NO ES de tipo CreateRqueste se tirara un PANIC
		//requStruct ya tendra todos ls valoers del request, pq se los pasa el middleware desde el parametor "request", asi que podems acceder diretamente
		reqStruct := request.(CreateRequest)

		//Validaciones
		if reqStruct.FirstName == "" {
			//return nil, errors.New("first_name is required")
			return nil, response.BadRequest("first_name is required")

		}
		if reqStruct.LastName == "" {
			return nil, response.BadRequest("last_name is required")
		}
		fmt.Println(reqStruct)
		reqStrucEnJson, _ := json.MarshalIndent(reqStruct, "", " ")
		fmt.Println(string(reqStrucEnJson))

		//Usaremos la s recibida como parametro (de la capa Service y usaremos el metodo CREATE con lo que debe recibir)
		usuarioNuevo, err := s.Create(ctx, reqStruct.FirstName, reqStruct.LastName, reqStruct.Email, reqStruct.Phone)
		if err != nil {
			return nil, response.InternalServerError(err.Error())
		}

		//Aqui retornarmeos la interface (otro middlewre se encargará de "enviar" la response en base al interface)
		return response.Created("success", usuarioNuevo, nil), nil
	}
}

func makeUpdateEndpoint(s Service) Controller {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		fmt.Println("update user")

		//Variable con struct de request (datos de atualizacion)
		reqStruct := request.(UpdateRequest)

		//Permite NO ENVIAR ESTOS CAMPOS, PERO NO SE PERMITE ENVIARLSO VACIOS
		if reqStruct.FirstName != nil && *reqStruct.FirstName == "" {
			return nil, response.BadRequest("first_name can't be empty")
		}

		if reqStruct.LastName != nil && *reqStruct.LastName == "" {
			return nil, response.BadRequest("last_name can't be empty")
		}

		id := reqStruct.ID

		err := s.Update(ctx, id, reqStruct.FirstName, reqStruct.LastName, reqStruct.Email, reqStruct.Phone)
		if err != nil {
			return nil, response.NotFound(err.Error())

		}

		return response.OK("success", map[string]string{"id": id}, nil), nil

	}
}

func makeGetEndpoint(s Service) Controller {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		fmt.Println("get user")

		getReq := request.(GetRequest) //Obtendremos un GetRequest (qie tiene el id)

		usuario, err := s.Get(ctx, getReq.ID)
		if err != nil {
			if usuario == nil { //Si usuario es vacio da 404
				return nil, response.NotFound(err.Error() + ". user with id " + getReq.ID + " doesn't exist")
				//json.NewEncoder(w).Encode(&Response{Status: 404, Err: err.Error() + ". user with id " + id + " doesn't exist"}) //Aqui devolvemo el posible erro
			} else {
				return nil, response.BadRequest(err.Error())
			}
		}
		return response.OK("success", usuario, nil), nil

	}
}

func makeGetAllEndpoint(s Service, config Config) Controller {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		fmt.Println("getall user")

		getAllParametros := request.(GetAllRequest)
		//Luego con podemos acceder a los parametos y guardarlos en el struct Filtro (creado en service.go)
		filtros := Filtros{
			FirstName: getAllParametros.FirstName,
			LastName:  getAllParametros.LastName,
		}

		//Ahora llamaremos al Count del service que creamos (antes de hacer la consulta completa)
		cantidad, err := s.Count(ctx, filtros)
		if err != nil {
			return nil, response.InternalServerError(err.Error())
		}
		//Luego crearemos un meta y le agregaremos la cantidad que consultamos, luego el meta lo ageregaremos a la respuesta
		meta, err := meta.New(getAllParametros.Page, getAllParametros.Limit, cantidad, config.LimitPageDefault)
		if err != nil {
			return nil, response.InternalServerError(err.Error())
		}

		allUsers, err := s.GetAll(ctx, filtros, meta.Offset(), meta.Limit()) //GetAll recibe el offset (desde q resultado mostrar) y el limit (cuantos desde el offset)
		if err != nil {
			return nil, response.InternalServerError(err.Error())
		}

		return response.OK("success", allUsers, meta), nil
	}
}
